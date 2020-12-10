package query

import "strings"

type operator string

// list of allowed operators.
const (
	ignoreCasePrefix = "i"
	mongoOpPrefix    = "$"

	operatorAll                 operator = "all"
	operatorIn                  operator = "in"
	operatorEqualArray          operator = "eqa"
	operatorEquals              operator = "eq"
	operatorExists              operator = "exists"
	operatorGreaterThan         operator = "gt"
	operatorGreaterThanOrEquals operator = "gte"
	operatorLessThan            operator = "lt"
	operatorLessThanOrEquals    operator = "lte"
	operatorNotEquals           operator = "ne"
	operatorNotIn                        = "n" + operatorIn

	operatorContains operator = "co"

	operatorContainsIgnoreCase   = ignoreCasePrefix + operatorContains
	operatorContainsIn           = operatorContains + operatorIn
	operatorContainsInIgnoreCase = ignoreCasePrefix + operatorContainsIn

	operatorRegex operator = "re"

	operatorRegexIgnoreCase   = ignoreCasePrefix + operatorRegex
	operatorRegexIn           = operatorRegex + operatorIn
	operatorRegexInIgnoreCase = ignoreCasePrefix + operatorRegexIn

	operatorStartsWith operator = "sw"

	operatorStartsWithIgnoreCase   = ignoreCasePrefix + operatorStartsWith
	operatorStartsWithIn           = operatorStartsWith + operatorIn
	operatorStartsWithInIgnoreCase = ignoreCasePrefix + operatorStartsWithIn

	allOperators = delimiter + operatorAll + delimiter +
		operatorIn + delimiter +
		operatorEqualArray + delimiter +
		operatorEquals + delimiter +
		operatorExists + delimiter +
		operatorGreaterThan + delimiter +
		operatorGreaterThanOrEquals + delimiter +
		operatorLessThan + delimiter +
		operatorLessThanOrEquals + delimiter +
		operatorNotEquals + delimiter +
		operatorNotIn + delimiter +
		operatorContains + delimiter +
		operatorContainsIgnoreCase + delimiter +
		operatorContainsIn + delimiter +
		operatorContainsInIgnoreCase + delimiter +
		operatorRegex + delimiter +
		operatorRegexIgnoreCase + delimiter +
		operatorRegexIn + delimiter +
		operatorRegexInIgnoreCase + delimiter +
		operatorStartsWith + delimiter +
		operatorStartsWithIgnoreCase + delimiter +
		operatorStartsWithIn + delimiter +
		operatorStartsWithInIgnoreCase + delimiter
)

// IsValid checks if an operator is in the list of the valid operators.
func (o operator) IsValid() (ok bool) {
	//nolint:gocritic
	// This is correct arguments order
	return strings.Contains(string(allOperators), string(o))
}

// IsMultiVal checks if an operator accepts multiple values.
func (o operator) IsMultiVal() (ok bool) {
	return o == operatorAll || o == operatorEqualArray ||
		strings.HasSuffix(string(o), string(operatorIn))
}

// IsRegexOperator checks if an operator is a RegEx operator, i.e. "re", "ire",
// "rein" and "irein".
func (o operator) IsRegexOperator() (ok bool) {
	return o == operatorRegex ||
		o == operatorRegexIgnoreCase ||
		o == operatorRegexIn ||
		o == operatorRegexInIgnoreCase
}

// IsStartsWithOperator checks if an operator checks for the beginning of
// a string.
func (o operator) IsStartsWithOperator() (ok bool) {
	return o == operatorStartsWith ||
		o == operatorStartsWithIgnoreCase ||
		o == operatorStartsWithIn ||
		o == operatorStartsWithInIgnoreCase
}

// IsContainsOperator checks if an operator checks for the content of a string.
func (o operator) IsContainsOperator() (ok bool) {
	return o == operatorContains ||
		o == operatorContainsIgnoreCase ||
		o == operatorContainsIn ||
		o == operatorContainsInIgnoreCase
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
	if o.IsMultiVal() && o != operatorAll && o != operatorEqualArray &&
		o != operatorNotIn {
		return mongoOpPrefix + string(operatorIn)
	}

	if o == operatorEqualArray || o.IsContainsOperator() ||
		o.IsRegexOperator() || o.IsStartsWithOperator() {
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
