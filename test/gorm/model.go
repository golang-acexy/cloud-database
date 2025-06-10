package gorm

import (
	"github.com/golang-acexy/cloud-database/databasecloud"
	"github.com/golang-acexy/starter-gorm/gormstarter"
)

var teacherRepo TeacherRepo = TeacherRepo{
	GormRepository: databasecloud.GormRepository[
		gormstarter.IBaseMapper[gormstarter.BaseMapper[Teacher], Teacher],
		gormstarter.BaseMapper[Teacher],
		Teacher,
	]{
		Mapper: TeacherMapper{},
	},
}

type Teacher struct {
	gormstarter.BaseModel[int64]
	Name    string
	Sex     uint
	Age     uint
	ClassNo uint
}

func (Teacher) TableName() string {
	return "demo_teacher"
}

//func (Teacher) DBType() gormstarter.DBType {
//	return gormstarter.DBTypeMySQL
//}

type TeacherMapper struct {
	gormstarter.BaseMapper[Teacher]
}

func (t TeacherMapper) CountAll() (total int64) {
	t.GormWithTableName().Count(&total)
	return total
}

type TeacherRepo struct {
	databasecloud.GormRepository[gormstarter.IBaseMapper[gormstarter.BaseMapper[Teacher], Teacher], gormstarter.BaseMapper[Teacher], Teacher]
}

func (t TeacherRepo) RawMapper() TeacherMapper {
	return t.RawIMapper().(TeacherMapper)
}

func NewTeacherRepo() TeacherRepo {
	return teacherRepo
}
func (t TeacherRepo) QueryByMap(result *Teacher) (int64, error) {
	return t.RawIMapper().SelectOneByMap(map[string]interface{}{"id": 1}, result)
}
