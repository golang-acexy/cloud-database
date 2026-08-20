package gorm

import (
	"errors"
	"reflect"
	"testing"

	"github.com/golang-acexy/cloud-database/databasecloud/rds"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"gorm.io/gorm"
)

func TestQueryFacade(t *testing.T) {
	timeRange := gormstarter.TimeRange{Field: "created_at"}

	cond := rds.NewCondQuery(Teacher{Sex: 1}).OrderBy("id desc").Select("id", "name").WithTimeRanges(timeRange).WithLimit(10)
	if cond.Condition.Sex != 1 || cond.OrderBySQL != "id desc" || cond.Limit != 10 || !reflect.DeepEqual(cond.SelectColumns, []string{"id", "name"}) || len(cond.TimeRanges) != 1 {
		t.Fatalf("CondQuery 门面转发异常: %+v", cond)
	}

	mapQuery := rds.NewMapQuery(map[string]any{"sex": 0}).OrderBy("id").Select("id").WithLimit(5)
	if mapQuery.Condition["sex"] != 0 || mapQuery.OrderBySQL != "id" || mapQuery.Limit != 5 {
		t.Fatalf("MapQuery 门面转发异常: %+v", mapQuery)
	}

	whereQuery := rds.NewWhereQuery("sex = ?", 1).OrderBy("id").WithTimeRanges(timeRange).WithLimit(3)
	if whereQuery.RawWhereSQL != "sex = ?" || !reflect.DeepEqual(whereQuery.Args, []any{1}) || whereQuery.Limit != 3 {
		t.Fatalf("WhereQuery 门面转发异常: %+v", whereQuery)
	}

	page := rds.NewPageQuery(Teacher{Sex: 1}, 2, 20).OrderBy("id").Select("id").WithTimeRanges(timeRange)
	mapPage := rds.NewMapPageQuery(map[string]any{"sex": 0}, 3, 30).OrderBy("id")
	wherePage := rds.NewWherePageQuery("sex = ?", 4, 40, 1).Select("id")
	if page.Number != 2 || page.Size != 20 || mapPage.Number != 3 || mapPage.Size != 30 || wherePage.Number != 4 || wherePage.Size != 40 {
		t.Fatalf("分页 Query 门面转发异常: page=%+v map=%+v where=%+v", page, mapPage, wherePage)
	}

	gormPage := rds.NewGormPageQuery(5, 50, func(*gorm.DB) {}, func(*gorm.DB) {})
	if gormPage.Number != 5 || gormPage.Size != 50 || gormPage.CountRawDB == nil || gormPage.PageRawDB == nil {
		t.Fatalf("GormPageQuery 构建异常: %+v", gormPage)
	}

	// 编译期赋值验证 Wrapper 门面与 starter 类型完全一致。
	var queryWrapper *rds.QueryWrapper[Teacher] = teacherRepo.Wrapper()
	var pageWrapper *rds.PageWrapper[Teacher] = teacherRepo.PageWrapper(1, 10)
	var modifyWrapper *rds.ModifyWrapper[Teacher] = teacherRepo.ModifyWrapper()
	if queryWrapper == nil || pageWrapper == nil || modifyWrapper == nil {
		t.Fatal("Wrapper 门面不应返回 nil")
	}
}

func TestRepositoryValidation(t *testing.T) {
	func() {
		defer func() {
			if recovered := recover(); !errors.Is(asError(recovered), rds.ErrNilRepositoryFactory) {
				t.Fatalf("空 Repository 工厂应触发 ErrNilRepositoryFactory，实际为 %v", recovered)
			}
		}()
		rds.NewRepository[TeacherRepo, TeacherMapper, Teacher](TeacherMapper{}, nil)
	}()

	if err := teacherRepo.Transaction(nil); !errors.Is(err, rds.ErrNilTransactionFunc) {
		t.Fatalf("空事务函数应返回 ErrNilTransactionFunc，实际为 %v", err)
	}

	var repository rds.Repository[TeacherRepo, TeacherMapper, Teacher]
	if err := repository.Transaction(func(TeacherRepo) error { return nil }); !errors.Is(err, rds.ErrRepositoryNotInitialized) {
		t.Fatalf("未初始化 Repository 应返回 ErrRepositoryNotInitialized，实际为 %v", err)
	}
}

func asError(value any) error {
	err, _ := value.(error)
	return err
}
