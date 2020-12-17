package query

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
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

func TestConvertArray(ts *testing.T) {
	ts.Parallel()

	conv := NewDefaultConverter(testOidPrimitive{})
	testValues := []string{"yes", "123456789012", "test", "123", "213.0"}
	expect := []interface{}{
		true,
		testObjectID{oid: "123456789012"},
		"test",
		int64(123), 213.0,
	}

	ts.Run("correct convertArray routine", func(t *testing.T) {
		t.Parallel()

		v, err := convertArray(testValues, operatorEqualArray, conv)
		assert.NoError(t, err)
		assert.Equal(t, expect, v)
	})

	ts.Run("$eq for multiple values returns error", func(t *testing.T) {
		t.Parallel()

		_, err := convertArray(testValues, operatorEquals, conv)
		assert.EqualError(t, err, ErrTooManyValues.Error())
	})

	ts.Run("$eq for a string returns that string", func(t *testing.T) {
		t.Parallel()

		test := strings.Join(testValues, ",")
		v, err := convertArray([]string{test}, operatorEquals, conv)
		assert.NoError(t, err)
		assert.Equal(t, test, v)
	})

	ts.Run("return error without converter", func(t *testing.T) {
		t.Parallel()

		_, err := convertArray(testValues, operatorIn, nil)
		assert.EqualError(t, err, ErrNoConverter.Error())
	})

	ts.Run("return an error", func(t *testing.T) {
		t.Parallel()

		_, err := convertArray(testValues, operatorAll, Bool())
		assert.True(t, errors.Is(err, ErrNoMatch),
			"unexpected err: %v", err)
	})

	ts.Run("return nil for empty array", func(t *testing.T) {
		t.Parallel()

		v, err := convertArray(nil, operatorNotEquals, Int())
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

	ts.Run("regex should return nil", func(t *testing.T) {
		t.Parallel()

		conv := p.regex("i", nop())
		assert.Nil(t, conv)
	})

	ts.Run("sort should not panic", func(t *testing.T) {
		t.Parallel()

		f, err := p.Parse(url.Values{"__sort": []string{"x"}})
		assert.Error(t, err)
		assert.Nil(t, f.Sort)
	})
}

func TestParserConvert(ts *testing.T) {
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

	ts.Run("validate fields: no field spec", func(t *testing.T) {
		t.Parallel()

		_, err := p2.convert("field4", operatorEquals, []string{"1"})
		assert.EqualError(t, err,
			fmt.Sprintf("convert: %v: field4", ErrNoFieldSpec))
	})

	ts.Run("convertArray error", func(t *testing.T) {
		t.Parallel()

		_, err := p2.convert("field3", operatorEquals, []string{"1"})
		assert.EqualError(t, err,
			fmt.Sprintf("convert: %v: field3", ErrNoMatch))
	})

	ts.Run("operator exists must be boolean", func(t *testing.T) {
		t.Parallel()

		val, err := p.convert("test", operatorExists, []string{"yes"})
		assert.NoError(t, err)
		assert.Equal(t, true, val)

		_, err = p.convert("test", operatorExists, []string{"1"})
		assert.EqualError(t, err,
			fmt.Sprintf("convert: %v: test", ErrNoMatch))
	})

	ts.Run("regex operator", func(t *testing.T) {
		t.Parallel()

		val, err := p.convert("test", operatorRegexIn,
			[]string{"[0-9]*", "[a-f]*"})
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{
			testRegEx{regex: "[0-9]*"}, testRegEx{regex: "[a-f]*"},
		}, val)
	})

	ts.Run("starts-with operator", func(t *testing.T) {
		t.Parallel()

		val, err := p.convert("test", operatorStartsWithIgnoreCase,
			[]string{"^"})
		assert.NoError(t, err)
		assert.Equal(t, testRegEx{regex: "^\\^", options: "i"}, val)
	})

	ts.Run("contains operator", func(t *testing.T) {
		t.Parallel()

		val, err := p.convert("test", operatorContainsInIgnoreCase,
			[]string{"$,x"})
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{
			testRegEx{regex: "\\$,x", options: "i"},
		}, val)
	})

	ts.Run("unknown operator", func(t *testing.T) {
		t.Parallel()

		_, err := p.convert("test", operator("unknown[]"), []string{"a"})
		assert.EqualError(t, err, fmt.Sprintf("convert: %v: %v",
			ErrUnknownOperator, operator("unknown[]").CommonOperator()))
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
		Converter: NewDefaultConverter(testOidPrimitive{
			forbidSortFields: map[string]struct{}{"forbidden": {}},
		}),
		ValidateFields: true,
	}

	p.Fields = Fields{
		"required": Field{
			Required:  true,
			Converter: Bool(),
		},
		"forbidden": Field{
			Required:  false,
			Converter: p.Converter,
		},
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

	ts.Run("error on AddSort()", func(t *testing.T) {
		t.Parallel()

		_, err := p.Parse(url.Values{
			"required": []string{"no"},
			"__sort":   []string{"-forbidden"},
		})

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNoSortField))
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
		assert.Nil(t, filter.Sort)
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
		assert.Equal(t, []map[string]interface{}{{"required": -1}}, filter.Sort)
		assert.Equal(t, M{"required": M{"$exists": true}},
			filter.Filter)
	})
}

func TestParserParseMultivalue(ts *testing.T) {
	ts.Parallel()

	p := Parser{
		Converter: NewDefaultConverter(testOidPrimitive{}),
	}

	ts.Run("__in with single value should be treated as eq",
		func(t *testing.T) {
			t.Parallel()

			q, err := p.Parse(url.Values{"field__in": []string{"a"}})
			assert.NoError(t, err)
			assert.Equal(t, M{"field": "a"}, q.Filter)
		})

	ts.Run("__in parameter should split string with commas",
		func(t *testing.T) {
			t.Parallel()

			q, err := p.Parse(url.Values{"field__in": []string{"a,b"}})
			assert.NoError(t, err)
			assert.Equal(t, M{"field": M{"$in": []interface{}{"a", "b"}}},
				q.Filter)
		})

	ts.Run("[] should be treated as __in", func(t *testing.T) {
		t.Parallel()

		q, err := p.Parse(url.Values{"field[]": []string{"a", "b"}})
		assert.NoError(t, err)
		assert.Equal(t, M{"field": M{"$in": []interface{}{"a", "b"}}},
			q.Filter)
	})

	ts.Run("[] parameter should not split string with commas",
		func(t *testing.T) {
			t.Parallel()

			q, err := p.Parse(url.Values{"field[]": []string{"a,b"}})
			assert.NoError(t, err)
			assert.Equal(t, M{"field": "a,b"}, q.Filter)
		})

	ts.Run("treat re[] as rein", func(t *testing.T) {
		t.Parallel()

		q, err := p.Parse(url.Values{
			"field__rein": []string{"a"},
			"field__re[]": []string{"b"},
		})

		assert.NoError(t, err)
		assert.Len(t, q.Filter, 1)
		assert.NotNil(t, q.Filter["field"])
		assert.NotNil(t, q.Filter["field"].(M))
		assert.Len(t, q.Filter["field"], 1)

		in := q.Filter["field"].(M)["$in"]
		assert.NotNil(t, in)

		inArr := in.([]interface{})
		assert.Len(t, inArr, 2)

		i1, i2 := inArr[0].(testRegEx), inArr[1].(testRegEx)
		assert.True(t, i1.regex != i2.regex && (i1.regex == "a" || i1.regex == "b"))
		assert.True(t, i1.regex != i2.regex && (i2.regex == "a" || i2.regex == "b"))
		assert.Zero(t, i1.options)
		assert.Zero(t, i2.options)
	})
}

//nolint:paralleltest
func TestNormalizeFields(t *testing.T) {
	expected := fieldsMap{
		"field1": operatorsMap{
			operatorIn: []string{"a", "b", "c"},
		},
		"field2": operatorsMap{
			operatorEquals: []string{"a,b,c"},
		},
		"field3": operatorsMap{
			operatorIn: []string{"a,b,c", "d"},
		},
		"field4": operatorsMap{
			operatorIn: []string{"a", "b"},
		},
		"field5": operatorsMap{
			operatorEquals: []string{"a"},
		},
	}

	acquired := normailzeFields(fieldsMap{
		// split string
		"field1": operatorsMap{
			operatorIn: []string{"a,b,c"},
		},
		// convert $in to $eq, do not split string for []
		"field2": operatorsMap{
			operatorInArray: []string{"a,b,c"},
		},
		// do not split string
		"field3": operatorsMap{
			operatorIn: []string{"a,b,c", "d"},
		},
		// merge __in and [] to $in
		"field4": operatorsMap{
			operatorInArray: []string{"a"},
			operatorIn:      []string{"b"},
		},
		// convert $in to $eq
		"field5": operatorsMap{
			operatorIn: []string{"a"},
		},
	})

	sort.Strings(acquired["field4"][operatorIn])
	assert.Equal(t, expected, acquired)
}

func TestExtractFields(ts *testing.T) {
	ts.Parallel()

	ts.Run("simple", func(t *testing.T) {
		t.Parallel()

		expected := fieldsMap{
			"field1": operatorsMap{
				operatorIn: []string{"a", "b", "c"},
			},
			"field2": operatorsMap{
				operatorRegexIn: []string{"a", "b"},
			},
		}

		acquired := extractFields(url.Values{
			"field1__in":   []string{"a,b,c"},
			"field2__re[]": []string{"b"},
			"field2__rein": []string{"a"},
		})

		sort.Strings(acquired["field2"][operatorRegexIn])
		assert.Equal(t, expected, acquired)
	})

	ts.Run("rein and re[] should be merged", func(t *testing.T) {
		t.Parallel()

		expected := fieldsMap{
			"field": operatorsMap{
				operatorRegexIn: []string{"a", "b"},
			},
		}

		acquired := extractFields(url.Values{
			"field__rein": []string{"a"},
			"field__re[]": []string{"b"},
		})

		sort.Strings(acquired["field"][operatorRegexIn])
		assert.Equal(t, expected, acquired)
	})

	ts.Run("convert map[like][fields] to struct.like.fields",
		func(t *testing.T) {
			t.Parallel()

			expected := fieldsMap{
				"field1.nested.nested2": operatorsMap{
					operatorIn: []string{
						"a", "b", "c", "d",
					},
				},
			}

			acquired := extractFields(url.Values{
				"field1[nested][nested2][]": []string{"a", "b"},
				"field1.nested.nested2[]":   []string{"c"},
				"field1[nested[nested2]][]": []string{"d"},
			})

			sort.Strings(acquired["field1.nested.nested2"][operatorIn])
			assert.Equal(t, expected, acquired)
		})
}
