package query

// Field is a structure that holds field specification.
type Field struct {
	// Converter defines a type of the field.
	Converter Converter
	// Required defines if the field is required.
	Required bool
}

// Fields is a map with fields specifications.
type Fields map[string]Field

// HasField check if a field with a given name is present in the
// fields specifications.
func (f Fields) HasField(name string) (ok bool) {
	_, ok = f[name]

	return
}

// Converter returns a specified converter for a field with a given name.
func (f Fields) Converter(name string) (converter Converter, ok bool) {
	field, ok := f[name]
	if ok {
		converter = field.Converter
	}

	return
}

// IsRequired returns true it a field with a given name is specified and
// is required.
func (f Fields) IsRequired(name string) (ok bool) {
	field, ok := f[name]
	if ok {
		ok = field.Required
	}

	return
}
