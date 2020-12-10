package query

import (
	"errors"
	"strings"
)

var (
	// ErrNoMatch is returned when the converter cannot match a string
	// value with any pattern.
	ErrNoMatch = errors.New("does not match")
	// ErrUnknownOperator is returned when an unknown operator is found.
	ErrUnknownOperator = errors.New("unknown operator")
	// ErrNoFieldSpec is returned when there is no field specification, but
	// the field is present in a query.
	ErrNoFieldSpec = errors.New("no field spec")
	// ErrNoConverter is returned when there is no converter found.
	ErrNoConverter = errors.New("no converter")
	// ErrNoSortField is returned when there is no field specification for
	// the sort field from the query.
	ErrNoSortField = errors.New("no sort field spec")
	// ErrMissingField is returned when some required field is missing in
	// the query.
	ErrMissingField = errors.New("missing required filter on field")
	// ErrTooManyValues is returned when a single value operator is assigned
	// to multiple values.
	ErrTooManyValues = errors.New("too many values")
)

// M is an alias for map[string]interface{}.
type M = map[string]interface{}

// Query is a structure that holds information about DB request.
type Query struct {
	// Filter is a document containing query operators.
	Filter M
	// Sort is a document specifying the order in which documents should
	// be returned.
	Sort map[string]int
	// Limit is the maximum number of documents to return.
	Limit int64
	// Skip is a number of documents to be skipped before adding documents
	// to the results.
	Skip int64
}

func appendArray(array, values interface{}) (retval interface{}) {
	type mongoArray = []interface{}

	fArray, isFArray := array.(mongoArray)
	vArray, isVarray := values.(mongoArray)

	if array != nil && isFArray {
		if isVarray {
			return append(fArray, vArray...)
		}

		return append(fArray, values)
	}

	if isVarray {
		fArray = make(mongoArray, len(vArray)+1)

		if array == nil {
			copy(fArray, vArray)
			fArray = fArray[:len(fArray)-1]
		} else {
			fArray[0] = array
			copy(fArray[1:], vArray)
		}

		return fArray
	}

	if array == nil {
		return mongoArray{values}
	}

	return mongoArray{array, values}
}

func addField(filter M, field string, op operator, val interface{}) (m M) {
	if m = filter; m == nil {
		m = make(M)
	}

	f, exists := m[field]
	if !exists {
		f = nil
	}

	mm, isMap := f.(M)
	if !isMap {
		if op == operatorEquals {
			m[field] = val

			return m
		}

		if f != nil {
			mm = M{operatorEquals.MongoOperator(): f}
		} else {
			mm = make(M)
		}
	}

	if op.IsMultiVal() {
		var arr interface{}

		if marr, hasOperator := mm[op.MongoOperator()]; hasOperator {
			arr = marr
		}

		val = appendArray(arr, val)
	}

	mm[op.MongoOperator()] = val
	m[field] = mm

	return m
}

// AppendFilter appends an operator, field and value to the filter.
func (f *Query) AddFilter(field string, op operator, value interface{}) {
	f.Filter = addField(f.Filter, field, op, value)
}

// AddSort adds a field to sort to the Sort document.
func (f *Query) AddSort(val string) (fieldName string) {
	if f.Sort == nil {
		f.Sort = make(map[string]int)
	}

	sortDirection := sortAsc

	fieldName = strings.TrimPrefix(val, sortAscPrefix)

	if strings.HasPrefix(fieldName, sortDescPrefix) {
		sortDirection, fieldName = sortDesc, fieldName[1:]
	}

	f.Sort[fieldName] = sortDirection

	return
}
