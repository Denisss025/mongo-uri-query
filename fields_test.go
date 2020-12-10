package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint:paralleltest
func TestFields(t *testing.T) {
	f := Fields{
		"field1": Field{
			Required:  true,
			Converter: Int(),
		},
		"field2": Field{
			Converter: Bool(),
		},
		"no-converter": Field{},
	}

	assert.True(t, f.HasField("field1"))
	assert.False(t, f.HasField("field3"))

	conv, hasField := f.Converter("field2")
	assert.NotNil(t, conv)
	assert.True(t, hasField)

	b, err := conv.Convert("yes")
	assert.NoError(t, err)
	assert.True(t, b.(bool))

	conv, hasField = f.Converter("no-converter")
	assert.Nil(t, conv)
	assert.True(t, hasField)

	conv, hasField = f.Converter("field3")
	assert.Nil(t, conv)
	assert.False(t, hasField)

	assert.True(t, f.IsRequired("field1"))
	assert.False(t, f.IsRequired("field2"))
	assert.False(t, f.IsRequired("field3"))
}
