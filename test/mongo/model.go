package mongo

import (
	"github.com/golang-acexy/cloud-database/databasecloud"
	"github.com/golang-acexy/starter-mongo/mongostarter"
)

var teacherRepo = TeacherRepo{
	MongoRepository: databasecloud.MongoRepository[
		mongostarter.IBaseMapper[mongostarter.BaseMapper[Teacher], Teacher],
		mongostarter.BaseMapper[Teacher], Teacher,
	]{
		Mapper: TeacherMapper{},
	},
}

type Teacher struct {
	ID      string `bson:"_id,omitempty" json:"id"`
	Name    string `bson:"name,omitempty" json:"name"`
	Sex     uint   `bson:"sex,omitempty" json:"sex"`
	Age     uint   `json:"age,omitempty"`
	ClassNo uint   `json:"class_no,omitempty"`
}

func (Teacher) CollectionName() string {
	return "demo_teacher"
}

type TeacherMapper struct {
	mongostarter.BaseMapper[Teacher]
}

type TeacherRepo struct {
	databasecloud.MongoRepository[mongostarter.IBaseMapper[mongostarter.BaseMapper[Teacher], Teacher], mongostarter.BaseMapper[Teacher], Teacher]
}

func (t TeacherRepo) RawMapper() TeacherMapper {
	return t.RawIMapper().(TeacherMapper)
}

func NewTeacherRepo() TeacherRepo {
	return teacherRepo
}
