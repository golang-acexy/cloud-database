package test

import (
	"github.com/golang-acexy/cloud-database/databasecloud"
	"github.com/golang-acexy/starter-gorm/gormstarter"
)

type Teacher struct {
	gormstarter.BaseModel[int64]
}

func (Teacher) TableName() string {
	return "demo_teacher"
}

//type TeacherMapper struct {
//	gormstarter.BaseMapper[Teacher]
//}

type TeacherDBService struct {
	databasecloud.GormDBService[
		gormstarter.IBaseMapper[gormstarter.BaseMapper[Teacher], Teacher], gormstarter.BaseMapper[Teacher],
		Teacher,
	]
}

func init() {
	var teacherDBService TeacherDBService
	var teacher Teacher
	teacherDBService.QueryByID(1, &teacher)
}
