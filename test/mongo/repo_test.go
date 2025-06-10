package mongo

import (
	"fmt"
	"github.com/acexy/golang-toolkit/util/json"
	"github.com/golang-acexy/cloud-database/databasecloud"
	"github.com/golang-acexy/starter-mongo/mongostarter"
	"testing"
)

func TestSave(t *testing.T) {
	fmt.Println(teacherRepo.Save(&Teacher{Name: "test"}))
	fmt.Println(teacherRepo.SaveBatch(&[]*Teacher{
		{Name: "test1"},
		{Name: "test3"},
	}))
	fmt.Println(teacherRepo.SaveBatch(&[]*Teacher{
		{Name: "test"},
		{Name: "test"},
		{Name: "test"},
		{Name: "test"},
		{Name: "test"},
		{Name: "test"},
		{Name: "test"},
		{Name: "test"},
		{Name: "test"},
		{Name: "test", Age: 10},
		{Name: "test", Age: 11},
		{Name: "test", Age: 12},
	}))
}

func TestQueryPage(t *testing.T) {
	pager := databasecloud.Pager[Teacher]{
		Number: 1,
		Size:   3,
	}
	fmt.Println(teacherRepo.QueryPageByCond(&Teacher{Name: "test"}, mongostarter.NewOrderBy("age", false), &pager))
	fmt.Println(json.ToJson(pager))
}
