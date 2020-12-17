package query

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
)

const (
	// Delimiters.
	delimiter      = "__"
	arrayDelimiter = ","

	// Params.
	limitParam = "limit"
	skipParam  = "skip"
	sortParam  = "sort"

	// Sort constraints.
	sortAscPrefix  = "+"
	sortDescPrefix = "-"
	sortAsc        = 1
	sortDesc       = -1
)

// Parser is a structure that parses url queries.
type Parser struct {
	// Converter is a TypeConverter that converts unspecified fields.
	Converter *TypeConverter
	// Fields is a fields specification.
	Fields Fields
	// ValidateFields enables or disables field specification validator.
	// When true, the parser will return ErrNoFieldSpec for every
	// unspecified field in url query.
	ValidateFields bool

	initRegescape sync.Once
	rxRegEscape   *strings.Replacer
}

type operatorsMap = map[operator][]string

type fieldsMap = map[string]map[operator][]string

func normailzeFields(fields fieldsMap) (normalized fieldsMap) {
	normalized = make(fieldsMap)

	for field, ops := range fields {
		ff := make(operatorsMap)

		for op, arr := range ops {
			cop := op.CommonOperator()

			if len(arr) == 1 && op.NeedSplitString() {
				arr = strings.Split(arr[0], arrayDelimiter)
			}

			ff[cop] = append(ff[cop], arr...)
		}

		for op, arr := range ff {
			if len(arr) != 1 || !op.IsMultiVal() {
				continue
			}

			ff[op.SingleValueOperator()] = arr
			delete(ff, op)
		}

		normalized[field] = ff
	}

	return normalized
}

func extractFields(query url.Values) (fields fieldsMap) {
	fields = make(fieldsMap)

	for k, v := range query {
		if strings.HasPrefix(k, delimiter) {
			continue
		}

		field, op := parseOperator(k)

		// convert map[like][field] to struct.like.field
		field = strings.ReplaceAll(
			strings.ReplaceAll(field, "[", "."),
			"]", "")

		f, ok := fields[field]
		if !ok {
			f = make(map[operator][]string)
		}

		if arr, hasOperator := f[op]; hasOperator {
			f[op] = append(arr, v...)
		} else {
			f[op] = v
		}

		fields[field] = f
	}

	return normailzeFields(fields)
}

func mapValues(values []string, c Converter) (i []interface{}, err error) {
	i = make([]interface{}, len(values))

	for n, val := range values {
		if i[n], err = c.Convert(val); err != nil {
			return nil, fmt.Errorf("map: %w", err)
		}
	}

	return i, nil
}

func convertArray(v []string, op operator, c Converter) (
	value interface{}, err error) {
	if c == nil {
		return nil, ErrNoConverter
	}

	if op.IsMultiVal() {
		return mapValues(v, c)
	}

	if len(v) > 1 {
		err = ErrTooManyValues
	} else if len(v) == 1 {
		value, err = c.Convert(v[0])
	}

	return value, err
}

func parseIntParam(params url.Values, name string) (val int64, err error) {
	str := params.Get(delimiter + name)
	if len(str) != 0 {
		val, err = strconv.ParseInt(str, 10, 31)
		if err != nil {
			err = fmt.Errorf("%s parameter: %w", name, err)
		}
	}

	return
}

func (p *Parser) regEscape(val string) (escaped string) {
	p.initRegescape.Do(
		func() {
			const (
				replaceChars = ".*?+^$[](){}|-"
				escapeSymbol = "\\"

				mul2 = 2
			)

			oldNew := make([]string, 0, len(replaceChars)*mul2)

			for _, c := range replaceChars {
				oldNew = append(oldNew, string(c),
					escapeSymbol+string(c))
			}

			p.rxRegEscape = strings.NewReplacer(oldNew...)
		},
	)

	return p.rxRegEscape.Replace(val)
}

func (p *Parser) regex(reOptions string, translate func(string) string) (
	conv ConvertFunc) {
	if p.Converter == nil || p.Converter.Primitives == nil {
		return nil
	}

	return func(val string) (rx interface{}, err error) {
		return p.Converter.Primitives.RegEx(
			translate(val), reOptions)
	}
}

func nop() (translate func(string) string) {
	return func(a string) string { return a }
}

func sw(f func(string) string) (translate func(string) string) {
	return func(a string) string { return "^" + f(a) }
}

func (p *Parser) convert(field string, op operator, v []string) (
	value interface{}, err error) {
	const errMsg = "convert: %w: %v"

	if !op.IsValid() {
		return nil, fmt.Errorf(errMsg, ErrUnknownOperator, op)
	}

	conv, hasField := p.Fields.Converter(field)
	if !hasField {
		if p.ValidateFields {
			return nil,
				fmt.Errorf(errMsg, ErrNoFieldSpec, field)
		}

		conv = p.Converter

		if op == operatorExists {
			conv = p.Converter.Bool
		}
	}

	switch {
	case op.IsRegex():
		conv = p.regex(op.RegexOpts(), nop())
	case op.IsContains():
		conv = p.regex(op.RegexOpts(), p.regEscape)
	case op.IsStartsWith():
		conv = p.regex(op.RegexOpts(), sw(p.regEscape))
	}

	value, err = convertArray(v, op, conv)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err, field)
	}

	return value, err
}

func getSortFields(params url.Values) (sortFields []string) {
	sortParams, hasSortParam := params[delimiter+sortParam]

	if !hasSortParam {
		return
	}

	sortFields = make([]string, 0, len(sortParams))

	for _, param := range sortParams {
		split := strings.Split(param, arrayDelimiter)
		sortFields = append(sortFields, split...)
	}

	return
}

func (p *Parser) parseFilter(query url.Values) (
	filter Query, errs *multierror.Error) {
	fields := extractFields(query)

	for field, operators := range fields {
		for op, values := range operators {
			value, parseErr := p.convert(field, op, values)
			if parseErr != nil {
				errs = multierror.Append(errs,
					fmt.Errorf("filter: %w: %s[%v]",
						parseErr, field, op))
			} else {
				filter.AddFilter(field, op, value)
			}
		}
	}

	for fieldName, field := range p.Fields {
		if field.Required {
			if _, hasField := filter.Filter[fieldName]; !hasField {
				errs = multierror.Append(errs,
					fmt.Errorf("filter: %w: %s",
						ErrMissingField, fieldName))
			}
		}
	}

	return filter, errs
}

// Parse parses a given url query.
func (p *Parser) Parse(params url.Values) (filter Query, err error) {
	var errs *multierror.Error

	filter, errs = p.parseFilter(params)

	filter.Limit, err = parseIntParam(params, limitParam)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	filter.Skip, err = parseIntParam(params, skipParam)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	sortFields := getSortFields(params)

	if len(sortFields) > 0 &&
		(p.Converter == nil || p.Converter.Primitives == nil) {
		errs = multierror.Append(errs, fmt.Errorf("no primitives: %w",
			ErrNoSortField))
	} else {
		for _, sort := range sortFields {
			sortField, sortErr := filter.AddSort(sort,
				p.Converter.Primitives.DocElem)

			if sortErr != nil {
				errs = multierror.Append(errs, sortErr)
			} else if p.ValidateFields && !p.Fields.HasField(sortField) {
				errs = multierror.Append(errs, fmt.Errorf(
					"%w: %s", ErrNoSortField, sortField))
			}
		}
	}

	if errs != nil {
		err = fmt.Errorf("parse: %w", errs.ErrorOrNil())
	}

	return filter, err
}
