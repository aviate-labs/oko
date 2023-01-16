package schema_test

import (
	"testing"

	"github.com/internet-computer/oko/config/schema"
)

func TestValidate(t *testing.T) {
	if err := schema.Validate([]byte("{}")); err == nil {
		t.Error()
	}
	if err := schema.Validate([]byte(`{
		"dependencies": []
	}`)); err != nil {
		t.Error(err)
	}
	if err := schema.Validate([]byte(`{
		"dependencies": [
			{
				"name": "test",
				"alts": [ "t" ],
				"repository": "url",
				"version": "v0.0.1"
			}
		]
	}`)); err != nil {
		t.Error(err)
	}
}
