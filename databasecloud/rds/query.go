package rds

import (
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"gorm.io/gorm"
)

// NewCondQuery delegates construction to the starter; validation is deferred and referenced inputs are not deeply copied.
func NewCondQuery[T gormstarter.Model](condition T) gormstarter.CondQuery[T] {
	return gormstarter.NewCondQuery(condition)
}

// NewMapQuery delegates construction to the starter; validation is deferred and referenced inputs are not deeply copied.
func NewMapQuery(condition map[string]any) gormstarter.MapQuery {
	return gormstarter.NewMapQuery(condition)
}

// NewWhereQuery delegates construction to the starter; validation is deferred and referenced inputs are not deeply copied.
func NewWhereQuery(rawWhereSQL string, args ...any) gormstarter.WhereQuery {
	return gormstarter.NewWhereQuery(rawWhereSQL, args...)
}

// NewPageQuery delegates construction to the starter; validation is deferred and referenced inputs are not deeply copied.
func NewPageQuery[T gormstarter.Model](condition T, number, size int) gormstarter.PageQuery[T] {
	return gormstarter.NewPageQuery(condition, number, size)
}

// NewMapPageQuery delegates construction to the starter; validation is deferred and referenced inputs are not deeply copied.
func NewMapPageQuery(condition map[string]any, number, size int) gormstarter.MapPageQuery {
	return gormstarter.NewMapPageQuery(condition, number, size)
}

// NewWherePageQuery delegates construction to the starter; validation is deferred and referenced inputs are not deeply copied.
func NewWherePageQuery(rawWhereSQL string, number, size int, args ...any) gormstarter.WherePageQuery {
	return gormstarter.NewWherePageQuery(rawWhereSQL, number, size, args...)
}

// NewGormPageQuery stores callbacks and metadata; callbacks must supply filtering and limit/offset.
func NewGormPageQuery(number, size int, countRawDB, pageRawDB func(*gorm.DB)) GormPageQuery {
	return GormPageQuery{CountRawDB: countRawDB, PageRawDB: pageRawDB, PageOptions: PageOptions{Number: number, Size: size}}
}
