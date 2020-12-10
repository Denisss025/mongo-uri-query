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

func mapValues(values []string, c Converter) (i []interface{}, err error) {
	i = make([]interface{}, len(values))

	for n, val := range values {
		if i[n], err = c.Convert(val); err != nil {
			return nil, fmt.Errorf("map: %w", err)
		}
	}

	return i, nil
}

func parse(v []string, op operator, c Converter) (
	value interface{}, err error) {
	if c == nil {
		return nil, ErrNoConverter
	}

	if len(v) == 0 {
		return nil, nil
	}

	if op.IsMultiVal() {
		if len(v) == 1 {
			v = strings.Split(v[0], arrayDelimiter)
		}

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

func (p *Parser) getValue(param string, v []string) (
	field string, op operator, value interface{}, err error) {
	field, op = param, operatorEquals

	if pos := strings.Index(param, delimiter); pos > 0 {
		field, op = param[:pos], operator(param[pos+len(delimiter):])
	}

	if !op.IsValid() {
		return "", "", nil,
			fmt.Errorf("%w: %s", ErrUnknownOperator, string(op))
	}

	conv, hasField := p.Fields.Converter(field)
	if !hasField {
		conv = p.Converter

		if op == operatorExists {
			conv = p.Converter.Bool
		}

		if p.ValidateFields {
			return "", "", nil,
				fmt.Errorf("%w: %s", ErrNoFieldSpec, field)
		}
	}

	switch {
	case op.IsRegexOperator():
		conv = ConvertFunc(
			func(val string) (rx interface{}, err error) {
				return p.Converter.Primitives.RegEx(
					val, op.RegexOpts())
			})
	case op.IsContainsOperator():
		conv = ConvertFunc(
			func(val string) (rx interface{}, err error) {
				return p.Converter.Primitives.RegEx(
					p.regEscape(val), op.RegexOpts())
			})
	case op.IsStartsWithOperator():
		conv = ConvertFunc(
			func(val string) (rx interface{}, err error) {
				return p.Converter.Primitives.RegEx(
					"^"+p.regEscape(val),
					op.RegexOpts())
			})
	}

	value, err = parse(v, op, conv)
	if err != nil {
		err = fmt.Errorf("convert %s: %w", field, err)
		field, op, value = "", "", nil
	}

	return field, op, value, err
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

func (p *Parser) parseFilter(params url.Values) (
	filter Query, errs *multierror.Error) {
	for k, v := range params {
		if strings.HasPrefix(k, delimiter) {
			continue
		}

		fieldName, op, value, parseErr := p.getValue(k, v)
		if parseErr != nil {
			errs = multierror.Append(errs,
				fmt.Errorf("filter: %w: %s", parseErr, k))
		} else {
			filter.AddFilter(fieldName, op, value)
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
