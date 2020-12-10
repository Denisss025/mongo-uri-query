package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testIntStr      = "-789"
	testFloatStr    = "-555.888"
	testDateStr     = "2021-01-01"
	testTimeStr     = "2020-12-08T12:50:37Z"
	testTimeNSecStr = "2020-11-07T03:17:56.001Z"
	testObjectIDStr = "1234567890ab"
	testYesStr      = "yes"
	testNoStr       = "no"
	testTrueStr     = "true"
	testFalseStr    = "false"
)

type testObjectID struct {
	oid string
}

type testRegEx struct {
	regex, options string
}

type testOidPrimitive struct{}

func (t testOidPrimitive) RegEx(v, o string) (i interface{}, err error) {
	return testRegEx{regex: v, options: o}, nil
}

func (t testOidPrimitive) ObjectID(val string) (i interface{}, err error) {
	return testObjectID{oid: val}, nil
}

//nolint:paralleltest
func TestDefaultConvertFuncs(t *testing.T) {
	i, err := Int()(testIntStr)
	assert.NoError(t, err)
	assert.Equal(t, int64(-789), i)

	_, err = Int()(testFalseStr)
	assert.Error(t, err)

	i, err = Double()(testFloatStr)
	assert.NoError(t, err)
	assert.Equal(t, -555.888, i)

	_, err = Double()(testTrueStr)
	assert.Error(t, err)

	i, err = Date()(testDateStr)
	assert.NoError(t, err)
	assert.Equal(t,
		time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
		i)

	i, err = Date()(testTimeStr)
	assert.NoError(t, err)
	assert.Equal(t,
		time.Date(2020, time.December, 8, 12, 50, 37, 0, time.UTC),
		i)

	i, err = Date()(testTimeNSecStr)
	assert.NoError(t, err)
	assert.Equal(t,
		time.Date(2020, time.November, 7, 3, 17, 56, 1000000,
			time.UTC),
		i)

	_, err = Date()(testIntStr)
	assert.Error(t, err)

	i, err = Bool()(testYesStr)
	assert.NoError(t, err)
	assert.True(t, i.(bool))

	i, err = Bool()(testTrueStr)
	assert.NoError(t, err)
	assert.True(t, i.(bool))

	i, err = Bool()(testNoStr)
	assert.NoError(t, err)
	assert.False(t, i.(bool))

	i, err = Bool()(testFalseStr)
	assert.NoError(t, err)
	assert.False(t, i.(bool))

	_, err = Bool()(testFloatStr)
	assert.Error(t, err)

	i, err = ObjectID(testOidPrimitive{})(testObjectIDStr)
	assert.NoError(t, err)
	assert.Equal(t, testObjectID{oid: testObjectIDStr}, i)

	_, err = ObjectID(testOidPrimitive{})(testTimeStr[:12])
	assert.Error(t, err)

	_, err = ObjectID(testOidPrimitive{})(testTrueStr)
	assert.Error(t, err)

	i, err = String()(testYesStr)
	assert.NoError(t, err)
	assert.Equal(t, testYesStr, i)

	x, err2 := String().Convert(testYesStr)
	assert.Equal(t, err, err2)
	assert.Equal(t, i, x)
}

//nolint:paralleltest
func TestDefaultConverter(t *testing.T) {
	converter := NewDefaultConverter(testOidPrimitive{})

	i, err := converter.Convert("")
	assert.NoError(t, err)
	assert.Equal(t, "", i)

	i, err = converter.Convert(testIntStr)
	assert.NoError(t, err)
	assert.Equal(t, int64(-789), i)

	i, err = converter.Convert(testDateStr)
	expected, _ := Date()(testDateStr)

	assert.NoError(t, err)
	assert.Equal(t, expected, i)

	i, err = converter.Convert(testFloatStr)
	assert.NoError(t, err)
	assert.Equal(t, -555.888, i)

	i, err = converter.Convert(testObjectIDStr)
	assert.NoError(t, err)
	assert.Equal(t, testObjectID{oid: testObjectIDStr}, i)

	i, err = converter.Convert(testNoStr)
	assert.NoError(t, err)
	assert.False(t, i.(bool))

	// remove string to string converter
	converter.Funcs = converter.Funcs[:len(converter.Funcs)-1]
	_, err = converter.Convert("")
	assert.Error(t, err)
}
