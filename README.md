# URI to MongoDB Query

[![Go Reference](https://pkg.go.dev/badge/github.com/Denisss025/mongo-uri-query.svg)](https://pkg.go.dev/github.com/Denisss025/mongo-uri-query)
[![Build Status](https://travis-ci.org/Denisss025/mongo-uri-query.svg?branch=master)
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
    return primitive.RegEx{Pattern: val, Options: opts}, nil
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
    q, err := h.Parse(r.URI.Query())
    if err != nil { ... }

    cursor, err := coll.Find(r.Context(), q.Filter, &options.FindOptions{
        Limit: &q.Limit,
        Skip:  &q.Skip,
        Sort:  q.Sort,
    })

    ...
}
```

## License

The URI to MongoDB Query library is licensed under the
[MIT License](https://github.com/Denisss025/mongo-uri-query/blob/master/LICENCE).

