package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint:paralleltest
func TestOperatorValidation(t *testing.T) {
	validOperators := []string{"all", "exists", "irein", "coin", "co"}
	invalidOperators := []string{"call", "nexists", "reini", "icon"}

	for _, valid := range validOperators {
		assert.True(t, operator(valid).IsValid())
	}

	for _, invalid := range invalidOperators {
		assert.False(t, operator(invalid).IsValid())
	}
}

//nolint:paralleltest
func TestOperatorMultiVal(t *testing.T) {
	multiValOperators := []string{
		"all", "eqa", "nin", "in", "rein",
		"icoin",
	}
	nonMultiValOperators := []string{"eq", "exists", "gt", "lte", "ne"}

	for _, op := range multiValOperators {
		assert.True(t, operator(op).IsValid())
		assert.True(t, operator(op).IsMultiVal())
	}

	for _, op := range nonMultiValOperators {
		assert.True(t, operator(op).IsValid())
		assert.False(t, operator(op).IsMultiVal())
	}
}

//nolint:paralleltest
func TestOperatorRegex(t *testing.T) {
	regexOps := []string{"re", "ire", "rein", "irein"}
	nonRegexOps := []string{"co", "sw", "all", "eqa", "gte", "icoin"}

	for _, op := range regexOps {
		assert.True(t, operator(op).IsValid(),
			"operator: %s", op)
		assert.True(t, operator(op).IsRegexOperator(),
			"operator: %s", op)
	}

	for _, op := range nonRegexOps {
		assert.True(t, operator(op).IsValid(),
			"operator: %s", op)
		assert.False(t, operator(op).IsRegexOperator(),
			"operator: %s", op)
	}
}

//nolint:paralleltest
func TestOperatorSW(t *testing.T) {
	swOps := []string{"sw", "isw", "swin", "iswin"}
	nonSWOps := []string{"eq", "ne", "nin", "in", "co", "re"}

	for _, op := range swOps {
		assert.True(t, operator(op).IsValid())
		assert.True(t, operator(op).IsStartsWithOperator())
	}

	for _, op := range nonSWOps {
		assert.True(t, operator(op).IsValid())
		assert.False(t, operator(op).IsStartsWithOperator())
	}
}

//nolint:paralleltest
func TestOperatorContains(t *testing.T) {
	coOps := []string{"co", "ico", "coin", "icoin"}
	nonCoOps := []string{"all", "rein", "iswin", "nin", "lt", "eqa"}

	for _, op := range coOps {
		assert.True(t, operator(op).IsValid())
		assert.True(t, operator(op).IsContainsOperator())
	}

	for _, op := range nonCoOps {
		assert.True(t, operator(op).IsValid())
		assert.False(t, operator(op).IsContainsOperator())
	}
}

//nolint:paralleltest
func TestOperatorIgnoreCase(t *testing.T) {
	icOps := []string{"ire", "irein", "ico", "icoin", "isw", "iswin"}
	nonICOps := []string{"rein", "co", "all", "eqa", "nin", "in"}

	for _, op := range icOps {
		assert.True(t, operator(op).IsValid())
		assert.True(t, operator(op).IsIgnoreCaseOperator())
		assert.Equal(t, "i", operator(op).RegexOpts())
	}

	for _, op := range nonICOps {
		assert.True(t, operator(op).IsValid())
		assert.False(t, operator(op).IsIgnoreCaseOperator())
		assert.Equal(t, "", operator(op).RegexOpts())
	}
}

//nolint:paralleltest
func TestOperatorMongoOp(t *testing.T) {
	ops := map[string]string{
		"all":   "$all",
		"eq":    "$eq",
		"re":    "$eq",
		"iswin": "$in",
		"in":    "$in",
		"eqa":   "$eq",
		"nin":   "$nin",
		"gte":   "$gte",
		"lt":    "$lt",
		"co":    "$eq",
		"icoin": "$in",
		"isw":   "$eq",
	}

	for op, mOp := range ops {
		assert.True(t, operator(op).IsValid(),
			"invalid operator: %s", op)
		assert.Equal(t, operator(op).IsIgnoreCaseOperator(),
			operator(op).RegexOpts() == "i",
			"for operator: %s", op)
		assert.Equal(t, mOp, operator(op).MongoOperator(),
			"for operator: %s", op)
	}
}
