package gorm

import (
	"fmt"
	"github.com/acexy/golang-toolkit/util/json"
	"github.com/golang-acexy/cloud-database/databasecloud"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"testing"
)

func TestTeacherRepo(t *testing.T) {
	teacherRepo = NewTeacherRepo()
	var teacher Teacher
	teacherRepo.QueryByID(1, &teacher)
	teacherRepo.QueryByID(2, &teacher)
	tx := gormstarter.RawMysqlGormDB().Begin()
	teacherDBTxService := teacherRepo.WithTxRepo(tx)
	teacherDBTxService.QueryByID(2, &teacher)
}

func TestTeacherTxRepo(t *testing.T) {
	teacherTxDBService1 := teacherRepo.NewTxRepo()
	teacherTxDBService2 := teacherRepo.NewTxRepo()
	var tx1Teacher = Teacher{
		Name: "tx1",
		Age:  18,
	}
	_, _ = teacherTxDBService1.Save(&tx1Teacher)
	tx1TeacherId := tx1Teacher.ID
	fmt.Println("事务1中 新增数据 id ", tx1TeacherId)
	row, _ := teacherTxDBService1.QueryByID(tx1TeacherId, &Teacher{})
	fmt.Println("事务1中 查询数据 该id后返回结果 ", row)

	row, _ = teacherTxDBService2.QueryByID(tx1TeacherId, &Teacher{})
	fmt.Println("事务2中 尝试查询数据 该id后返回结果 ", row)
	row, _ = teacherRepo.QueryByID(tx1TeacherId, &Teacher{})
	fmt.Println("无事务中 尝试查询数据 该id后返回结果 ", row)

	fmt.Println("事务1中 提交事务")
	_ = teacherTxDBService1.CurrentGormDB().Commit()
	row, _ = teacherTxDBService2.QueryByID(tx1TeacherId, &Teacher{})
	fmt.Println("事务2中 尝试查询数据 该id后返回结果 ", row)
	row, _ = teacherRepo.QueryByID(tx1TeacherId, &Teacher{})
	fmt.Println("无事务中 尝试查询数据 该id后返回结果 ", row)

	var tx2Teacher = Teacher{
		Name: "tx2",
		Age:  19,
	}
	_, _ = teacherTxDBService2.Save(&tx2Teacher)
	tx2TeacherId := tx2Teacher.ID
	fmt.Println("事务2中 新增数据 id ", tx2TeacherId)
	row, _ = teacherTxDBService2.QueryByID(tx2TeacherId, &Teacher{})
	fmt.Println("事务2中 查询数据 该id后返回结果 ", row)
	fmt.Println("事务2中 回滚事务")
	teacherTxDBService2.CurrentGormDB().Rollback()
	row, _ = teacherRepo.QueryByID(tx2TeacherId, &Teacher{})
	fmt.Println("无事务中 尝试查询数据 该id后返回结果 ", row)
}

func TestTeacherPager(t *testing.T) {
	var teacher = Teacher{
		Name: "tx1",
	}
	pager := databasecloud.Pager[Teacher]{
		Number: 2,
		Size:   3,
	}
	err := teacherRepo.QueryPageByCond(&teacher, "id desc", &pager)
	if err != nil {
		fmt.Println("查询失败", err)
	}
	fmt.Println(json.ToJson(pager))

	err = teacherRepo.QueryPageByMap(map[string]any{"name": "tx1"}, "", &pager, "id", "class_no")
	if err != nil {
		fmt.Println("查询失败", err)
	}
	fmt.Println(json.ToJson(pager))

	err = teacherRepo.QueryPageByWhere("name = ?", "", &pager, []any{"tx1"}, "id")
	if err != nil {
		fmt.Println("查询失败", err)
	}
	fmt.Println(json.ToJson(pager))
}

func TestRawMapper(t *testing.T) {
	_ = teacherRepo.RawIMapper()
	teacherMapper := teacherRepo.RawMapper()
	fmt.Println(teacherMapper.CountAll())
}
