package gorm

import (
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"gorm.io/gorm"
)

type Teacher struct {
	ID      int64
	Name    string
	Sex     uint
	Age     uint
	ClassNo uint
}

func (Teacher) TableName() string {
	return "demo_teacher"
}

type TeacherMapper struct {
	gormstarter.BaseMapper[Teacher]
}

type TeacherColumns struct {
	ID   gormstarter.Column[Teacher, int64]
	Name gormstarter.Column[Teacher, string]
	Age  gormstarter.Column[Teacher, uint]
}

var teacherColumns = TeacherColumns{
	ID:   gormstarter.NewColumn(func(teacher *Teacher) *int64 { return &teacher.ID }),
	Name: gormstarter.NewColumn(func(teacher *Teacher) *string { return &teacher.Name }),
	Age:  gormstarter.NewColumn(func(teacher *Teacher) *uint { return &teacher.Age }),
}

// Columns 返回 Teacher Mapper 的类型安全字段集合。
func (TeacherMapper) Columns() *TeacherColumns {
	return &teacherColumns
}

// WithTxMapper 返回绑定事务的新 Mapper，避免修改共享 Mapper 实例。
func (t TeacherMapper) WithTxMapper(tx *gorm.DB) TeacherMapper {
	return TeacherMapper{BaseMapper: t.GetBaseMapperWithTx(tx)}
}

func (t TeacherMapper) CountAll() (total int64) {
	db := t.TableGormDB()
	db.Count(&total)
	return total
}
