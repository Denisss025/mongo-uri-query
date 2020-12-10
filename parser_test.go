package query

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint:paralleltest
func TestMapValues(t *testing.T) {
	floatValues := []float64{1.1, 2.2, -555.8}

	stringValues := make([]string, len(floatValues))
	ix := make([]interface{}, len(floatValues))

	for i, v := range floatValues {
		stringValues[i] = strconv.FormatFloat(v, 'f', 1, 64)
		ix[i] = v
	}

	result, err := mapValues(stringValues, Double())
	assert.NoError(t, err)
	assert.Equal(t, ix, result)

	stringValues = append(stringValues, "string")
	result, err = mapValues(stringValues, Double())

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestParse(ts *testing.T) {
	ts.Parallel()

	conv := NewDefaultConverter(testOidPrimitive{})
	testValues := []string{"yes", "123456789012", "test", "123", "213.0"}
	expect := []interface{}{
		true,
		testObjectID{oid: "123456789012"},
		"test",
		int64(123), 213.0,
	}

	ts.Run("correct parse routine", func(t *testing.T) {
		t.Parallel()

		v, err := parse(testValues, operatorEqualArray, conv)
		assert.NoError(t, err)
		assert.Equal(t, expect, v)
	})

	ts.Run("$eq for multiple values returns error", func(t *testing.T) {
		t.Parallel()

		_, err := parse(testValues, operatorEquals, conv)
		assert.EqualError(t, err, ErrTooManyValues.Error())
	})

	ts.Run("$eq for a string returns that string", func(t *testing.T) {
		t.Parallel()

		test := strings.Join(testValues, ",")
		v, err := parse([]string{test}, operatorEquals, conv)
		assert.NoError(t, err)
		assert.Equal(t, test, v)
	})

	ts.Run("split string for multival operator", func(t *testing.T) {
		t.Parallel()

		test := strings.Join(testValues, ",")
		v, err := parse([]string{test}, operatorContainsIn, conv)
		assert.NoError(t, err)
		assert.Equal(t, expect, v)
	})

	ts.Run("return error without converter", func(t *testing.T) {
		t.Parallel()

		_, err := parse(testValues, operatorIn, nil)
		assert.EqualError(t, err, ErrNoConverter.Error())
	})

	ts.Run("return an error", func(t *testing.T) {
		t.Parallel()

		_, err := parse(testValues, operatorAll, Bool())
		assert.True(t, errors.Is(err, ErrNoMatch),
			"unexpected err: %v", err)
	})

	ts.Run("return nil for empty array", func(t *testing.T) {
		t.Parallel()

		v, err := parse(nil, operatorNotEquals, Int())
		assert.NoError(t, err)
		assert.Nil(t, v)
	})
}

//nolint:paralleltest
func TestParseIntParam(t *testing.T) {
	params := url.Values{
		"__test1": []string{"10"},
		"__test2": []string{"20", "30"},
		"__test3": []string{"yes", "40"},
	}

	i, err := parseIntParam(params, "test1")
	assert.NoError(t, err)
	assert.EqualValues(t, 10, i)

	i, err = parseIntParam(params, "test2")
	assert.NoError(t, err)
	assert.EqualValues(t, 20, i)

	_, err = parseIntParam(params, "test3")
	assert.Error(t, err)

	i, err = parseIntParam(params, "test4")
	assert.NoError(t, err)
	assert.Zero(t, i)
}

func TestParserRegexEscape(ts *testing.T) {
	ts.Parallel()

	var p Parser

	ts.Run("escaped", func(t *testing.T) {
		t.Parallel()

		test := "^([0-9]?.*){1,2}|n/a+$"
		expected := "\\^\\(\\[0\\-9\\]\\?\\.\\*\\)\\{1,2\\}\\|n/a\\+\\$"

		acquired := p.regEscape(test)
		assert.Equal(t, expected, acquired)
	})

	ts.Run("noescape", func(t *testing.T) {
		t.Parallel()

		test := "0xabcdef"
		acquired := p.regEscape(test)
		assert.Equal(t, test, acquired)
	})
}

func TestParserGetValue(ts *testing.T) {
	ts.Parallel()

	p := Parser{
		Converter: NewDefaultConverter(testOidPrimitive{}),
		Fields: Fields{
			"field1": Field{Required: true},
			"field2": Field{Converter: Int()},
			"field3": Field{
				Required:  true,
				Converter: Bool(),
			},
		},
	}

	p2 := Parser{
		Converter:      NewDefaultConverter(testOidPrimitive{}),
		Fields:         p.Fields,
		ValidateFields: true,
	}

	ts.Run("invalid operator", func(t *testing.T) {
		t.Parallel()

		_, _, _, err := p.getValue("field1__test", []string{"1"})
		assert.EqualError(t, err,
			fmt.Sprintf("%v: test", ErrUnknownOperator))
	})

	ts.Run("validate fields: no field spec", func(t *testing.T) {
		t.Parallel()

		_, _, _, err := p2.getValue("field4", []string{"1"})
		assert.EqualError(t, err,
			fmt.Sprintf("%v: field4", ErrNoFieldSpec))
	})

	ts.Run("parse error", func(t *testing.T) {
		t.Parallel()

		_, _, _, err := p2.getValue("field3", []string{"1"})
		assert.EqualError(t, err,
			fmt.Sprintf("convert field3: %v", ErrNoMatch))
	})

	ts.Run("operator exists must be boolean", func(t *testing.T) {
		t.Parallel()

		field, op, val, err := p.getValue("test__exists",
			[]string{"yes"})
		assert.NoError(t, err)
		assert.Equal(t, "test", field)
		assert.Equal(t, true, val)
		assert.Equal(t, operatorExists, op)

		_, _, _, err = p.getValue("test__exists", []string{"1"})
		assert.EqualError(t, err,
			fmt.Sprintf("convert test: %v", ErrNoMatch))
	})

	ts.Run("regex operator", func(t *testing.T) {
		t.Parallel()

		field, op, val, err := p.getValue("test__rein",
			[]string{"[0-9]*", "[a-f]*"})
		assert.NoError(t, err)
		assert.Equal(t, "test", field)
		assert.Equal(t, operatorRegexIn, op)
		assert.Equal(t, []interface{}{
			testRegEx{regex: "[0-9]*"}, testRegEx{regex: "[a-f]*"},
		}, val)
	})

	ts.Run("starts-with operator", func(t *testing.T) {
		t.Parallel()

		field, op, val, err := p.getValue("test__isw", []string{"^"})
		assert.NoError(t, err)
		assert.Equal(t, "test", field)
		assert.Equal(t, operatorStartsWithIgnoreCase, op)
		assert.Equal(t, testRegEx{regex: "^\\^", options: "i"}, val)
	})

	ts.Run("contains operator", func(t *testing.T) {
		t.Parallel()

		field, op, val, err := p.getValue("test__icoin", []string{"$"})
		assert.NoError(t, err)
		assert.Equal(t, "test", field)
		assert.Equal(t, operatorContainsInIgnoreCase, op)
		assert.Equal(t, []interface{}{
			testRegEx{regex: "\\$", options: "i"},
		}, val)
	})
}

//nolint:paralleltest
func TestGetSortFields(t *testing.T) {
	fields := getSortFields(url.Values{})
	assert.Len(t, fields, 0)

	fields = getSortFields(url.Values{
		"__sort": []string{"a,b,-c", "d", "e,f"},
	})

	assert.Equal(t, []string{"a", "b", "-c", "d", "e", "f"}, fields)
}

func TestParserParseFields(ts *testing.T) {
	ts.Parallel()

	p := Parser{
		Converter: NewDefaultConverter(testOidPrimitive{}),
		Fields: Fields{
			"required": Field{
				Required:  true,
				Converter: Bool(),
			},
		},
		ValidateFields: true,
	}

	ts.Run("ignore directives", func(t *testing.T) {
		t.Parallel()

		filter, err := p.parseFilter(url.Values{
			"required": []string{"yes"},
			"__limit":  []string{"25"},
			"__skip":   []string{"75"},
			"__sort":   []string{"x,y,z"},
		})

		assert.Nil(t, err)
		assert.NotNil(t, filter.Filter)
		assert.Equal(t, M{"required": true}, filter.Filter)
		assert.Nil(t, filter.Sort)
		assert.Zero(t, filter.Limit)
		assert.Zero(t, filter.Skip)
	})

	ts.Run("no required field", func(t *testing.T) {
		t.Parallel()

		filter, err := p.parseFilter(url.Values{
			"__limit": []string{"25"},
			"__skip":  []string{"75"},
			"__sort":  []string{"x,y,z"},
		})

		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, ErrMissingField))
		assert.Nil(t, filter.Filter)
		assert.Nil(t, filter.Sort)
		assert.Zero(t, filter.Limit)
		assert.Zero(t, filter.Skip)
	})

	ts.Run("bad conversion", func(t *testing.T) {
		t.Parallel()

		filter, err := p.parseFilter(url.Values{
			"required": []string{"test"},
			"__limit":  []string{"25"},
			"__skip":   []string{"75"},
			"__sort":   []string{"x,y,z"},
		})

		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, ErrNoMatch))
		assert.Nil(t, filter.Filter)
		assert.Nil(t, filter.Sort)
		assert.Zero(t, filter.Limit)
		assert.Zero(t, filter.Skip)
	})
}

func TestParserParse(ts *testing.T) {
	ts.Parallel()

	p := Parser{
		Converter: NewDefaultConverter(testOidPrimitive{}),
		Fields: Fields{
			"required": Field{
				Required:  true,
				Converter: Bool(),
			},
		},
		ValidateFields: true,
	}

	ts.Run("bad skip parameter", func(t *testing.T) {
		t.Parallel()

		filter, err := p.Parse(url.Values{
			"required": []string{"yes"},
			"__skip":   []string{"required"},
			"__limit":  []string{"10"},
		})

		assert.Error(t, err)
		assert.NotNil(t, filter.Filter)
		assert.True(t, filter.Filter["required"].(bool))
		assert.Zero(t, filter.Skip)
		assert.EqualValues(t, 10, filter.Limit)
	})

	ts.Run("bad limit parameter", func(t *testing.T) {
		t.Parallel()

		filter, err := p.Parse(url.Values{
			"required": []string{"no"},
			"__limit":  []string{"ten"},
			"__skip":   []string{"1000"},
		})

		assert.Error(t, err)
		assert.NotNil(t, filter.Filter)
		assert.False(t, filter.Filter["required"].(bool))
		assert.Zero(t, filter.Limit)
		assert.EqualValues(t, 1000, filter.Skip)
		assert.Nil(t, filter.Sort)
	})

	ts.Run("sort without spec", func(t *testing.T) {
		t.Parallel()

		filter, err := p.Parse(url.Values{
			"required": []string{"no"},
			"__sort":   []string{"field"},
		})

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNoSortField))
		assert.NotNil(t, filter.Filter)
		assert.False(t, filter.Filter["required"].(bool))
		assert.Zero(t, filter.Limit)
		assert.Zero(t, filter.Skip)
		assert.Len(t, filter.Sort, 1)
	})

	ts.Run("bad field conversion", func(t *testing.T) {
		t.Parallel()

		filter, err := p.Parse(url.Values{
			"required": []string{"nope"},
		})

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrMissingField) ||
			errors.Is(err, ErrNoMatch))
		assert.Nil(t, filter.Filter)
		assert.Zero(t, filter.Limit)
		assert.Zero(t, filter.Skip)
		assert.Len(t, filter.Sort, 0)
	})

	ts.Run("normal request", func(t *testing.T) {
		t.Parallel()

		filter, err := p.Parse(url.Values{
			"__sort":           []string{"-required"},
			"required__exists": []string{"true"},
		})

		assert.NoError(t, err)
		assert.Zero(t, filter.Skip)
		assert.Zero(t, filter.Limit)
		assert.Equal(t, map[string]int{"required": -1}, filter.Sort)
		assert.Equal(t, M{"required": M{"$exists": true}},
			filter.Filter)
	})
}
