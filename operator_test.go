package query

import (
	"strings"
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
		"icoin", "[]", "ire[]", "sw[]",
	}
	nonMultiValOperators := []string{"eq", "exists", "gt", "lte", "ne"}

	for _, op := range multiValOperators {
		assert.True(t, operator(op).IsValid())
		assert.True(t, operator(op).IsMultiVal(),
			"operator: %v", op)
		assert.Equal(t, strings.HasSuffix(op, "in") ||
			op == "all" || op == "eqa",
			operator(op).NeedSplitString(),
			"operator %s needs string splitting: %v",
			op, !operator(op).NeedSplitString())
	}

	for _, op := range nonMultiValOperators {
		assert.True(t, operator(op).IsValid())
		assert.False(t, operator(op).IsMultiVal())
		assert.False(t, operator(op).NeedSplitString())
	}
}

func TestOperatorIs(ts *testing.T) {
	ts.Parallel()

	ts.Run("should return true", func(t *testing.T) {
		t.Parallel()

		commonTypes := map[string][]string{
			"[]":    {"[]", "re[]", "ire[]", "co[]", "ico[], sw[]"},
			"in":    {"in", "[]", "re[]", "ire[]", "irein", "co[]"},
			"re":    {"re", "irein", "rein", "ire[]", "re[]", "ire"},
			"ire[]": {"ire[]"},
			"sw":    {"sw", "swin", "isw[]", "isw", "sw[]", "iswin"},
			"swin":  {"iswin", "isw[]", "swin", "isw[]"},
			"co":    {"co", "ico", "coin", "co[]", "ico[]", "icoin"},
			"ico":   {"ico", "icoin", "ico[]"},
			"all":   {"all", "all[]"},
			"eq":    {"eq", "eqa"},
		}

		for common, arr := range commonTypes {
			for _, op := range arr {
				assert.True(t,
					operator(op).Is(operator(common)),
					"%s must be also %s",
					op, common)
			}
		}
	})

	ts.Run("should return false", func(t *testing.T) {
		t.Parallel()

		commonTypes := map[string][]string{
			"re[]": {"in", "[]", "co[]", "ico[], sw[]"},
			"swin": {"in", "[]", "re[]", "ire[]", "irein", "co[]"},
			"in":   {"re", "ire", "ico", "isw", "eq"},
			"[]": {
				"sw", "swin", "iswin", "isw",
				"in", "eq", "all", "eqa",
			},
			"ico": {"co", "coin", "co[]", "eq", "in", "[]"},
			"all": {"in", "eq", "[]", "eqa"},
			"eqa": {"eq", "in", "[]"},
		}

		for common, arr := range commonTypes {
			for _, op := range arr {
				assert.False(t,
					operator(op).Is(operator(common)),
					"%s must not be %s",
					op, common)
			}
		}
	})
}

//nolint:paralleltest
func TestOperatorRegex(t *testing.T) {
	regexOps := []string{"re", "ire", "rein", "irein"}
	nonRegexOps := []string{"co", "sw", "all", "eqa", "gte", "icoin"}

	for _, op := range regexOps {
		assert.True(t, operator(op).IsValid(),
			"operator: %s", op)
		assert.True(t, operator(op).IsRegex(),
			"operator: %s", op)
	}

	for _, op := range nonRegexOps {
		assert.True(t, operator(op).IsValid(),
			"operator: %s", op)
		assert.False(t, operator(op).IsRegex(),
			"operator: %s", op)
	}
}

//nolint:paralleltest
func TestOperatorSW(t *testing.T) {
	swOps := []string{"sw", "isw", "swin", "iswin"}
	nonSWOps := []string{"eq", "ne", "nin", "in", "co", "re"}

	for _, op := range swOps {
		assert.True(t, operator(op).IsValid())
		assert.True(t, operator(op).IsStartsWith())
	}

	for _, op := range nonSWOps {
		assert.True(t, operator(op).IsValid())
		assert.False(t, operator(op).IsStartsWith())
	}
}

//nolint:paralleltest
func TestOperatorContains(t *testing.T) {
	coOps := []string{"co", "ico", "coin", "icoin"}
	nonCoOps := []string{"all", "rein", "iswin", "nin", "lt", "eqa"}

	for _, op := range coOps {
		assert.True(t, operator(op).IsValid())
		assert.True(t, operator(op).IsContains())
	}

	for _, op := range nonCoOps {
		assert.True(t, operator(op).IsValid())
		assert.False(t, operator(op).IsContains())
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

//nolint:paralleltest
func TestParseOperator(t *testing.T) {
	f, op := parseOperator("field[]")
	assert.Equal(t, "field", f)
	assert.Equal(t, operatorInArray, op)
	assert.True(t, op.IsValid())
	assert.True(t, op.IsMultiVal())
	assert.False(t, op.NeedSplitString())
	assert.True(t, operatorIn.NeedSplitString())
	assert.Equal(t, "$in", op.MongoOperator())
	assert.Equal(t, operatorEquals, op.SingleValueOperator())
	assert.True(t, op.Is(operatorIn))
	assert.False(t, op.IsRegex())
	assert.False(t, op.IsStartsWith())
	assert.False(t, op.IsContains())

	f, op = parseOperator("field__all[]")
	assert.Equal(t, "field", f)
	assert.Equal(t, operatorAllArray, op)
	assert.False(t, op.Is(operatorIn))
	assert.True(t, op.Is(operatorAllArray))
	assert.True(t, op.IsValid())
	assert.True(t, op.IsMultiVal())
	assert.False(t, op.NeedSplitString())
	assert.True(t, operatorAll.NeedSplitString())
	assert.Equal(t, "$all", op.MongoOperator())
	assert.Equal(t, operatorEquals, op.SingleValueOperator())
	assert.False(t, op.IsRegex())
	assert.False(t, op.IsStartsWith())
	assert.False(t, op.IsContains())

	f, op = parseOperator("field__ire[]")
	assert.Equal(t, "field", f)
	assert.Equal(t, operatorRegexInArrayIgnoreCase, op)
	assert.True(t, op.IsValid())
	assert.True(t, op.IsMultiVal())
	assert.False(t, op.NeedSplitString())
	assert.True(t, operatorRegexIn.NeedSplitString())
	assert.Equal(t, "$in", op.MongoOperator())
	assert.Equal(t, operatorRegexIgnoreCase, op.SingleValueOperator())
	assert.True(t, op.IsRegex())
	assert.False(t, op.IsStartsWith())
	assert.False(t, op.IsContains())
}
