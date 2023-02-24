package schema

import (
	"github.com/xeipuuv/gojsonschema"

	_ "embed"
)

//go:embed testdata/config.schema.json
var rawJSONSchema []byte

var schemaLoader gojsonschema.JSONLoader

func Validate(raw []byte) error {
	result, err := gojsonschema.Validate(schemaLoader, gojsonschema.NewBytesLoader(raw))
	if err != nil {
		return NewSchemaError(err)
	}
	if !result.Valid() {
		return NewValidationError(result.Errors())
	}
	return nil
}

func init() {
	schemaLoader = gojsonschema.NewBytesLoader(rawJSONSchema)
}
