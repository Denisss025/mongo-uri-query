package query

type Field struct {
	Converter Converter
	Required  bool
}

type Fields map[string]Field

func (f Fields) HasField(name string) (ok bool) {
	_, ok = f[name]

	return
}

func (f Fields) Converter(name string) (converter Converter, ok bool) {
	field, ok := f[name]
	if ok {
		converter = field.Converter
	}

	return
}

func (f Fields) IsRequired(name string) (ok bool) {
	field, ok := f[name]
	if ok {
		ok = field.Required
	}

	return
}
