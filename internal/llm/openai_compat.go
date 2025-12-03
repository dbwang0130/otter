package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

type Option func(*openAIModel)

func WithAPIKey(apiKey string) Option {
	return func(cfg *openAIModel) {
		cfg.apiKey = apiKey
	}
}

func WithBaseURL(baseUrl string) Option {
	return func(cfg *openAIModel) {
		cfg.baseUrl = baseUrl
	}
}

func WithModelName(modelName string) Option {
	return func(cfg *openAIModel) {
		cfg.modelName = modelName
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(cfg *openAIModel) {
		cfg.httpClient = httpClient
	}
}

type openAIModel struct {
	apiKey     string
	baseUrl    string
	modelName  string
	httpClient *http.Client
}

func NewOpenAICompatModel(opts ...Option) (model.LLM, error) {
	model := &openAIModel{
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(model)
	}

	if model.apiKey == "" {
		return nil, fmt.Errorf("API key is required. Please set the API key in the configuration")
	}

	if model.baseUrl == "" {
		model.baseUrl = "https://api.openai.com/v1"
	}

	if model.httpClient == nil {
		model.httpClient = http.DefaultClient
	}

	return model, nil
}

func (m *openAIModel) Name() string {
	return m.modelName
}

func (m *openAIModel) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	m.maybeAppendUserContent(req)

	openaiReq, err := m.convertRequest(req)
	if err != nil {
		return func(yield func(*model.LLMResponse, error) bool) {
			yield(nil, fmt.Errorf("failed to convert request: %w", err))
		}
	}

	if stream {
		return m.generateStream(ctx, openaiReq)
	}

	return m.generate(ctx, openaiReq)
}

type openAIRequest struct {
	Model          string                `json:"model"`
	Messages       []openAIMessage       `json:"messages"`
	Tools          []openAITool          `json:"tools,omitempty"`
	Temperature    *float64              `json:"temperature,omitempty"`
	MaxTokens      *int                  `json:"max_tokens,omitempty"`
	TopP           *float64              `json:"top_p,omitempty"`
	N              *int                  `json:"n,omitempty"`
	Stop           []string              `json:"stop,omitempty"`
	Stream         bool                  `json:"stream,omitempty"`
	ResponseFormat *openAIResponseFormat `json:"response_format,omitempty"`
}

type openAIResponseFormat struct {
	Type string `json:"type"`
}

type openAIMessage struct {
	Role             string           `json:"role"`
	Content          any              `json:"content,omitempty"`
	ToolCalls        []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID       string           `json:"tool_call_id,omitempty"`
	ReasoningContent any              `json:"reasoning_content,omitempty"`
}

type openAIToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"` // "function"
	Function openAIFunctionCall `json:"function"`
}

type openAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAITool struct {
	Type     string         `json:"type"` // "function"
	Function openAIFunction `json:"function"`
}

type openAIFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type openAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int            `json:"created"`
	Model   string         `json:"model"`
	Choices []openAIChoice `json:"choices"`
	Usage   *openAIUsage   `json:"usage,omitempty"`
}

type openAIChoice struct {
	Index        int            `json:"index"`
	Message      *openAIMessage `json:"message,omitempty"`
	Delta        *openAIMessage `json:"delta,omitempty"`
	FinishReason string         `json:"finish_reason,omitempty"`
}

type openAIUsage struct {
	PromptTokens       int                  `json:"prompt_tokens"`
	CompletionTokens   int                  `json:"completion_tokens"`
	TotalTokens        int                  `json:"total_tokens"`
	PromptTokensDetail *promptTokensDetails `json:"prompt_tokens_details,omitempty"`
}

type promptTokensDetails struct {
	CachedTokens int `json:"cached_tokens,omitempty"`
}

func (m *openAIModel) convertRequest(req *model.LLMRequest) (*openAIRequest, error) {
	openaiReq := &openAIRequest{
		Model:    m.modelName,
		Messages: make([]openAIMessage, 0),
	}

	if req.Config != nil && req.Config.SystemInstruction != nil {
		sysContent := extractTextFromContent(req.Config.SystemInstruction)
		if sysContent != "" {
			openaiReq.Messages = append(openaiReq.Messages, openAIMessage{
				Role:    "system",
				Content: sysContent,
			})
		}
	}

	for _, content := range req.Contents {
		msgs, err := m.convertContent(content)
		if err != nil {
			return nil, fmt.Errorf("failed to convert content: %w", err)
		}
		openaiReq.Messages = append(openaiReq.Messages, msgs...)
	}

	if req.Config != nil && len(req.Config.Tools) > 0 {
		for _, tool := range req.Config.Tools {
			if tool.FunctionDeclarations != nil {
				for _, fn := range tool.FunctionDeclarations {
					openaiReq.Tools = append(openaiReq.Tools, convertFunctionDeclaration(fn))
				}
			}
		}
	}

	if req.Config != nil {
		if req.Config.Temperature != nil {
			temp := float64(*req.Config.Temperature)
			openaiReq.Temperature = &temp
		}

		if req.Config.MaxOutputTokens > 0 {
			maxTokens := int(req.Config.MaxOutputTokens)
			openaiReq.MaxTokens = &maxTokens
		}

		if req.Config.TopP != nil {
			topP := float64(*req.Config.TopP)
			openaiReq.TopP = &topP
		}

		if len(req.Config.StopSequences) > 0 {
			openaiReq.Stop = req.Config.StopSequences
		}

		if req.Config.ResponseMIMEType == "application/json" {
			openaiReq.ResponseFormat = &openAIResponseFormat{
				Type: "json_object",
			}
		}
	}
	return openaiReq, nil
}

func (m *openAIModel) convertContent(content *genai.Content) ([]openAIMessage, error) {
	if content == nil || len(content.Parts) == 0 {
		return nil, nil
	}

	role := content.Role
	if role == "model" {
		role = "assistant"
	}

	var toolMessages []openAIMessage
	for _, part := range content.Parts {
		if part.FunctionResponse != nil {
			responseJSON, err := json.Marshal(part.FunctionResponse.Response)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal function response: %w", err)
			}

			toolCallID := part.FunctionResponse.ID
			if toolCallID == "" {
				toolCallID = "call_" + uuid.New().String()[:8]
			}
			toolMessages = append(toolMessages, openAIMessage{
				Role:       "tool",
				Content:    string(responseJSON),
				ToolCallID: toolCallID,
			})
		}
	}

	if len(toolMessages) > 0 {
		return toolMessages, nil
	}

	var textParts []string
	var contentArray []map[string]any
	var toolCalls []openAIToolCall

	for _, part := range content.Parts {
		if part.Text != "" {
			textParts = append(textParts, part.Text)
		} else if part.InlineData != nil && len(part.InlineData.Data) > 0 {
			mimeType := part.InlineData.MIMEType
			base64Data := base64.StdEncoding.EncodeToString(part.InlineData.Data)
			dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)

			if strings.HasPrefix(mimeType, "image/") {
				contentArray = append(contentArray, map[string]any{
					"type": "image_url",
					"image_url": map[string]string{
						"url": dataURI,
					},
				})
			} else if strings.HasPrefix(mimeType, "audio/") {
				contentArray = append(contentArray, map[string]any{
					"type": "audio_url",
					"audio_url": map[string]string{
						"url": dataURI,
					},
				})
			} else if strings.HasPrefix(mimeType, "video/") {
				contentArray = append(contentArray, map[string]any{
					"type": "video_url",
					"video_url": map[string]string{
						"url": dataURI,
					},
				})
			} else if mimeType == "application/pdf" || mimeType == "application/json" {
				contentArray = append(contentArray, map[string]any{
					"type": "file",
					"file": map[string]string{
						"flile_data": dataURI,
					},
				})
			} else if strings.HasPrefix(mimeType, "text/") {
				textParts = append(textParts, string(part.InlineData.Data))
			}
		} else if part.FileData != nil && part.FileData.FileURI != "" {
			contentArray = append(contentArray, map[string]any{
				"type": "file",
				"file": map[string]any{
					"file_id": part.FileData.FileURI,
				},
			})
		} else if part.FunctionCall != nil {
			argsJSON, err := json.Marshal(part.FunctionCall.Args)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal function call arguments: %w", err)
			}

			callID := part.FunctionCall.ID
			if callID == "" {
				callID = "call_" + uuid.New().String()[:8]
			}
			toolCalls = append(toolCalls, openAIToolCall{
				ID:   callID,
				Type: "function",
				Function: openAIFunctionCall{
					Name:      part.FunctionCall.Name,
					Arguments: string(argsJSON),
				},
			})
		}
	}

	msg := openAIMessage{Role: role}
	if len(toolCalls) > 0 {
		msg.ToolCalls = toolCalls
		if len(textParts) > 0 {
			msg.Content = strings.Join(textParts, "\n")
		}
	} else if len(contentArray) > 0 {
		textMaps := make([]map[string]any, len(textParts))
		for i, text := range textParts {
			textMaps[i] = map[string]any{
				"type": "text",
				"text": text,
			}
		}
		msg.Content = append(textMaps, contentArray...)
	} else if len(textParts) > 0 {
		msg.Content = strings.Join(textParts, "\n")
	}

	return []openAIMessage{msg}, nil
}

func extractTextFromContent(content *genai.Content) string {
	if content == nil {
		return ""
	}

	var texts []string
	for _, part := range content.Parts {
		if part.Text != "" {
			texts = append(texts, part.Text)
		}
	}

	return strings.Join(texts, "\n")
}

func convertFunctionDeclaration(fn *genai.FunctionDeclaration) openAITool {
	params := convertFunctionParameters(fn)

	return openAITool{
		Type: "function",
		Function: openAIFunction{
			Name:        fn.Name,
			Description: fn.Description,
			Parameters:  params,
		},
	}
}

func convertFunctionParameters(fn *genai.FunctionDeclaration) map[string]any {
	if fn.ParametersJsonSchema != nil {
		if params := tryConvertJsonSchema(fn.ParametersJsonSchema); params != nil {
			return params
		}
	}

	if fn.Parameters != nil {
		return convertLegacyParameters(fn.Parameters)
	}

	return make(map[string]any)
}

func tryConvertJsonSchema(schema any) map[string]any {
	if params, ok := schema.(map[string]any); ok {
		return params
	}

	jsonBytes, err := json.Marshal(schema)
	if err != nil {
		return nil
	}

	var params map[string]any
	if err := json.Unmarshal(jsonBytes, &params); err != nil {
		return nil
	}

	return params
}

func convertLegacyParameters(schema *genai.Schema) map[string]any {
	params := map[string]any{
		"type": "object",
	}

	if schema.Properties != nil {
		props := make(map[string]any)
		for k, v := range schema.Properties {
			props[k] = schemaToMap(v)
		}
		params["properties"] = props
	}

	if len(schema.Required) > 0 {
		params["required"] = schema.Required
	}

	return params
}

func schemaToMap(schema *genai.Schema) any {
	result := make(map[string]any)

	if schema.Type != genai.TypeUnspecified {
		result["type"] = strings.ToLower(string(schema.Type))
	}

	if schema.Description != "" {
		result["description"] = schema.Description
	}

	if schema.Items != nil {
		result["items"] = schemaToMap(schema.Items)
	}

	if schema.Properties != nil {
		props := make(map[string]any)
		for k, v := range schema.Properties {
			props[k] = schemaToMap(v)
		}
		result["properties"] = props
	}

	if schema.Enum != nil {
		result["enum"] = schema.Enum
	}

	return result
}

func (m *openAIModel) generate(ctx context.Context, openaiReq *openAIRequest) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		resp, err := m.doRequest(ctx, openaiReq)
		if err != nil {
			yield(nil, err)
			return
		}

		llmResp, err := m.convertResponse(resp)
		if err != nil {
			yield(nil, err)
			return
		}

		yield(llmResp, nil)
	}
}

func (m *openAIModel) generateStream(ctx context.Context, openaiReq *openAIRequest) iter.Seq2[*model.LLMResponse, error] {
	openaiReq.Stream = true

	return func(yield func(*model.LLMResponse, error) bool) {
		httpResp, err := m.sendRequest(ctx, openaiReq)
		if err != nil {
			yield(nil, err)
			return
		}

		defer httpResp.Body.Close()

		scanner := bufio.NewScanner(httpResp.Body)
		var textbuffer strings.Builder
		var toolCalls []openAIToolCall
		var usage *openAIUsage

		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var chunk openAIResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) == 0 {
				continue
			}

			choice := chunk.Choices[0]
			delta := choice.Delta
			if delta == nil {
				continue
			}

			if delta.Content != nil {
				if text, ok := delta.Content.(string); ok && text != "" {
					textbuffer.WriteString(text)
					llmResp := &model.LLMResponse{
						Content: &genai.Content{
							Role: "model",
							Parts: []*genai.Part{
								{Text: text},
							},
						},
						Partial: true,
					}
					if !yield(llmResp, nil) {
						return
					}
				}
			}

			if len(delta.ToolCalls) > 0 {
				for idx, tc := range delta.ToolCalls {
					for len(toolCalls) <= idx {
						toolCalls = append(toolCalls, openAIToolCall{})
					}
					if tc.ID != "" {
						toolCalls[idx].ID = tc.ID
					}
					if tc.Type != "" {
						toolCalls[idx].Type = tc.Type
					}
					if tc.Function.Name != "" {
						toolCalls[idx].Function.Name = tc.Function.Name
					}
					toolCalls[idx].Function.Arguments += tc.Function.Arguments
				}
			}

			if chunk.Usage != nil {
				usage = chunk.Usage
			}

			if chunk.Usage != nil {
				usage = chunk.Usage
			}

			if choice.FinishReason != "" {
				finalResp := m.buildFinalResponse(textbuffer.String(), toolCalls, usage, choice.FinishReason)
				yield(finalResp, nil)
				return
			}
		}

		if err := scanner.Err(); err != nil {
			yield(nil, fmt.Errorf("stream error: %w", err))
			return
		}

		if textbuffer.Len() > 0 || len(toolCalls) > 0 {
			finalResp := m.buildFinalResponse(textbuffer.String(), toolCalls, usage, "stop")
			yield(finalResp, nil)
		}
	}
}

func (m *openAIModel) sendRequest(ctx context.Context, openaiReq *openAIRequest) (*http.Response, error) {
	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	baseURL := strings.TrimSuffix(m.baseUrl, "/")
	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		slog.Error("failed to create HTTP request: ", "error", err)
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	httpResp, err := m.httpClient.Do(req)
	if err != nil {
		slog.Error("failed to send request: ", "error", err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()

		slog.Error("API error: ", "status", httpResp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("API error: %d - %s", httpResp.StatusCode, string(body))
	}

	return httpResp, nil
}

func (m *openAIModel) doRequest(ctx context.Context, openaiReq *openAIRequest) (*openAIResponse, error) {
	httpResp, err := m.sendRequest(ctx, openaiReq)
	if err != nil {
		return nil, err
	}

	defer httpResp.Body.Close()

	decoder := json.NewDecoder(httpResp.Body)
	var openaiResp openAIResponse
	if err := decoder.Decode(&openaiResp); err != nil {
		slog.Error("failed to decode response: ", "error", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &openaiResp, nil
}

func (m *openAIModel) convertResponse(openaiResp *openAIResponse) (*model.LLMResponse, error) {
	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := openaiResp.Choices[0]
	msg := choice.Message
	if msg == nil {
		slog.Error("no message in response")
		return nil, fmt.Errorf("no message in response")
	}

	var parts []*genai.Part
	if reasoningParts := extractReasoningParts(msg.ReasoningContent); len(reasoningParts) > 0 {
		parts = append(parts, reasoningParts...)
	}

	toolCalls := msg.ToolCalls
	textContent := ""

	if msg.Content != nil {
		if text, ok := msg.Content.(string); ok {
			textContent = text
		}
	}

	if len(toolCalls) == 0 && textContent != "" {
		parsedCalls, remainder := parseToolCallsFromText(textContent)
		if len(parsedCalls) > 0 {
			toolCalls = parsedCalls
			textContent = remainder
		}
	}

	if textContent != "" {
		parts = append(parts, genai.NewPartFromText(textContent))
	}

	for _, tc := range toolCalls {
		var args map[string]any
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			slog.Error("failed to unmarshal function arguments: ", "error", err)
			return nil, fmt.Errorf("failed to unmarshal function arguments: %w", err)
		}

		part := genai.NewPartFromFunctionCall(tc.Function.Name, args)
		part.FunctionCall.ID = tc.ID
		parts = append(parts, part)
	}

	llmResp := &model.LLMResponse{
		Content: &genai.Content{
			Role:  "model",
			Parts: parts,
		},
	}

	llmResp.UsageMetadata = buildUsageMetadata(openaiResp.Usage)
	llmResp.FinishReason = mapFinishReason(choice.FinishReason)
	return llmResp, nil
}

func (m *openAIModel) buildFinalResponse(text string, toolCalls []openAIToolCall, usage *openAIUsage, finishReason string) *model.LLMResponse {
	var parts []*genai.Part

	if text != "" {
		parts = append(parts, genai.NewPartFromText(text))
	}

	for _, tc := range toolCalls {
		var args map[string]any
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err == nil {
			slog.Error("failed to unmarshal function arguments: ", "error", err)
			continue
		}

		part := genai.NewPartFromFunctionCall(tc.Function.Name, args)
		part.FunctionCall.ID = tc.ID
		parts = append(parts, part)
	}

	llmResp := &model.LLMResponse{
		Content: &genai.Content{
			Role:  "model",
			Parts: parts,
		},
		FinishReason:  mapFinishReason(finishReason),
		UsageMetadata: buildUsageMetadata(usage),
	}

	return llmResp
}

func buildUsageMetadata(usage *openAIUsage) *genai.GenerateContentResponseUsageMetadata {
	if usage == nil {
		return nil
	}

	metadata := &genai.GenerateContentResponseUsageMetadata{
		PromptTokenCount:     int32(usage.PromptTokens),
		CandidatesTokenCount: int32(usage.CompletionTokens),
		TotalTokenCount:      int32(usage.TotalTokens),
	}

	if usage.PromptTokensDetail != nil {
		metadata.CachedContentTokenCount = int32(usage.PromptTokensDetail.CachedTokens)
	}
	return metadata
}

func extractReasoningParts(reasoningContent any) []*genai.Part {
	if reasoningContent == nil {
		return nil
	}

	var parts []*genai.Part
	extractTexts(reasoningContent, &parts)
	return parts
}

func extractTexts(content any, parts *[]*genai.Part) {
	if content == nil {
		return
	}

	switch v := content.(type) {
	case string:
		if v != "" {
			*parts = append(*parts, &genai.Part{Text: v, Thought: true})
		}
	case []any:
		for _, item := range v {
			extractTexts(item, parts)
		}
	case map[string]any:
		for _, key := range []string{"text", "content", "reasoning", "reasoning_content"} {
			if text, ok := v[key].(string); ok && text != "" {
				*parts = append(*parts, &genai.Part{Text: text, Thought: true})
			}
		}
	}
}

func parseToolCallsFromText(text string) ([]openAIToolCall, string) {
	if text == "" {
		return nil, ""
	}

	var toolCalls []openAIToolCall
	var remainder strings.Builder
	cursor := 0

	for cursor < len(text) {
		braceIndex := strings.Index(text[cursor:], "{")
		if braceIndex == -1 {
			remainder.WriteString(text[cursor:])
			break
		}

		braceIndex += cursor

		remainder.WriteString(text[cursor:braceIndex])

		var candidate map[string]any
		decoder := json.NewDecoder(strings.NewReader(text[braceIndex:]))
		if err := decoder.Decode(&candidate); err == nil {
			remainder.WriteString(text[braceIndex : braceIndex+1])
			cursor = braceIndex + 1
			continue
		}

		endPos := braceIndex + int(decoder.InputOffset())

		name, hasName := candidate["name"].(string)
		args, hasArgs := candidate["arguments"]
		if hasName && hasArgs {
			argsStr := ""
			switch a := args.(type) {
			case string:
				argsStr = a
			default:
				if jsonBytes, err := json.Marshal(a); err == nil {
					argsStr = string(jsonBytes)
				}
			}

			callID := "call_" + uuid.New().String()[:8]
			if id, ok := candidate["id"].(string); ok && id != "" {
				callID = id
			}

			toolCalls = append(toolCalls, openAIToolCall{
				ID:   callID,
				Type: "function",
				Function: openAIFunctionCall{
					Name:      name,
					Arguments: argsStr,
				},
			})
		} else {
			remainder.WriteString(text[braceIndex:endPos])
		}

		cursor = endPos
	}

	return toolCalls, strings.TrimSpace(remainder.String())
}

func mapFinishReason(reason string) genai.FinishReason {
	switch reason {
	case "stop":
		return genai.FinishReasonStop
	case "length":
		return genai.FinishReasonMaxTokens
	case "tool_calls":
		return genai.FinishReasonStop
	case "content_filter":
		return genai.FinishReasonSafety
	default:
		return genai.FinishReasonOther
	}
}

func (m *openAIModel) maybeAppendUserContent(req *model.LLMRequest) {
	if len(req.Contents) == 0 {
		req.Contents = append(req.Contents, genai.NewContentFromText("Handle the requests as specified in the System Instruction.", "user"))
		return
	}

	if last := req.Contents[len(req.Contents)-1]; last.Role == "user" {
		req.Contents = append(req.Contents, genai.NewContentFromText("Continue processing previous requests as instructed. Exit or provide a summary if no more outputs are needed.", "user"))
	}
}
