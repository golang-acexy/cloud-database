package gorm

import (
	"errors"
	"testing"

	"github.com/golang-acexy/cloud-database/databasecloud"
	"github.com/golang-acexy/cloud-database/databasecloud/rds"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"gorm.io/gorm"
)

var teacherRepo = NewTeacherRepo()

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
	return t.RawMapper().SelectOneByMap(gormstarter.MapQuery{Condition: map[string]any{"id": 1}}, result)
}

func (t TeacherRepo) CountByName(name string) (int64, error) {
	return t.CountByMap(rds.MapQuery{Condition: map[string]any{"name": name}})
}

func saveTeacher(t *testing.T, name string, age uint) *Teacher {
	t.Helper()
	teacher := &Teacher{Name: name, Age: age, Sex: 1, ClassNo: 1}
	if rows, err := teacherRepo.Save(teacher); err != nil || rows != 1 {
		t.Fatalf("save teacher failed: rows=%d err=%v", rows, err)
	}
	if teacher.ID == 0 {
		t.Fatal("expected generated teacher ID")
	}
	return teacher
}

func removeTeachers(t *testing.T, ids ...int64) {
	t.Helper()
	if len(ids) == 0 {
		return
	}
	values := make([]any, 0, len(ids))
	for _, id := range ids {
		values = append(values, id)
	}
	if _, err := teacherRepo.RemoveByIDs(values); err != nil {
		t.Fatal(err)
	}
}

func TestSaveAndQueryVariants(t *testing.T) {
	teacher := saveTeacher(t, "rsave", 18)
	zero := &Teacher{Name: "rzero", Age: 19}
	if _, err := teacherRepo.SaveWithoutZeroFields(zero); err != nil {
		t.Fatal(err)
	}
	batch := []*Teacher{{Name: "rb1", Age: 20}, {Name: "rb2", Age: 21}}
	if rows, err := teacherRepo.SaveBatch(batch); err != nil || rows != 2 {
		t.Fatalf("save batch failed: rows=%d err=%v", rows, err)
	}
	defer removeTeachers(t, teacher.ID, zero.ID, batch[0].ID, batch[1].ID)

	if rows, err := teacherRepo.SaveWithMap(map[string]any{"name": "rmap", "age": 22}); err != nil || rows != 1 {
		t.Fatalf("save map failed: rows=%d err=%v", rows, err)
	}
	mapTeacher, err := teacherRepo.QueryOneByMap(rds.MapQuery{Condition: map[string]any{"name": "rmap"}})
	if err != nil {
		t.Fatalf("query map teacher failed: err=%v", err)
	}
	defer removeTeachers(t, mapTeacher.ID)

	teacher.Name = "rup"
	if rows, err := teacherRepo.SaveOrModifyByPrimaryKey(teacher); err != nil || rows != 1 {
		t.Fatalf("save or modify failed: rows=%d err=%v", rows, err)
	}

	exists, err := teacherRepo.ExistsByID(teacher.ID)
	if err != nil || !exists {
		t.Fatalf("expected teacher to exist: exists=%v err=%v", exists, err)
	}
	selected, err := teacherRepo.QueryByID(teacher.ID)
	if err != nil || selected.Name != "rup" {
		t.Fatalf("unexpected ID result: teacher=%+v err=%v", selected, err)
	}
	selectedBatch, err := teacherRepo.QueryByIDs([]any{batch[0].ID, batch[1].ID})
	if err != nil || len(selectedBatch) != 2 {
		t.Fatalf("unexpected IDs result: teachers=%+v err=%v", selectedBatch, err)
	}
	if selected, err = teacherRepo.QueryOneByCond(rds.CondQuery[Teacher]{Condition: Teacher{Name: "rb1"}}); err != nil {
		t.Fatalf("query one by condition failed: err=%v", err)
	}
	if selected, err = teacherRepo.QueryOneByWhere(rds.WhereQuery{RawWhereSQL: "id = ?", Args: []any{teacher.ID}}); err != nil {
		t.Fatalf("query one by where failed: err=%v", err)
	}
	if selected, err = teacherRepo.QueryOneByGorm(func(db *gorm.DB) { db.Where("id = ?", teacher.ID) }); err != nil {
		t.Fatalf("query one by GORM failed: err=%v", err)
	}

	list, err := teacherRepo.QueryByCond(rds.NewCondQuery(Teacher{Name: "rb1"}).OrderBy("id").WithLimit(1))
	if err != nil || len(list) != 1 {
		t.Fatalf("query by condition failed: err=%v", err)
	}
	if list, err = teacherRepo.QueryByMap(rds.MapQuery{Condition: map[string]any{"name": "rb2"}, QueryOptions: rds.QueryOptions{OrderBySQL: "id"}}); err != nil || len(list) != 1 {
		t.Fatalf("query by map failed: err=%v", err)
	}
	if list, err = teacherRepo.QueryByWhere(rds.WhereQuery{RawWhereSQL: "name in ?", Args: []any{[]string{"rb1", "rb2"}}, QueryOptions: rds.QueryOptions{OrderBySQL: "id"}}); err != nil || len(list) != 2 {
		t.Fatalf("query by where failed: err=%v", err)
	}
	if list, err = teacherRepo.QueryByGorm(func(db *gorm.DB) { db.Where("name in ?", []string{"rb1", "rb2"}) }); err != nil || len(list) != 2 {
		t.Fatalf("query by GORM failed: err=%v", err)
	}

	if count, err := teacherRepo.CountByCond(rds.CondQuery[Teacher]{Condition: Teacher{Name: "rb1"}}); err != nil || count != 1 {
		t.Fatalf("unexpected condition count: count=%d err=%v", count, err)
	}
	if count, err := teacherRepo.CountByMap(rds.MapQuery{Condition: map[string]any{"name": "rb2"}}); err != nil || count != 1 {
		t.Fatalf("unexpected map count: count=%d err=%v", count, err)
	}
	if count, err := teacherRepo.CountByWhere(rds.WhereQuery{RawWhereSQL: "name in ?", Args: []any{[]string{"rb1", "rb2"}}}); err != nil || count != 2 {
		t.Fatalf("unexpected where count: count=%d err=%v", count, err)
	}
	if count, err := teacherRepo.CountByGorm(func(db *gorm.DB) { db.Where("name in ?", []string{"rb1", "rb2"}) }); err != nil || count != 2 {
		t.Fatalf("unexpected GORM count: count=%d err=%v", count, err)
	}

	if teacherRepo.RawMapper().CountAll() == 0 {
		t.Fatal("expected custom mapper method to return records")
	}
	if db := teacherRepo.CurrentGormDB(); db == nil {
		t.Fatal("current Gorm DB unavailable")
	}
	if db := teacherRepo.TableGormDB(); db == nil {
		t.Fatal("table Gorm DB unavailable")
	}
}

func TestPaginationVariants(t *testing.T) {
	teachers := make([]*Teacher, 0, 5)
	ids := make([]int64, 0, 5)
	for age := uint(1); age <= 5; age++ {
		teacher := saveTeacher(t, "rpage", age)
		teachers = append(teachers, teacher)
		ids = append(ids, teacher.ID)
	}
	defer removeTeachers(t, ids...)

	assertPage := func(name string, pager databasecloud.Pager[Teacher], err error) {
		t.Helper()
		if err != nil || pager.Total != 5 || len(pager.Records) != 2 {
			t.Fatalf("unexpected %s page: total=%d records=%d err=%v", name, pager.Total, len(pager.Records), err)
		}
	}

	pager, err := teacherRepo.QueryPageByCond(rds.NewPageQuery(Teacher{Name: "rpage"}, 2, 2).OrderBy("age"))
	assertPage("condition", pager, err)
	pager, err = teacherRepo.QueryPageByMap(rds.NewMapPageQuery(map[string]any{"name": "rpage"}, 2, 2).OrderBy("age").Select("id", "age"))
	assertPage("map", pager, err)
	pager, err = teacherRepo.QueryPageByWhere(rds.NewWherePageQuery("name = ?", 2, 2, "rpage").OrderBy("age"))
	assertPage("where", pager, err)

	pager, err = teacherRepo.QueryPageByGorm(rds.GormPageQuery{
		CountRawDB:  func(db *gorm.DB) { db.Where("name = ?", "rpage") },
		PageRawDB:   func(db *gorm.DB) { db.Where("name = ?", "rpage").Order("age").Offset(2).Limit(2) },
		PageOptions: rds.PageOptions{Number: 2, Size: 2},
	})
	if err != nil || pager.Total != 5 || len(pager.Records) != 2 {
		t.Fatalf("unexpected GORM page: total=%d records=%d err=%v", pager.Total, len(pager.Records), err)
	}
}

func TestWrapperVariants(t *testing.T) {
	first := saveTeacher(t, "rwrapper", 18)
	second := saveTeacher(t, "rwrapper", 20)
	defer removeTeachers(t, first.ID, second.ID)

	c := teacherRepo.RawMapper().Columns()

	selected, err := teacherRepo.QueryOneByWrapper(
		teacherRepo.Wrapper().Eq(c.ID, first.ID),
	)
	if err != nil || selected == nil || selected.ID != first.ID {
		t.Fatalf("unexpected Wrapper single result: teacher=%+v err=%v", selected, err)
	}

	list, err := teacherRepo.QueryByWrapper(
		teacherRepo.Wrapper().Eq(c.Name, "rwrapper").OrderByAsc(c.Age),
	)
	if err != nil || len(list) != 2 || list[0].Age != 18 || list[1].Age != 20 {
		t.Fatalf("unexpected Wrapper results: teachers=%+v err=%v", list, err)
	}

	count, err := teacherRepo.CountByWrapper(teacherRepo.Wrapper().Eq(c.Name, "rwrapper"))
	if err != nil || count != 2 {
		t.Fatalf("unexpected Wrapper count: count=%d err=%v", count, err)
	}

	pager, err := teacherRepo.QueryPageByWrapper(
		teacherRepo.PageWrapper(2, 1).Eq(c.Name, "rwrapper").OrderByAsc(c.Age),
	)
	if err != nil || pager.Number != 2 || pager.Size != 1 || pager.Total != 2 || len(pager.Records) != 1 || pager.Records[0].ID != second.ID {
		t.Fatalf("unexpected Wrapper page: pager=%+v err=%v", pager, err)
	}

	rows, err := teacherRepo.ModifyByWrapper(
		teacherRepo.ModifyWrapper().Eq(c.ID, first.ID).Set(c.Age, uint(0)),
	)
	if err != nil || rows != 1 {
		t.Fatalf("unexpected Wrapper update: rows=%d err=%v", rows, err)
	}
	updated, err := teacherRepo.QueryByID(first.ID)
	if err != nil || updated.Age != 0 {
		t.Fatalf("unexpected Wrapper update result: teacher=%+v err=%v", updated, err)
	}
}

func TestModifyAndRemoveVariants(t *testing.T) {
	teacher := saveTeacher(t, "rm1", 10)
	defer removeTeachers(t, teacher.ID)

	teacher.Name = "rm2"
	if _, err := teacherRepo.ModifyByID(teacher, teacher.ID); err != nil {
		t.Fatal(err)
	}
	teacher.Name, teacher.Age, teacher.Sex = "rm3", 0, 0
	if _, err := teacherRepo.ModifyByIDWithoutZeroFields(teacher, teacher.ID, "sex"); err != nil {
		t.Fatal(err)
	}
	if _, err := teacherRepo.ModifyByIDWithMap(map[string]any{"name": "rm4"}, teacher.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := teacherRepo.ModifyByCond(&Teacher{Name: "rm5"}, Teacher{ID: teacher.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := teacherRepo.ModifyByCondWithZeroFields(&Teacher{Age: 0}, Teacher{ID: teacher.ID}, "age"); err != nil {
		t.Fatal(err)
	}
	if _, err := teacherRepo.ModifyByMap(map[string]any{"name": "rm6"}, map[string]any{"id": teacher.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := teacherRepo.ModifyByWhere(&Teacher{Name: "rm7"}, "id = ?", teacher.ID); err != nil {
		t.Fatal(err)
	}
	selected, err := teacherRepo.QueryByID(teacher.ID)
	if err != nil || selected.Name != "rm7" || selected.Age != 0 || selected.Sex != 0 {
		t.Fatalf("unexpected modified teacher: %+v err=%v", selected, err)
	}

	byID := saveTeacher(t, "rdid", 1)
	byIDs1 := saveTeacher(t, "rdi1", 1)
	byIDs2 := saveTeacher(t, "rdi2", 1)
	byCond := saveTeacher(t, "rdc", 1)
	byMap := saveTeacher(t, "rdm", 1)
	byWhere := saveTeacher(t, "rdw", 1)
	if rows, err := teacherRepo.RemoveByID(byID.ID); err != nil || rows != 1 {
		t.Fatalf("remove by ID failed: rows=%d err=%v", rows, err)
	}
	if rows, err := teacherRepo.RemoveByIDs([]any{byIDs1.ID, byIDs2.ID}); err != nil || rows != 2 {
		t.Fatalf("remove by IDs failed: rows=%d err=%v", rows, err)
	}
	if rows, err := teacherRepo.RemoveByCond(Teacher{ID: byCond.ID}); err != nil || rows != 1 {
		t.Fatalf("remove by condition failed: rows=%d err=%v", rows, err)
	}
	if rows, err := teacherRepo.RemoveByMap(map[string]any{"id": byMap.ID}); err != nil || rows != 1 {
		t.Fatalf("remove by map failed: rows=%d err=%v", rows, err)
	}
	if rows, err := teacherRepo.RemoveByWhere("id = ?", byWhere.ID); err != nil || rows != 1 {
		t.Fatalf("remove by where failed: rows=%d err=%v", rows, err)
	}
}

func TestTransactionRepositories(t *testing.T) {
	txRepo := teacherRepo.NewTxRepo()
	manual := &Teacher{Name: "rtx1", Age: 1}
	if _, err := txRepo.Save(manual); err != nil {
		t.Fatal(err)
	}
	if count, err := txRepo.CountByName("rtx1"); err != nil || count != 1 {
		t.Fatalf("custom repository method unavailable in transaction: count=%d err=%v", count, err)
	}
	tx := txRepo.CurrentGormDB()
	if err := tx.Rollback().Error; err != nil {
		t.Fatal(err)
	}
	if exists, err := teacherRepo.ExistsByID(manual.ID); err != nil || exists {
		t.Fatalf("manual rollback failed: exists=%v err=%v", exists, err)
	}

	db := gormstarter.RawMysqlGormDB()
	if db == nil {
		t.Fatal("mysql database is not initialized")
	}
	externalTx := db.Begin()
	externalRepo := teacherRepo.WithTxRepo(externalTx)
	external := &Teacher{Name: "rtx2", Age: 2}
	if _, err := externalRepo.Save(external); err != nil {
		t.Fatal(err)
	}
	if err := externalTx.Rollback().Error; err != nil {
		t.Fatal(err)
	}
	if exists, err := teacherRepo.ExistsByID(external.ID); err != nil || exists {
		t.Fatalf("external rollback failed: exists=%v err=%v", exists, err)
	}

	var committedID int64
	if err := teacherRepo.Transaction(func(repo TeacherRepo) error {
		committed := &Teacher{Name: "rtxc", Age: 3}
		if _, saveErr := repo.Save(committed); saveErr != nil {
			return saveErr
		}
		committedID = committed.ID
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	defer removeTeachers(t, committedID)
	if exists, err := teacherRepo.ExistsByID(committedID); err != nil || !exists {
		t.Fatalf("automatic commit failed: exists=%v err=%v", exists, err)
	}

	rollbackErr := errors.New("rollback transaction")
	var rolledBackID int64
	err := teacherRepo.Transaction(func(repo TeacherRepo) error {
		rolledBack := &Teacher{Name: "rtxr", Age: 4}
		if _, saveErr := repo.Save(rolledBack); saveErr != nil {
			return saveErr
		}
		rolledBackID = rolledBack.ID
		return rollbackErr
	})
	if !errors.Is(err, rollbackErr) {
		t.Fatalf("expected rollback error, got %v", err)
	}
	if exists, err := teacherRepo.ExistsByID(rolledBackID); err != nil || exists {
		t.Fatalf("automatic rollback failed: exists=%v err=%v", exists, err)
	}
}

func TestEmptyQueryAndInvalidPage(t *testing.T) {
	missingName := "cloud_database_missing_teacher"
	c := teacherRepo.RawMapper().Columns()

	if result, err := teacherRepo.QueryByID(int64(-1)); err != nil || result != nil {
		t.Fatalf("不存在的 ID 应返回 nil：result=%+v err=%v", result, err)
	}
	if result, err := teacherRepo.QueryOneByCond(rds.NewCondQuery(Teacher{Name: missingName})); err != nil || result != nil {
		t.Fatalf("实体条件无结果时应返回 nil：result=%+v err=%v", result, err)
	}
	if result, err := teacherRepo.QueryOneByMap(rds.NewMapQuery(map[string]any{"name": missingName})); err != nil || result != nil {
		t.Fatalf("Map 条件无结果时应返回 nil：result=%+v err=%v", result, err)
	}
	if result, err := teacherRepo.QueryOneByWhere(rds.NewWhereQuery("name = ?", missingName)); err != nil || result != nil {
		t.Fatalf("Where 条件无结果时应返回 nil：result=%+v err=%v", result, err)
	}
	if result, err := teacherRepo.QueryOneByGorm(func(db *gorm.DB) { db.Where("name = ?", missingName) }); err != nil || result != nil {
		t.Fatalf("GORM 条件无结果时应返回 nil：result=%+v err=%v", result, err)
	}
	if result, err := teacherRepo.QueryOneByWrapper(teacherRepo.Wrapper().Eq(c.Name, missingName)); err != nil || result != nil {
		t.Fatalf("Wrapper 条件无结果时应返回 nil：result=%+v err=%v", result, err)
	}

	if _, err := teacherRepo.QueryPageByCond(rds.NewPageQuery(Teacher{}, 0, 10)); !errors.Is(err, gormstarter.ErrInvalidPage) {
		t.Fatalf("非法实体分页参数应返回 ErrInvalidPage，实际为 %v", err)
	}
	if _, err := teacherRepo.QueryPageByMap(rds.NewMapPageQuery(map[string]any{}, 1, 0)); !errors.Is(err, gormstarter.ErrInvalidPage) {
		t.Fatalf("非法 Map 分页参数应返回 ErrInvalidPage，实际为 %v", err)
	}
	if _, err := teacherRepo.QueryPageByWhere(rds.NewWherePageQuery("1 = 1", -1, 10)); !errors.Is(err, gormstarter.ErrInvalidPage) {
		t.Fatalf("非法 Where 分页参数应返回 ErrInvalidPage，实际为 %v", err)
	}
}

func TestTransactionRollsBackOnPanic(t *testing.T) {
	panicValue := "transaction panic"
	var teacherID int64

	func() {
		defer func() {
			if recovered := recover(); recovered != panicValue {
				t.Fatalf("事务应继续抛出原 panic，实际为 %v", recovered)
			}
		}()
		_ = teacherRepo.Transaction(func(repo TeacherRepo) error {
			teacher := &Teacher{Name: "rtxp", Age: 5}
			if _, err := repo.Save(teacher); err != nil {
				return err
			}
			teacherID = teacher.ID
			panic(panicValue)
		})
	}()

	if exists, err := teacherRepo.ExistsByID(teacherID); err != nil || exists {
		t.Fatalf("panic 后事务应回滚：exists=%v err=%v", exists, err)
	}
}
