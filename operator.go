package query

import "strings"

type operator string

// list of allowed operators.
const (
	ignoreCasePrefix = "i"
	mongoOpPrefix    = "$"

	operatorIn                  operator = "in"
	operatorInArray             operator = "[]"
	operatorEqualArray          operator = "eqa"
	operatorEquals              operator = "eq"
	operatorExists              operator = "exists"
	operatorGreaterThan         operator = "gt"
	operatorGreaterThanOrEquals operator = "gte"
	operatorLessThan            operator = "lt"
	operatorLessThanOrEquals    operator = "lte"
	operatorNotEquals           operator = "ne"
	operatorNotIn                        = "n" + operatorIn

	operatorAll operator = "all"

	operatorAllArray = operatorAll + operatorInArray

	operatorContains operator = "co"

	operatorContainsIgnoreCase   = ignoreCasePrefix + operatorContains
	operatorContainsIn           = operatorContains + operatorIn
	operatorContainsInIgnoreCase = ignoreCasePrefix + operatorContainsIn
	operatorContainsInArray      = operatorContains + operatorInArray

	operatorContainsInArrayIgnoreCase = ignoreCasePrefix +
		operatorContainsInArray

	operatorRegex operator = "re"

	operatorRegexIgnoreCase        = ignoreCasePrefix + operatorRegex
	operatorRegexIn                = operatorRegex + operatorIn
	operatorRegexInIgnoreCase      = ignoreCasePrefix + operatorRegexIn
	operatorRegexInArray           = operatorRegex + operatorInArray
	operatorRegexInArrayIgnoreCase = ignoreCasePrefix + operatorRegexInArray

	operatorStartsWith operator = "sw"

	operatorStartsWithIgnoreCase   = ignoreCasePrefix + operatorStartsWith
	operatorStartsWithIn           = operatorStartsWith + operatorIn
	operatorStartsWithInIgnoreCase = ignoreCasePrefix + operatorStartsWithIn
	operatorStartsWithInArray      = operatorStartsWith + operatorInArray

	operatorStartsWithInArrayIgnoreCase = ignoreCasePrefix +
		operatorStartsWithInArray

	allOperators = delimiter + operatorAll +
		delimiter + operatorAllArray +
		delimiter + operatorContains +
		delimiter + operatorContainsIgnoreCase +
		delimiter + operatorContainsIn +
		delimiter + operatorContainsInArray +
		delimiter + operatorContainsInArrayIgnoreCase +
		delimiter + operatorContainsInIgnoreCase +
		delimiter + operatorEqualArray +
		delimiter + operatorEquals +
		delimiter + operatorExists +
		delimiter + operatorGreaterThan +
		delimiter + operatorGreaterThanOrEquals +
		delimiter + operatorIn +
		delimiter + operatorInArray +
		delimiter + operatorLessThan +
		delimiter + operatorLessThanOrEquals +
		delimiter + operatorNotEquals +
		delimiter + operatorNotIn +
		delimiter + operatorRegex +
		delimiter + operatorRegexIgnoreCase +
		delimiter + operatorRegexIn +
		delimiter + operatorRegexInArray +
		delimiter + operatorRegexInArrayIgnoreCase +
		delimiter + operatorRegexInIgnoreCase +
		delimiter + operatorStartsWith +
		delimiter + operatorStartsWithIgnoreCase +
		delimiter + operatorStartsWithIn +
		delimiter + operatorStartsWithInArray +
		delimiter + operatorStartsWithInArrayIgnoreCase +
		delimiter + operatorStartsWithInIgnoreCase +
		delimiter
)

func parseOperator(fieldName string) (field string, op operator) {
	field, op = fieldName, operatorEquals

	if pos := strings.Index(field, delimiter); pos > 0 {
		op = operator(field[pos+len(delimiter):])
		field = field[:pos]
	} else if pos < 0 &&
		strings.HasSuffix(field, string(operatorInArray)) {
		op = operatorInArray
		field = field[:len(field)-len(op)]
	}

	return
}

func (o operator) String() (s string) { return string(o.CommonOperator()) }

// IsValid checks if an operator is in the list of the valid operators.
func (o operator) IsValid() (ok bool) {
	//nolint:gocritic
	// This is correct arguments order
	return strings.Contains(string(allOperators), string(o))
}

// IsMultiVal checks if an operator accepts multiple values.
func (o operator) IsMultiVal() (ok bool) {
	return o.Is(operatorIn) ||
		o.Is(operatorAll) ||
		o == operatorEqualArray
}

// NeedSplitString checks if an operator is multival and needs to split
// a string value into a slice.
func (o operator) NeedSplitString() (ok bool) {
	return o.IsMultiVal() && !o.Is(operatorInArray)
}

// SingleValueOperator returns a single value operator.
func (o operator) SingleValueOperator() (op operator) {
	commonOp := o.CommonOperator()
	if commonOp == operatorIn ||
		commonOp == operatorAll ||
		commonOp == operatorEquals {
		return operatorEquals
	}

	return operator(
		strings.TrimSuffix(string(commonOp), string(operatorIn)))
}

// Is checks if an operator is a subset of another operator
func (o operator) Is(op operator) (ok bool) {
	if strings.HasSuffix(string(op), string(operatorInArray)) {
		return strings.HasSuffix(string(o), string(op))
	}

	if o == operatorAllArray {
		return op == operatorAll || op == o
	}

	s := strings.ReplaceAll(string(o), string(operatorInArray),
		string(operatorIn))

	if !strings.HasPrefix(string(op), ignoreCasePrefix) {
		s = strings.TrimPrefix(s, ignoreCasePrefix)
	}

	if op == operatorIn {
		return strings.HasSuffix(s, string(op))
	}

	return strings.HasPrefix(s, string(op))
}

func (o operator) CommonOperator() (op operator) {
	if !o.Is(operatorInArray) {
		return o
	}

	if o == operatorAllArray {
		return operatorAll
	}

	return operator(strings.TrimSuffix(string(o),
		string(operatorInArray))) + operatorIn
}

// IsRegex checks if an operator is a RegEx operator, i.e. "re", "ire",
// "rein" and "irein".
func (o operator) IsRegex() (ok bool) {
	return o.Is(operatorRegex)
}

// IsStartsWith checks if an operator checks for the beginning of
// a string.
func (o operator) IsStartsWith() (ok bool) {
	return o.Is(operatorStartsWith)
}

// IsContains checks if an operator checks for the content of a string.
func (o operator) IsContains() (ok bool) {
	return o.Is(operatorContains)
}

// IsIgnoreCaseOperator checks if an operator has the Ignore Case flag.
func (o operator) IsIgnoreCaseOperator() (ok bool) {
	return o == operatorContainsInIgnoreCase ||
		o == operatorContainsIgnoreCase ||
		o == operatorRegexIgnoreCase ||
		o == operatorRegexInIgnoreCase ||
		o == operatorStartsWithIgnoreCase ||
		o == operatorStartsWithInIgnoreCase
}

// MongoOperator converts an operator to the mongo operator.
func (o operator) MongoOperator() (mongoOp string) {
	if o == operatorAllArray {
		return operatorAll.MongoOperator()
	}

	if o.IsMultiVal() && o != operatorAll && o != operatorEqualArray &&
		o != operatorNotIn {
		return mongoOpPrefix + string(operatorIn)
	}

	if o == operatorEqualArray || o.IsContains() ||
		o.IsRegex() || o.IsStartsWith() {
		return mongoOpPrefix + string(operatorEquals)
	}

	return mongoOpPrefix + string(o)
}

// RegexOpts gets RegEx options string.
func (o operator) RegexOpts() (opts string) {
	if o.IsIgnoreCaseOperator() {
		opts += ignoreCasePrefix
	}

	return
}
