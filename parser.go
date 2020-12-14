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

type Parser struct {
	Converter      *TypeConverter
	Fields         Fields
	ValidateFields bool

	initRegescape sync.Once
	rxRegEscape   *strings.Replacer
}

type operatorsMap = map[operator][]string

type fieldsMap = map[string]map[operator][]string

func normailzeFields(fields fieldsMap) (normalized fieldsMap) {
	normalized = make(fieldsMap)

	for field, ops := range fields {
		normalized[field] = make(operatorsMap)

		for op, arr := range ops {
			cop := op.CommonOperator()
			if op == cop {
				continue
			}

			common, exists := ops[cop]
			if exists {
				normalized[field][cop] = append(common, arr...)
			} else {
				normalized[field][op] = arr
			}
		}
	}

	for field, ops := range fields {
		ff := normalized[field]

		for op, arr := range ops {
			commonOp := op.CommonOperator()
			commonArr, hasOp := ff[commonOp]

			if hasOp && len(commonArr) > 1 {
				continue
			}

			if len(arr) == 1 && op.NeedSplitString() {
				arr = strings.Split(arr[0], arrayDelimiter)
			}

			if len(arr) == 1 {
				delete(ff, op)
				op = op.SingleValueOperator()
			}

			ff[op] = arr
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

	if len(v) == 0 {
		return nil, nil
	}

	if op.IsMultiVal() {
		return mapValues(v, c)
	}

	if len(v) > 1 {
		return nil, ErrTooManyValues
	}

	return c.Convert(v[0])
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

func (p *Parser) convert(field string, op operator, v []string) (
	value interface{}, err error) {
	conv, hasField := p.Fields.Converter(field)
	if !hasField {
		if p.ValidateFields {
			return nil,
				fmt.Errorf("%w: %s", ErrNoFieldSpec, field)
		}

		conv = p.Converter

		if op == operatorExists {
			conv = p.Converter.Bool
		}
	}

	switch {
	case op.IsRegex():
		conv = ConvertFunc(
			func(val string) (rx interface{}, err error) {
				return p.Converter.Primitives.RegEx(
					val, op.RegexOpts())
			})
	case op.IsContains():
		conv = ConvertFunc(
			func(val string) (rx interface{}, err error) {
				return p.Converter.Primitives.RegEx(
					p.regEscape(val), op.RegexOpts())
			})
	case op.IsStartsWith():
		conv = ConvertFunc(
			func(val string) (rx interface{}, err error) {
				return p.Converter.Primitives.RegEx(
					"^"+p.regEscape(val),
					op.RegexOpts())
			})
	}

	value, err = convertArray(v, op, conv)
	if err != nil {
		return nil, fmt.Errorf("convert %s: %w", field, err)
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

func (p *Parser) Parse(params url.Values) (
	filter Query, err error) {
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

	for _, sort := range sortFields {
		sortField := filter.AddSort(sort)

		if p.ValidateFields && !p.Fields.HasField(sortField) {
			errs = multierror.Append(errs, fmt.Errorf(
				"%w: %s", ErrNoSortField, sortField))
		}
	}

	if errs != nil {
		err = fmt.Errorf("parse: %w", errs.ErrorOrNil())
	}

	return filter, err
}
