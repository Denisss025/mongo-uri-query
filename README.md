# URI to MongoDB Query

[![Go Reference](https://pkg.go.dev/badge/github.com/Denisss025/mongo-uri-query.svg)](https://pkg.go.dev/github.com/Denisss025/mongo-uri-query)
[![Build Status](https://travis-ci.org/Denisss025/mongo-uri-query.svg?branch=master)](https://travis-ci.org/Denisss025/mongo-uri-query)
[![Go Report](https://goreportcard.com/badge/Denisss025/mongo-uri-query)](https://goreportcard.com/report/Denisss025/mongo-uri-query)
[![Maintainability](https://api.codeclimate.com/v1/badges/be2fde656b7fbc1e5795/maintainability)](https://codeclimate.com/github/Denisss025/mongo-uri-query/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/5dcb97ef85e043fa0208/test_coverage)](https://codeclimate.com/github/Denisss025/mongo-uri-query/test_coverage)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/Denisss025/mongo-uri-query/blob/master/LICENSE)

The URI to MongoDB query conversion library for Go
programming language    .

The library is deeply inspired by the
[query-params-mongo](https://github.com/vasansr/query-params-mongo)
NodeJS library.

## Installation

The recommended way to get started using the URI to MongoDB Query
library is by using go modules to install the dependency in your
project.
This can be done by importing packages from
`github.com/Denisss025/mongo-uri-query` and having the build step
to install the dependency.

Another way is to get the library by explicitly running
```SH
go get github.com/Denisss025/mongo-uri-query
```

## Usage
## Example

A request of the form:
```URL
/employees?name=John&age__lte=45&category__in=A,B&__limit=10&__sort=-age
```
Is translated to:
```Go
Query{
    Filter: map[string]interface{}{
        "name":     "John",
        "age":      map[string]interface{}{"$lte": 45},
        "category": map[string]interface{}{$in: []interface{}{"A", "B"}},
    },
    Sort:  map[string]int{ "age": -1 },
    Limit: 10,
    Skip:  0,
}
```

Now the filter, sort, limit and skip can be alltogether used with
[mongo-go-driver](https://github.com/mongodb/mongo-go-driver) or other
MongoDB library, i.e. with an old
[go-mgo/mgo](https://github.com/go-mgo/mgo) library.

[go-mgo/mgo](https://github.com/go-mgo/mgo) package.

```Go
package example

import (
	"errors"
	"net/http"

	query "github.com/Denisss025/mongo-uri-query"

	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2/mgo"
)

type primitives struct{}

func (p primitives) ObjectID(val string) (oid interface{}, err error) {
	if !bson.IsObjectIdHex(val) {
		return nil, errors.New("not an ObjectIdHex")
    }
    
    return bson.ObjectIdHex(val), nil
}

func (p primitives) RegEx(p, o string) (re interface{}, err error) {
	return bson.RegEx{Pattern: p, Options: o}
}

func (p primitives) DocElem(k string, v interface{}) (
	kv interface{}, err error) {
	return bson.DocElem{Name: k, Value: v}, nil
}

type RequestHandler struct {
	parser query.Parser
}

func NewHandler() *RequestHandler {
	return &RequestHandler{parser: query.Parser{
		Converter: query.NewDefaultConverter(primitives{}),
	}}
}

func (h *RequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var coll *mgo.Collection
	...
	q, err := h.parser.Parse(r.URL.Query())
	if err != nil { ... }

	bsonSort, _ := q.Sort.([]bson.DocElem)
	
	sortFields := make([]string, len(bsonSort))
	for i, s := range bsonSort {
		sf := "-" + s.Key
		sortFields[i] = sf[(s.(int)+1)/2:]
	}

	cursor := coll.Find(q.Filter).
		Limit(int(q.Limit)).
		Skip(int(q.Skip)).
		Sort(sortFields...)

	...
}
```

MongoDB [driver](https://github.com/mongodb/mongo-go-driver)

```Go
package example

import (
	"errors"
	"net/http"

	query "github.com/Denisss025/mongo-uri-query"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type primitives struct{}

func (p primitives) ObjectID(val string) (oid interface{}, err error) {
	return primitive.ObjectIDFromHex(val)
}

func (p primitives) RegEx(p, o string) (re interface{}, err error) {
    return primitive.Regex{Pattern: p, Options: o}, nil
}

func (p primitives) DocElem(k string, v interface{}) (
	kv interface{}, err error) {
	return primitive.E{Key: k, Value: v}, nil
}

type RequestHandler struct {
	parser query.Parser
}

func NewHandler() *RequestHandler {
	return &RequestHandler{parser: query.Parser{
		Converter: query.NewDefaultConverter(primitives{}),
	}}
}

func (h *RequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var coll *mongo.Collection
	...
	q, err := h.parser.Parse(r.URL.Query())
	if err != nil { ... }

	cursor, err := coll.Find(r.Context(), q.Filter, &options.FindOptions{
		Limit: &q.Limit,
		Skip:  &q.Skip,
		Sort:  q.Sort,
	})

	...
}
```

## API Reference

### Create a parser

The parser is a structure that has a `Parse()` function that can process the request query.
You can create as many instances as you like, but typically your app would need only one.
The behaviour of the parser is controlled with `TypeConverter`, `Fields` and `ValidateFields`
member fields.

```Go
parser := query.Parser{TypeConverter: ..., Fields: ..., ValidateFields: ...}
```

#### Fields

* `TypeConverter` is a structure that is able to automatically detect value type
  and convert string to it.

* `Fields` is a map that holds fields specifications:

  * `Required`: the parser checks all the required fields to be given in a query.
 
  * `Converter` is a custom type converter for a given field.
 
* `ValidateFields`: when `true` the parser checks every given query param to be present in
   the `Fields` map.
   
The `TypeConverter` can be created either with `NewConverter()` or with `NewDefaultConverter()`
functions. The `NewDefaultConverter()` function creates a `TypeConverter` that automatically
detects such types as `ObjectID` (`[0-9a-f]{12}`), `int64`, `float64`, `bool` (`true|yes|false|no`) and `time.Time` (i.e. `2006-01-02T15:04:05Z0700`).

The `TypeConverter` also has a `Primitives` field which is used to convert strings to `ObjectID` and `RegEx`.
`Primitives` is an interface with two functions:

```Go
type Primitives interface {
    RegEx(val, opts string) (interface{}, error)
    ObjectID(val string) (interface{}, error)
    DocElem(key string, val interface{}) (interface{}, error)
}
```

The `RegEx()` function is used with `re`, `co` and `sw` operators.

The `DocElem()` function is used with `__sort` directive. It allows to
define sort order for `Sort()` function or for `FindOptions.Sort` field.

### Parse a query

```Go
q, err := parser.Parse(r.URL.Query())
```

`r` is a pointer to an `http.Request{}`, `q` is a `Query{}`.

The `Query{}` structure has `Filter`, `Sort`, `Limit` and `Skip` fields.

* `Filter` is a mongo-db find filter.

* `Sort` is a mongo-db sort specification.

* `Limit` is a value for `Cursor.Limit()` to limit the number of documents in the query result.

* `Skip` is a value for `Cursor.Skip()` to skip the number of documents in the query result.


## License

The URI to MongoDB Query library is licensed under the
[MIT License](https://github.com/Denisss025/mongo-uri-query/blob/master/LICENCE).

