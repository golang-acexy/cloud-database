package mongo

import (
	"github.com/golang-acexy/starter-mongo/mongostarter"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// NewCondQuery 创建实体条件查询。
func NewCondQuery[T mongostarter.Model](condition T) mongostarter.CondQuery[T] {
	return mongostarter.NewCondQuery(condition)
}

// NewBSONQuery 创建 BSON 条件查询。
func NewBSONQuery(condition bson.M) mongostarter.BSONQuery {
	return mongostarter.NewBSONQuery(condition)
}

// NewPageQuery 创建实体条件分页查询。
func NewPageQuery[T mongostarter.Model](condition T, number, size int) mongostarter.PageQuery[T] {
	return mongostarter.NewPageQuery(condition, number, size)
}

// NewBSONPageQuery 创建 BSON 条件分页查询。
func NewBSONPageQuery(condition bson.M, number, size int) mongostarter.BSONPageQuery {
	return mongostarter.NewBSONPageQuery(condition, number, size)
}

// NewFilterPageQuery 创建原生 Filter 分页查询。
func NewFilterPageQuery(filter any, number, size int) mongostarter.FilterPageQuery {
	return mongostarter.NewFilterPageQuery(filter, number, size)
}
