package gorm

import (
	"github.com/golang-acexy/cloud-database/databasecloud/rds"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"gorm.io/gorm"
)

var teacherRepo = NewTeacherRepo()

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

// WithTxMapper 返回绑定事务的新 Mapper，避免修改共享 Mapper 实例。
func (t TeacherMapper) WithTxMapper(tx *gorm.DB) TeacherMapper {
	return TeacherMapper{BaseMapper: t.GetBaseMapperWithTx(tx)}
}

func (t TeacherMapper) CountAll() (total int64) {
	db, err := t.GormWithTableName()
	if err != nil {
		return 0
	}
	db.Count(&total)
	return total
}

type TeacherRepo struct {
	rds.Repository[TeacherRepo, TeacherMapper, Teacher]
}

func NewTeacherRepo() TeacherRepo {
	repository := rds.NewRepository(
		TeacherMapper{},
		func(base rds.Repository[TeacherRepo, TeacherMapper, Teacher]) TeacherRepo {
			return TeacherRepo{Repository: base}
		},
	)
	return TeacherRepo{Repository: repository}
}
func (t TeacherRepo) QueryTeacherByMap(result *Teacher) (int64, error) {
	return t.RawMapper().SelectOneByMap(map[string]any{"id": 1}, result)
}

func (t TeacherRepo) CountByName(name string) (int64, error) {
	return t.CountByMap(map[string]any{"name": name})
}
