package rds

import (
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"gorm.io/gorm"
)

// NewCondQuery 创建实体条件查询。
func NewCondQuery[T gormstarter.Model](condition T) gormstarter.CondQuery[T] {
	return gormstarter.NewCondQuery(condition)
}

// NewMapQuery 创建 Map 条件查询。
func NewMapQuery(condition map[string]any) gormstarter.MapQuery {
	return gormstarter.NewMapQuery(condition)
}

// NewWhereQuery 创建原始 SQL 条件查询，args 按占位符顺序传入。
func NewWhereQuery(rawWhereSQL string, args ...any) gormstarter.WhereQuery {
	return gormstarter.NewWhereQuery(rawWhereSQL, args...)
}

// NewPageQuery 创建实体条件分页查询。
func NewPageQuery[T gormstarter.Model](condition T, number, size int) gormstarter.PageQuery[T] {
	return gormstarter.NewPageQuery(condition, number, size)
}

// NewMapPageQuery 创建 Map 条件分页查询。
func NewMapPageQuery(condition map[string]any, number, size int) gormstarter.MapPageQuery {
	return gormstarter.NewMapPageQuery(condition, number, size)
}

// NewWherePageQuery 创建原始 SQL 条件分页查询，args 按占位符顺序传入。
func NewWherePageQuery(rawWhereSQL string, number, size int, args ...any) gormstarter.WherePageQuery {
	return gormstarter.NewWherePageQuery(rawWhereSQL, number, size, args...)
}

// NewGormPageQuery 创建分别构建统计和分页数据的原始 GORM 查询。
func NewGormPageQuery(number, size int, countRawDB, pageRawDB func(*gorm.DB)) GormPageQuery {
	return GormPageQuery{CountRawDB: countRawDB, PageRawDB: pageRawDB, PageOptions: PageOptions{Number: number, Size: size}}
}
