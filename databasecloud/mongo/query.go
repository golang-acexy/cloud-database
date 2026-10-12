package mongo

import (
	"github.com/golang-acexy/starter-mongo/mongostarter"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// NewCondQuery delegates construction to the starter; validation is deferred and referenced inputs are not deeply copied.
func NewCondQuery[T mongostarter.Model](condition T) mongostarter.CondQuery[T] {
	return mongostarter.NewCondQuery(condition)
}

// NewBSONQuery delegates construction to the starter; validation is deferred and referenced inputs are not deeply copied.
func NewBSONQuery(condition bson.M) mongostarter.BSONQuery {
	return mongostarter.NewBSONQuery(condition)
}

// NewPageQuery delegates construction to the starter; validation is deferred and referenced inputs are not deeply copied.
func NewPageQuery[T mongostarter.Model](condition T, number, size int) mongostarter.PageQuery[T] {
	return mongostarter.NewPageQuery(condition, number, size)
}

// NewBSONPageQuery delegates construction to the starter; validation is deferred and referenced inputs are not deeply copied.
func NewBSONPageQuery(condition bson.M, number, size int) mongostarter.BSONPageQuery {
	return mongostarter.NewBSONPageQuery(condition, number, size)
}

// NewFilterPageQuery delegates construction to the starter; validation is deferred and referenced inputs are not deeply copied.
func NewFilterPageQuery(filter any, number, size int) mongostarter.FilterPageQuery {
	return mongostarter.NewFilterPageQuery(filter, number, size)
}
