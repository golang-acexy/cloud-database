package mongo

import (
	"github.com/golang-acexy/cloud-database/databasecloud/mongo"
	"github.com/golang-acexy/starter-mongo/mongostarter"
)

var teacherRepo = TeacherRepo{
	Repository: mongo.NewRepository[TeacherMapper, Teacher](TeacherMapper{}),
}

type Teacher struct {
	ID      string `bson:"_id,omitempty" json:"id"`
	Name    string `bson:"name,omitempty" json:"name"`
	Sex     uint   `bson:"sex,omitempty" json:"sex"`
	Age     uint   `bson:"age,omitempty" json:"age,omitempty"`
	ClassNo uint   `bson:"class_no,omitempty" json:"class_no,omitempty"`
}

func (Teacher) CollectionName() string {
	return "demo_teacher"
}

type TeacherMapper struct {
	mongostarter.BaseMapper[Teacher]
}

type TeacherRepo struct {
	mongo.Repository[TeacherMapper, Teacher]
}

func NewTeacherRepo() TeacherRepo {
	return teacherRepo
}
