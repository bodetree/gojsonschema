package gojsonschema

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

const schema = `
{
	"type": "object",
	"properties": {
		"num": {
			"type": "number",
			"range": [10, 20]
		},
		"num2": {
			"type": "number",
			"range": [20, 30]
		}
	}
}`

const deepSchema = `
{
  "type": "object",
  "properties": {
    "a": {
			"type": "object",
			"properties": {
				"b" : {
					"type": "object",
					"properties": {
						"c": {
							"type": "number",
							"range": [10, 20]
						}
					}
				},
				"d": {
					"type": "number",
					"range": [20, 30]
				}
			}
		}
	}
}`

const invalidSchema = `
{
  "type": "object",
  "properties": {
    "num": {
      "type": "number",
      "range": "invalid"
    },
		"num2": {
			"type": "number",
			"range": [20, 30]
		}
  }
}`

type testRangeKeyword struct{}

func (testRangeKeyword) GetKeyword() string {
	return "range"
}

func (testRangeKeyword) ValidateSchema(keywordValue interface{}) (err error) {
	err = errors.New("range must be an array containing exactly 2 numbers")
	kv, ok := keywordValue.([]interface{})
	if !ok || len(kv) != 2 {
		return err
	}

	_, ok = kv[0].(json.Number)
	if !ok {
		return err
	}

	_, ok = kv[1].(json.Number)
	if !ok {
		return err
	}

	return nil
}

func (testRangeKeyword) Validate(keywordValue interface{}, documentValue interface{}) (*CustomKeywordError, ErrorDetails) {
	dv, _ := documentValue.(json.Number)
	value, _ := dv.Float64()

	kv, _ := keywordValue.([]interface{})

	_min, _max := kv[0].(json.Number), kv[1].(json.Number)
	min, _ := _min.Float64()
	max, _ := _max.Float64()

	details := ErrorDetails{"min": min, "max": max}

	if value < min || value > max {
		return NewCustomKeywordError("range", "must be between {{.min}} and {{.max}}"), details
	}

	return nil, nil
}

func TestCustomKeywordDeepSchema(t *testing.T) {
	validDocLoader := NewStringLoader(`{"a":{"b":{"c":15,"d":25}}}`)
	invalidDocLoader := NewStringLoader(`{"a":{"b":{"c":1,"d":50}}}`)

	schemaLoader := NewStringLoader(deepSchema)
	schemaLoader.AddCustomKeyword(testRangeKeyword{})

	schema, err := NewSchema(schemaLoader)
	if !assert.NoError(t, err) {
		return
	}

	{
		result, err := schema.Validate(validDocLoader)
		if !assert.NoError(t, err) {
			return
		}

		assert.Len(t, result.Errors(), 0)
	}

	{
		result, err := schema.Validate(invalidDocLoader)
		if !assert.NoError(t, err) {
			return
		}

		assert.Len(t, result.Errors(), 1)
	}
}

func TestCustomKeywordPass(t *testing.T) {
	documentLoader := NewStringLoader(`{"num":15}`)
	schemaLoader := NewStringLoader(schema)
	schemaLoader.AddCustomKeyword(testRangeKeyword{})

	schema, err := NewSchema(schemaLoader)
	if !assert.NoError(t, err) {
		return
	}

	result, err := schema.Validate(documentLoader)
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, result.Errors(), 0)
}

func TestCustomKeywordFail(t *testing.T) {
	documentLoader := NewStringLoader(`{"num":1,"num2":100}`)
	schemaLoader := NewStringLoader(schema)
	schemaLoader.AddCustomKeyword(testRangeKeyword{})

	schema, err := NewSchema(schemaLoader)
	if !assert.NoError(t, err) {
		return
	}

	result, err := schema.Validate(documentLoader)
	if !assert.NoError(t, err) {
		return
	}

	assert.False(t, result.Valid())
	assert.Len(t, result.Errors(), 2)
}

func TestCustomKeywordInvalidSchema(t *testing.T) {
	schemaLoader := NewStringLoader(invalidSchema)
	schemaLoader.AddCustomKeyword(testRangeKeyword{})

	_, err := NewSchema(schemaLoader)
	assert.Error(t, err)
	assert.EqualError(t, err, "range must be an array containing exactly 2 numbers")
}
