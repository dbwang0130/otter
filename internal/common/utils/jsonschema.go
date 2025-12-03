package utils

import (
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
)

const (
	timeFormatDescription = ". 支持格式：RFC3339 (2006-01-02T15:04:05Z07:00)、ISO 8601 (2006-01-02T15:04:05)、日期时间 (2006-01-02 15:04:05)、日期 (2006-01-02)"
)

var (
	timeType = reflect.TypeOf(time.Time{})
)

// isTimeType 检查类型是否为 time.Time（包括指针类型）
func isTimeType(t reflect.Type) bool {
	if t == timeType {
		return true
	}
	if t.Kind() == reflect.Ptr && t.Elem() == timeType {
		return true
	}
	return false
}

// getElemType 获取类型的元素类型（处理指针）
func getElemType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}

// schemaFromType 根据类型生成 schema（核心类型转换函数）
func schemaFromType(t reflect.Type) *jsonschema.Schema {
	// 处理指针类型：递归处理元素类型
	if t.Kind() == reflect.Ptr {
		return schemaFromType(t.Elem())
	}

	// 处理 time.Time 类型
	if t == timeType {
		return &jsonschema.Schema{
			Type:   "string",
			Format: "date-time",
		}
	}

	// 处理结构体类型：递归调用 SchemaFromStruct
	if t.Kind() == reflect.Struct {
		return SchemaFromStruct(reflect.New(t).Elem().Interface())
	}

	// 处理基本类型
	schema := &jsonschema.Schema{}
	switch t.Kind() {
	case reflect.String:
		schema.Type = "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema.Type = "integer"
	case reflect.Float32, reflect.Float64:
		schema.Type = "number"
	case reflect.Bool:
		schema.Type = "boolean"
	case reflect.Slice, reflect.Array:
		schema.Type = "array"
		elemType := getElemType(t.Elem())
		if elemType.Kind() == reflect.Struct {
			schema.Items = schemaFromType(elemType)
		}
	case reflect.Map:
		schema.Type = "object"
		valueType := getElemType(t.Elem())
		schema.AdditionalProperties = schemaFromType(valueType)
	default:
		schema.Type = "string"
	}

	return schema
}

// parseJSONTag 解析 JSON 标签，返回字段名和是否为可选
func parseJSONTag(tag string, fieldName string) (jsonName string, isOptional bool) {
	if tag == "" || tag == "-" {
		return "", true
	}

	parts := strings.Split(tag, ",")
	jsonName = parts[0]
	if jsonName == "" {
		jsonName = strings.ToLower(fieldName)
	}

	isOptional = strings.Contains(tag, "omitempty")
	return jsonName, isOptional
}

// setFieldDescription 设置字段描述信息
func setFieldDescription(schema *jsonschema.Schema, fieldName string, fieldType reflect.Type) {
	if isTimeType(fieldType) {
		schema.Description = fieldName + timeFormatDescription
	} else if schema.Description == "" {
		schema.Description = fieldName
	}
}

// SchemaFromStruct 从结构体生成 JSON Schema
// 自动从结构体的字段、JSON 标签和类型信息生成对应的 JSON Schema
func SchemaFromStruct(v interface{}) *jsonschema.Schema {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	schema := &jsonschema.Schema{
		Type:       "object",
		Properties: make(map[string]*jsonschema.Schema),
	}

	var required []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonName, isOptional := parseJSONTag(field.Tag.Get("json"), field.Name)
		if jsonName == "" {
			continue
		}

		// 生成字段 schema
		fieldSchema := schemaFromType(field.Type)
		setFieldDescription(fieldSchema, field.Name, field.Type)

		// 记录必需字段
		if !isOptional {
			required = append(required, jsonName)
		}

		schema.Properties[jsonName] = fieldSchema
	}

	if len(required) > 0 {
		schema.Required = required
	}

	slog.Debug("SchemaFromStruct", "schema", schema)

	return schema
}
