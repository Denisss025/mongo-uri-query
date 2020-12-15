# URI to MongoDB Query

[![Go Reference](https://pkg.go.dev/badge/github.com/Denisss025/mongo-uri-query.svg)](https://pkg.go.dev/github.com/Denisss025/mongo-uri-query)
![Build Status](https://travis-ci.org/Denisss025/mongo-uri-query.svg?branch=master)
[![Go Report](https://goreportcard.com/badge/Denisss025/mongo-uri-query)](https://goreportcard.com/report/Denisss025/mongo-uri-query)
[![Maintainability](https://api.codeclimate.com/v1/badges/5dcb97ef85e043fa0208/maintainability)](https://codeclimate.com/github/Denisss025/mongo-uri-query/maintainability)
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

```Go
package example

import (
    "net/http"

    query "github.com/Denisss025/mongo-uri-query"

    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type MongoPrimitives struct {}

func (p MongoPrimitives) ObjectID(val string) (interface{}, error) {
    return primitive.ObjectIDFromHex(val)
}

func (p MongoPrimitives) RegEx(val, opts string) (interface{}, error) {
    return primitive.Regex{Pattern: val, Options: opts}, nil
}

type RequestHandler struct {
    parser query.Parser
}

func NewHandler() *RequestHandler {
    return &RequestHandler{parser: query.Parser{
        Converter: query.NewDefaultConverter(MongoPrimitives{}),
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
}
```

The `RegEx()` function is used with `re`, `co` and `sw` operators.

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

