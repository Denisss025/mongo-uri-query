package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint:paralleltest
func TestAddSort(t *testing.T) {
	var q Query

	assert.Nil(t, q.Sort)

	q.AddSort("test")
	assert.Len(t, q.Sort, 1)
	assert.Equal(t, 1, q.Sort["test"])

	q.AddSort("-test")
	assert.Len(t, q.Sort, 1)
	assert.Equal(t, -1, q.Sort["test"])

	q.AddSort("field")
	assert.Len(t, q.Sort, 2)
	assert.Equal(t, -1, q.Sort["test"])
	assert.Equal(t, 1, q.Sort["field"])
}

//nolint:paralleltest
func TestAppendArray(t *testing.T) {
	var arr interface{}

	val := interface{}("1")

	arr = appendArray(arr, val)
	assert.NotNil(t, arr)
	assert.Len(t, arr, 1)
	assert.Equal(t, []interface{}{val}, arr)

	arr = appendArray(arr, val)
	assert.Len(t, arr, 2)
	assert.Equal(t, []interface{}{val, val}, arr)

	arr = appendArray(arr, arr)
	assert.Len(t, arr, 4)
	assert.Equal(t, []interface{}{val, val, val, val}, arr)

	arr = appendArray(val, arr)
	assert.Len(t, arr, 5)

	arr = appendArray(val, val)
	assert.Len(t, arr, 2)

	arr2 := appendArray(nil, arr)
	assert.Len(t, arr2, 2)
	assert.Equal(t, arr, arr2)
}

//nolint:paralleltest
func TestAddFilter(t *testing.T) {
	var q Query

	val := interface{}("value")

	q.AddFilter("field", operatorEquals, val)
	assert.Len(t, q.Filter, 1)
	assert.Equal(t, val, q.Filter["field"])

	arr := appendArray(val, val)
	q.Filter = nil
	q.AddFilter("field", operatorIn, arr)
	assert.Len(t, q.Filter, 1)
	assert.Equal(t, M{"$in": []interface{}{val, val}}, q.Filter["field"])

	q.AddFilter("field2", operatorEquals, val)
	assert.Len(t, q.Filter, 2)
	assert.Equal(t, M{"$in": []interface{}{val, val}}, q.Filter["field"])
	assert.Equal(t, val, q.Filter["field2"])

	q.AddFilter("field3", operatorGreaterThan, val)
	assert.Len(t, q.Filter, 3)
	assert.Equal(t, M{"$gt": val}, q.Filter["field3"])

	q.AddFilter("field", operatorIn, val)
	assert.Len(t, q.Filter, 3)
	assert.Equal(t, M{"$in": []interface{}{val, val, val}},
		q.Filter["field"])

	q.AddFilter("field", operatorNotIn, val)
	assert.Len(t, q.Filter, 3)
	assert.Equal(t, M{
		"$in":  []interface{}{val, val, val},
		"$nin": []interface{}{val},
	},
		q.Filter["field"])

	q.Filter = nil
	q.AddFilter("field", operatorEquals, val)
	q.AddFilter("field", operatorEqualArray, val)

	assert.Len(t, q.Filter, 1)
	assert.Equal(t, M{"$eq": []interface{}{val, val}}, q.Filter["field"])
}
