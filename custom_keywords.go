package gojsonschema

type CustomKeyword interface {
	GetKeyword() string
	Validate(keywordValue interface{}, documentValue interface{}) (*CustomKeywordError, ErrorDetails)
	ValidateSchema(keywordValue interface{}) error
}

type customKeywordValue struct {
	value         interface{}
	customKeyword CustomKeyword
}

type CustomKeywordError struct {
	Keyword   string
	LocaleStr string
	ResultErrorFields
}

func NewCustomKeywordError(keyword, localeStr string) *CustomKeywordError {
	return &CustomKeywordError{
		Keyword:   keyword,
		LocaleStr: localeStr,
	}
}
