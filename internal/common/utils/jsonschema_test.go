package utils

import (
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSchemaFromStruct_BasicTypes 测试基本类型
func TestSchemaFromStruct_BasicTypes(t *testing.T) {
	type TestStruct struct {
		StringField string  `json:"string_field"`
		IntField    int     `json:"int_field"`
		FloatField  float64 `json:"float_field"`
		BoolField   bool    `json:"bool_field"`
		StringPtr   *string `json:"string_ptr,omitempty"`
		IntPtr      *int    `json:"int_ptr,omitempty"`
	}

	schema := SchemaFromStruct(TestStruct{})
	log.Println(json.Marshal(schema))

	assert.NotNil(t, schema)
	assert.Equal(t, "object", schema.Type)
	assert.NotNil(t, schema.Properties)

	// 检查必需字段
	assert.Contains(t, schema.Required, "string_field")
	assert.Contains(t, schema.Required, "int_field")
	assert.Contains(t, schema.Required, "float_field")
	assert.Contains(t, schema.Required, "bool_field")
	assert.NotContains(t, schema.Required, "string_ptr")
	assert.NotContains(t, schema.Required, "int_ptr")

	// 检查字段类型
	assert.Equal(t, "string", schema.Properties["string_field"].Type)
	assert.Equal(t, "integer", schema.Properties["int_field"].Type)
	assert.Equal(t, "number", schema.Properties["float_field"].Type)
	assert.Equal(t, "boolean", schema.Properties["bool_field"].Type)
	assert.Equal(t, "string", schema.Properties["string_ptr"].Type)
	assert.Equal(t, "integer", schema.Properties["int_ptr"].Type)
}

// TestSchemaFromStruct_TimeType 测试 time.Time 类型
func TestSchemaFromStruct_TimeType(t *testing.T) {
	type TestStruct struct {
		StartTime time.Time  `json:"start_time"`
		EndTime   *time.Time `json:"end_time,omitempty"`
	}

	schema := SchemaFromStruct(TestStruct{})

	assert.NotNil(t, schema)
	assert.Equal(t, "string", schema.Properties["start_time"].Type)
	assert.Equal(t, "date-time", schema.Properties["start_time"].Format)
	assert.Equal(t, "string", schema.Properties["end_time"].Type)
	assert.Equal(t, "date-time", schema.Properties["end_time"].Format)
	assert.Contains(t, schema.Required, "start_time")
	assert.NotContains(t, schema.Required, "end_time")
}

// TestSchemaFromStruct_RequiredAndOptional 测试必需和可选字段
func TestSchemaFromStruct_RequiredAndOptional(t *testing.T) {
	type TestStruct struct {
		RequiredField  string `json:"required_field"`
		OptionalField  string `json:"optional_field,omitempty"`
		RequiredInt    int    `json:"required_int"`
		OptionalInt    *int   `json:"optional_int,omitempty"`
		IgnoredField   string `json:"-"`
		NoJsonTagField string // 没有 JSON 标签的字段应该被忽略
	}

	schema := SchemaFromStruct(TestStruct{})

	assert.NotNil(t, schema)
	assert.Contains(t, schema.Required, "required_field")
	assert.Contains(t, schema.Required, "required_int")
	assert.NotContains(t, schema.Required, "optional_field")
	assert.NotContains(t, schema.Required, "optional_int")
	assert.NotContains(t, schema.Properties, "ignored_field")
	assert.NotContains(t, schema.Properties, "NoJsonTagField")
}

// TestSchemaFromStruct_EmptyStruct 测试空结构体
func TestSchemaFromStruct_EmptyStruct(t *testing.T) {
	type EmptyStruct struct{}

	schema := SchemaFromStruct(EmptyStruct{})

	assert.NotNil(t, schema)
	assert.Equal(t, "object", schema.Type)
	assert.Empty(t, schema.Properties)
	assert.Empty(t, schema.Required)
}

// TestSchemaFromStruct_PointerToStruct 测试指针类型结构体
func TestSchemaFromStruct_PointerToStruct(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	schema := SchemaFromStruct(&TestStruct{})

	assert.NotNil(t, schema)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "name")
	assert.Contains(t, schema.Properties, "age")
}

// TestSchemaFromStruct_ArrayAndSlice 测试数组和切片类型
func TestSchemaFromStruct_ArrayAndSlice(t *testing.T) {
	type TestStruct struct {
		StringSlice []string `json:"string_slice,omitempty"`
		IntArray    [5]int   `json:"int_array,omitempty"`
	}

	schema := SchemaFromStruct(TestStruct{})

	assert.NotNil(t, schema)
	assert.Equal(t, "array", schema.Properties["string_slice"].Type)
	assert.Equal(t, "array", schema.Properties["int_array"].Type)
}

// TestSchemaFromStruct_AllNumericTypes 测试所有数值类型
func TestSchemaFromStruct_AllNumericTypes(t *testing.T) {
	type TestStruct struct {
		Int8Field    int8    `json:"int8_field"`
		Int16Field   int16   `json:"int16_field"`
		Int32Field   int32   `json:"int32_field"`
		Int64Field   int64   `json:"int64_field"`
		UintField    uint    `json:"uint_field"`
		Uint8Field   uint8   `json:"uint8_field"`
		Uint16Field  uint16  `json:"uint16_field"`
		Uint32Field  uint32  `json:"uint32_field"`
		Uint64Field  uint64  `json:"uint64_field"`
		Float32Field float32 `json:"float32_field"`
		Float64Field float64 `json:"float64_field"`
	}

	schema := SchemaFromStruct(TestStruct{})

	assert.NotNil(t, schema)
	// 所有整数类型应该是 "integer"
	assert.Equal(t, "integer", schema.Properties["int8_field"].Type)
	assert.Equal(t, "integer", schema.Properties["int16_field"].Type)
	assert.Equal(t, "integer", schema.Properties["int32_field"].Type)
	assert.Equal(t, "integer", schema.Properties["int64_field"].Type)
	assert.Equal(t, "integer", schema.Properties["uint_field"].Type)
	assert.Equal(t, "integer", schema.Properties["uint8_field"].Type)
	assert.Equal(t, "integer", schema.Properties["uint16_field"].Type)
	assert.Equal(t, "integer", schema.Properties["uint32_field"].Type)
	assert.Equal(t, "integer", schema.Properties["uint64_field"].Type)
	// 所有浮点类型应该是 "number"
	assert.Equal(t, "number", schema.Properties["float32_field"].Type)
	assert.Equal(t, "number", schema.Properties["float64_field"].Type)
}

// TestSchemaFromStruct_JsonTagName 测试 JSON 标签名称
func TestSchemaFromStruct_JsonTagName(t *testing.T) {
	type TestStruct struct {
		FieldName string `json:"custom_name"`
		NoName    string `json:",omitempty"` // 空名称应该使用字段名的小写
	}

	schema := SchemaFromStruct(TestStruct{})

	assert.NotNil(t, schema)
	assert.Contains(t, schema.Properties, "custom_name")
	assert.Contains(t, schema.Properties, "noname") // 空名称时使用字段名的小写
}

// TestSchemaFromStruct_RealWorldExample 测试真实世界的例子（CreateEventRequest）
func TestSchemaFromStruct_RealWorldExample(t *testing.T) {
	type CreateEventRequest struct {
		DtStart     time.Time  `json:"dtstart"`
		DtEnd       *time.Time `json:"dtend,omitempty"`
		Duration    *string    `json:"duration,omitempty"`
		Summary     *string    `json:"summary,omitempty"`
		Description *string    `json:"description,omitempty"`
		Location    *string    `json:"location,omitempty"`
	}

	schema := SchemaFromStruct(CreateEventRequest{})

	assert.NotNil(t, schema)
	assert.Equal(t, "object", schema.Type)

	// 检查必需字段
	assert.Contains(t, schema.Required, "dtstart")
	assert.NotContains(t, schema.Required, "dtend")
	assert.NotContains(t, schema.Required, "duration")
	assert.NotContains(t, schema.Required, "summary")
	assert.NotContains(t, schema.Required, "description")
	assert.NotContains(t, schema.Required, "location")

	// 检查字段类型
	assert.Equal(t, "string", schema.Properties["dtstart"].Type)
	assert.Equal(t, "date-time", schema.Properties["dtstart"].Format)
	assert.Equal(t, "string", schema.Properties["dtend"].Type)
	assert.Equal(t, "date-time", schema.Properties["dtend"].Format)
	assert.Equal(t, "string", schema.Properties["duration"].Type)
	assert.Equal(t, "string", schema.Properties["summary"].Type)
	assert.Equal(t, "string", schema.Properties["description"].Type)
	assert.Equal(t, "string", schema.Properties["location"].Type)
}

// TestSchemaFromStruct_Description 测试描述字段
func TestSchemaFromStruct_Description(t *testing.T) {
	type TestStruct struct {
		FieldName string `json:"field_name"`
	}

	schema := SchemaFromStruct(TestStruct{})

	assert.NotNil(t, schema)
	assert.Equal(t, "FieldName", schema.Properties["field_name"].Description)
}

// TestSchemaFromStruct_UnknownType 测试未知类型（应该默认为 string）
func TestSchemaFromStruct_UnknownType(t *testing.T) {
	type CustomType string

	type TestStruct struct {
		CustomField CustomType `json:"custom_field"`
	}

	schema := SchemaFromStruct(TestStruct{})

	assert.NotNil(t, schema)
	// 未知类型应该默认为 string
	assert.Equal(t, "string", schema.Properties["custom_field"].Type)
}

// TestSchemaFromStruct_MultipleRequiredFields 测试多个必需字段
func TestSchemaFromStruct_MultipleRequiredFields(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
		Field3 bool   `json:"field3"`
		Field4 string `json:"field4,omitempty"`
	}

	schema := SchemaFromStruct(TestStruct{})

	assert.NotNil(t, schema)
	assert.Len(t, schema.Required, 3)
	assert.Contains(t, schema.Required, "field1")
	assert.Contains(t, schema.Required, "field2")
	assert.Contains(t, schema.Required, "field3")
	assert.NotContains(t, schema.Required, "field4")
}

// TestSchemaFromStruct_AllOptionalFields 测试所有字段都是可选的
func TestSchemaFromStruct_AllOptionalFields(t *testing.T) {
	type TestStruct struct {
		Field1 *string `json:"field1,omitempty"`
		Field2 *int    `json:"field2,omitempty"`
		Field3 *bool   `json:"field3,omitempty"`
	}

	schema := SchemaFromStruct(TestStruct{})

	assert.NotNil(t, schema)
	assert.Empty(t, schema.Required)
	assert.Len(t, schema.Properties, 3)
}

// TestSchemaFromStruct_NoRequiredFields 测试没有必需字段
func TestSchemaFromStruct_NoRequiredFields(t *testing.T) {
	type TestStruct struct {
		Optional1 string `json:"optional1,omitempty"`
		Optional2 *int   `json:"optional2,omitempty"`
	}

	schema := SchemaFromStruct(TestStruct{})

	assert.NotNil(t, schema)
	assert.Empty(t, schema.Required)
}

// TestSchemaFromStruct_Consistency 测试一致性（多次调用应该返回相同结果）
func TestSchemaFromStruct_Consistency(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	schema1 := SchemaFromStruct(TestStruct{})
	schema2 := SchemaFromStruct(TestStruct{})

	assert.NotNil(t, schema1)
	assert.NotNil(t, schema2)
	assert.Equal(t, schema1.Type, schema2.Type)
	assert.Equal(t, len(schema1.Properties), len(schema2.Properties))
	assert.Equal(t, len(schema1.Required), len(schema2.Required))
}

// TestSchemaFromStruct_ComplexExample 测试复杂示例
func TestSchemaFromStruct_ComplexExample(t *testing.T) {
	type ComplexStruct struct {
		ID        uint       `json:"id"`
		Name      string     `json:"name"`
		Email     *string    `json:"email,omitempty"`
		Age       int        `json:"age"`
		Score     float64    `json:"score"`
		IsActive  bool       `json:"is_active"`
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt *time.Time `json:"updated_at,omitempty"`
		Tags      []string   `json:"tags,omitempty"`
		Ignored   string     `json:"-"`
		NoJsonTag string
	}

	schema := SchemaFromStruct(ComplexStruct{})

	assert.NotNil(t, schema)
	assert.Equal(t, "object", schema.Type)

	// 检查必需字段
	requiredFields := []string{"id", "name", "age", "score", "is_active", "created_at"}
	for _, field := range requiredFields {
		assert.Contains(t, schema.Required, field, "字段 %s 应该是必需的", field)
	}

	// 检查可选字段不在 Required 中
	optionalFields := []string{"email", "updated_at", "tags"}
	for _, field := range optionalFields {
		assert.NotContains(t, schema.Required, field, "字段 %s 不应该是必需的", field)
	}

	// 检查被忽略的字段
	assert.NotContains(t, schema.Properties, "ignored")
	assert.NotContains(t, schema.Properties, "NoJsonTag")

	// 检查类型
	assert.Equal(t, "integer", schema.Properties["id"].Type)
	assert.Equal(t, "string", schema.Properties["name"].Type)
	assert.Equal(t, "string", schema.Properties["email"].Type)
	assert.Equal(t, "integer", schema.Properties["age"].Type)
	assert.Equal(t, "number", schema.Properties["score"].Type)
	assert.Equal(t, "boolean", schema.Properties["is_active"].Type)
	assert.Equal(t, "string", schema.Properties["created_at"].Type)
	assert.Equal(t, "date-time", schema.Properties["created_at"].Format)
	assert.Equal(t, "string", schema.Properties["updated_at"].Type)
	assert.Equal(t, "date-time", schema.Properties["updated_at"].Format)
	assert.Equal(t, "array", schema.Properties["tags"].Type)
}

func TestSchemaFromStruct_SearchCalendarItemsRequest(t *testing.T) {
	type TimeRangeInput struct {
		Start *string `json:"start,omitempty"`
		End   *string `json:"end,omitempty"`
	}
	type SearchCalendarItemsRequest struct {
		FieldKeywords map[string]string         `json:"field_keywords,omitempty"`
		TimeRanges    map[string]TimeRangeInput `json:"time_ranges,omitempty"`
	}

	schema := SchemaFromStruct(SearchCalendarItemsRequest{})
	assert.NotNil(t, schema)
	assert.Equal(t, "object", schema.Type)

	// 检查属性都存在，并为对象类型
	assert.Contains(t, schema.Properties, "field_keywords")
	assert.Equal(t, "object", schema.Properties["field_keywords"].Type)
	assert.Contains(t, schema.Properties, "time_ranges")
	assert.Equal(t, "object", schema.Properties["time_ranges"].Type)

	// 可选字段，不应在 required
	if schema.Required != nil {
		assert.NotContains(t, schema.Required, "field_keywords")
		assert.NotContains(t, schema.Required, "time_ranges")
	}
}
