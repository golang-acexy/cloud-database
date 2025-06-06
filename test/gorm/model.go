package gorm

import (
	"github.com/golang-acexy/cloud-database/databasecloud"
	"github.com/golang-acexy/starter-gorm/gormstarter"
)

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
func (Teacher) DBType() gormstarter.DBType {
	return gormstarter.DBTypeMySQL
}

type TeacherMapper struct {
	gormstarter.BaseMapper[Teacher]
}

type TeacherDBService struct {
	databasecloud.GormDBService[gormstarter.IBaseMapper[gormstarter.BaseMapper[Teacher], Teacher], gormstarter.BaseMapper[Teacher], Teacher]
}

func NewTeacherDBService() TeacherDBService {
	return TeacherDBService{
		GormDBService: databasecloud.GormDBService[
			gormstarter.IBaseMapper[gormstarter.BaseMapper[Teacher], Teacher],
			gormstarter.BaseMapper[Teacher],
			Teacher,
		]{
			Mapper: TeacherMapper{},
		},
	}
}
func (t TeacherDBService) QueryByMap(result *Teacher) (int64, error) {
	return t.RawMapper().SelectOneByMap(map[string]interface{}{"id": 1}, result)
}
