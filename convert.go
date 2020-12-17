package query

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Converter is an interface that converts strings to known types.
type Converter interface {
	// Convert checks a string val and converts it when possible to some
	// type.
	Convert(val string) (i interface{}, err error)
}

// ConvertFunc is a function to check and convert a string val.
type ConvertFunc func(val string) (i interface{}, err error)

// static assertion: ConvertFunc must implement Converter interface.
var _ = Converter(ConvertFunc(nil))

// Convert calls ConvertFunc itself.
func (c ConvertFunc) Convert(val string) (i interface{}, err error) {
	return c(val)
}

// Primitives gives access to the RegEx and ObjectID convertors.
type Primitives interface {
	// RegEx converts pattern and options to bson.Regex.
	RegEx(pattern, options string) (rx interface{}, err error)
	// ObjectID converts val to bson.ObjectID.
	ObjectID(val string) (oid interface{}, err error)
	// DocElem converts key and val o bson.DocElem, which is a bson.D element.
	DocElem(key string, val interface{}) (d interface{}, err error)
}

// String returns a string val.
func String() (convert ConvertFunc) {
	return func(val string) (i interface{}, err error) {
		return val, nil
	}
}

// Int tries to convert a val string to an int value.
func Int() (convert ConvertFunc) {
	return func(val string) (i interface{}, err error) {
		return strconv.ParseInt(val, 10, 64)
	}
}

// Double tries to convert a val string to a float64 value.
func Double() (convert ConvertFunc) {
	return func(val string) (i interface{}, err error) {
		return strconv.ParseFloat(val, 64)
	}
}

// Bool tries to convert a val string to a boolean value.
func Bool() (convert ConvertFunc) {
	return func(val string) (i interface{}, err error) {
		switch strings.ToLower(val) {
		case "true", "yes":
			return true, nil
		case "false", "no":
			return false, nil
		}

		return nil, ErrNoMatch
	}
}

// ObjectID checks if a string can be converted to an ObjectID value and
// converts it.
func ObjectID(primitive Primitives) (convert ConvertFunc) {
	objectIDConvert := primitive.ObjectID
	rx := regexp.MustCompile("^[0-9a-fA-F]{12}")

	return func(val string) (i interface{}, err error) {
		if !rx.MatchString(val) {
			return nil, ErrNoMatch
		}

		return objectIDConvert(val)
	}
}

// Date checks if a string matches some of the known patterns and tries to
// convert it to time.Time.
func Date() (convert ConvertFunc) {
	const (
		dateFmt            = "2006-01-02"
		utcTimeFmt         = "2006-01-02T15:04:05Z"
		utcTimeWithNsecFmt = "2006-01-02T15:04:05.999Z"
		timeFmt            = utcTimeFmt + "-0700"
		timeWithNsecFmt    = utcTimeWithNsecFmt + "-0700"
	)

	formats := []string{
		dateFmt, utcTimeFmt, timeFmt,
		utcTimeWithNsecFmt, timeWithNsecFmt,
	}

	return func(val string) (i interface{}, err error) {
		for _, layout := range formats {
			if i, err = time.Parse(layout, val); err == nil {
				return
			}
		}

		return nil, ErrNoMatch
	}
}

// TypeConverter is a type that detects type and converts strings to that type.
type TypeConverter struct {
	// Bool is a boolean type converter
	Bool ConvertFunc
	// Primitives gives access to mongodb-driver primitives.
	Primitives Primitives
	// Funcs checks and converts strings to known types.
	Funcs []ConvertFunc
}

// static assertion: *TypeConverter must implement Converter interface.
var _ = Converter((*TypeConverter)(nil))

// NewConverter creates an instance of the TypeConverter.
func NewConverter(boolConvert ConvertFunc, p Primitives,
	convert ...ConvertFunc) (c *TypeConverter) {
	c = &TypeConverter{
		Bool:       boolConvert,
		Primitives: p,
		Funcs:      make([]ConvertFunc, 0, len(convert)+1),
	}

	if p != nil {
		c.Funcs = append(c.Funcs, ObjectID(p))
	}

	for _, cx := range convert {
		if cx != nil {
			c.Funcs = append(c.Funcs, cx)
		}
	}

	return
}

// NewDefaultConverter creates a default TypeConverter instance.
func NewDefaultConverter(p Primitives) (c *TypeConverter) {
	return NewConverter(Bool(), p,
		Int(), Double(), Date(), String())
}

// Convert checks string value for patterns and converts it to matched types.
func (c TypeConverter) Convert(val string) (i interface{}, err error) {
	if i, err = c.Bool(val); err == nil {
		return i, nil
	}

	for _, convert := range c.Funcs {
		if i, err = convert(val); err == nil {
			return i, nil
		}
	}

	return nil, ErrNoMatch
}
