package mongo

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang-acexy/cloud-database/databasecloud"
	databasecloudmongo "github.com/golang-acexy/cloud-database/databasecloud/mongo"
	"github.com/golang-acexy/starter-mongo/mongostarter"
	"go.mongodb.org/mongo-driver/v2/bson"
	drivermongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func newTestScope(t *testing.T) string {
	t.Helper()
	scope := fmt.Sprintf("cloud_database_test_%s_%d", t.Name(), time.Now().UnixNano())
	t.Cleanup(func() {
		if _, err := teacherRepo.RemoveWithOptions(bson.M{"name": bson.M{"$regex": "^" + scope}}); err != nil {
			t.Errorf("清理测试数据失败: %v", err)
		}
	})
	return scope
}

func saveTeacher(t *testing.T, teacher *Teacher) string {
	t.Helper()
	id, err := teacherRepo.Save(teacher)
	if err != nil {
		t.Fatalf("保存教师失败: %v", err)
	}
	if id == "" {
		t.Fatal("保存后应返回 ObjectID")
	}
	return id
}

func TestSaveAndQueryVariants(t *testing.T) {
	scope := newTestScope(t)
	id := saveTeacher(t, &Teacher{Name: scope + "_entity", Sex: 1, Age: 31, ClassNo: 1})
	bsonID, err := teacherRepo.SaveWithBSON(bson.M{"name": scope + "_bson", "age": uint(32), "class_no": uint(1)})
	if err != nil || bsonID == "" {
		t.Fatalf("BSON 保存失败: id=%q err=%v", bsonID, err)
	}
	optionID, err := teacherRepo.SaveWithOptions(bson.M{"name": scope + "_options", "age": uint(33), "class_no": uint(1)}, options.InsertOne())
	if err != nil || optionID == "" {
		t.Fatalf("Options 保存失败: id=%q err=%v", optionID, err)
	}

	batchIDs, err := teacherRepo.SaveBatch([]*Teacher{{Name: scope + "_batch_1", Age: 34, ClassNo: 2}, {Name: scope + "_batch_2", Age: 35, ClassNo: 2}})
	if err != nil || len(batchIDs) != 2 {
		t.Fatalf("实体批量保存结果不正确: ids=%v err=%v", batchIDs, err)
	}
	bsonBatchIDs, err := teacherRepo.SaveBatchWithBSON(bson.A{
		bson.M{"name": scope + "_bson_batch_1", "age": uint(36), "class_no": uint(3)},
		bson.M{"name": scope + "_bson_batch_2", "age": uint(37), "class_no": uint(3)},
	})
	if err != nil || len(bsonBatchIDs) != 2 {
		t.Fatalf("BSON 批量保存结果不正确: ids=%v err=%v", bsonBatchIDs, err)
	}
	optionBatchIDs, err := teacherRepo.SaveBatchWithOptions([]any{
		bson.M{"name": scope + "_option_batch_1", "age": uint(38), "class_no": uint(4)},
		bson.M{"name": scope + "_option_batch_2", "age": uint(39), "class_no": uint(4)},
	}, options.InsertMany().SetOrdered(true))
	if err != nil || len(optionBatchIDs) != 2 {
		t.Fatalf("Options 批量保存结果不正确: ids=%v err=%v", optionBatchIDs, err)
	}

	exists, err := teacherRepo.ExistsByID(id)
	if err != nil || !exists {
		t.Fatalf("已保存文档应存在: exists=%v err=%v", exists, err)
	}
	byID, err := teacherRepo.QueryByID(id)
	if err != nil || byID.Name != scope+"_entity" {
		t.Fatalf("按 ID 查询结果不正确: result=%+v err=%v", byID, err)
	}
	byIDs, err := teacherRepo.QueryByIDs([]any{id, bsonID})
	if err != nil || len(byIDs) != 2 {
		t.Fatalf("按 IDs 查询结果不正确: size=%d err=%v", len(byIDs), err)
	}

	one, err := teacherRepo.QueryOneByCond(databasecloudmongo.NewCondQuery(Teacher{Name: scope + "_entity"}).Select("name", "age"))
	if err != nil || one.Age != 31 {
		t.Fatalf("实体条件单条查询失败: result=%+v err=%v", one, err)
	}
	one, err = teacherRepo.QueryOneByBSON(databasecloudmongo.BSONQuery{Condition: bson.M{"name": scope + "_bson"}})
	if err != nil || one.Name != scope+"_bson" {
		t.Fatalf("BSON 单条查询失败: result=%+v err=%v", one, err)
	}
	one, err = teacherRepo.QueryOneWithOptions(bson.M{"name": scope + "_options"}, options.FindOne().SetProjection(bson.M{"name": 1}))
	if err != nil || one.Name != scope+"_options" {
		t.Fatalf("Options 单条查询失败: result=%+v err=%v", one, err)
	}

	records, err := teacherRepo.QueryByCond(databasecloudmongo.NewCondQuery(Teacher{ClassNo: 2}).WithOrderBy(mongostarter.OrderBy{Column: "age", Desc: true}).WithLimit(1))
	if err != nil || len(records) != 1 || records[0].Age != 35 {
		t.Fatalf("实体条件列表查询失败: records=%+v err=%v", records, err)
	}
	records, err = teacherRepo.QueryByBSON(databasecloudmongo.BSONQuery{Condition: bson.M{"class_no": uint(3)}, QueryOptions: databasecloudmongo.QueryOptions{OrderBy: mongostarter.NewOrderBy("age", false)}})
	if err != nil || len(records) != 2 || records[0].Age != 36 {
		t.Fatalf("BSON 列表查询失败: records=%+v err=%v", records, err)
	}
	records, err = teacherRepo.QueryWithOptions(bson.M{"class_no": uint(4)}, options.Find().SetSort(bson.D{{Key: "age", Value: -1}}))
	if err != nil || len(records) != 2 || records[0].Age != 39 {
		t.Fatalf("Options 列表查询失败: records=%+v err=%v", records, err)
	}

	if count, err := teacherRepo.CountByCond(databasecloudmongo.CondQuery[Teacher]{Condition: Teacher{ClassNo: 2}}); err != nil || count != 2 {
		t.Fatalf("实体条件计数失败: count=%d err=%v", count, err)
	}
	if count, err := teacherRepo.CountByBSON(databasecloudmongo.BSONQuery{Condition: bson.M{"class_no": uint(3)}}); err != nil || count != 2 {
		t.Fatalf("BSON 计数失败: count=%d err=%v", count, err)
	}
	if count, err := teacherRepo.CountWithOptions(bson.M{"class_no": uint(4)}, options.Count()); err != nil || count != 2 {
		t.Fatalf("Options 计数失败: count=%d err=%v", count, err)
	}

	if teacherRepo.RawMapper() != (TeacherMapper{}) {
		t.Fatal("RawMapper 应返回构造 Repository 时传入的具体 Mapper")
	}
	collection := teacherRepo.Collection()
	if collection == nil || collection.Name() != (Teacher{}).CollectionName() {
		t.Fatalf("Collection 获取失败: collection=%v", collection)
	}
}

func TestPaginationVariants(t *testing.T) {
	scope := newTestScope(t)
	entities := make([]*Teacher, 0, 5)
	for age := uint(21); age <= 25; age++ {
		entities = append(entities, &Teacher{Name: scope, Age: age, ClassNo: 9})
	}
	if ids, err := teacherRepo.SaveBatch(entities); err != nil || len(ids) != 5 {
		t.Fatalf("准备分页数据失败: ids=%v err=%v", ids, err)
	}
	assertPage := func(name string, pager databasecloud.Pager[Teacher], err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s 分页失败: %v", name, err)
		}
		if pager.Total != 5 || len(pager.Records) != 2 || pager.Records[0].Age != 23 {
			t.Fatalf("%s 分页结果不正确: %+v", name, pager)
		}
	}
	pager, err := teacherRepo.QueryPageByCond(databasecloudmongo.NewPageQuery(Teacher{Name: scope}, 2, 2).WithOrderBy(mongostarter.OrderBy{Column: "age"}))
	assertPage("Cond", pager, err)
	pager, err = teacherRepo.QueryPageByBSON(databasecloudmongo.NewBSONPageQuery(bson.M{"name": scope}, 2, 2).WithOrderBy(mongostarter.OrderBy{Column: "age"}))
	assertPage("BSON", pager, err)
	pager, err = teacherRepo.QueryPageWithOptions(databasecloudmongo.NewFilterPageQuery(bson.M{"name": scope}, 2, 2).WithOrderBy(mongostarter.OrderBy{Column: "age"}))
	assertPage("Options", pager, err)
}

func TestModifyAndRemoveVariants(t *testing.T) {
	scope := newTestScope(t)
	ids, err := teacherRepo.SaveBatch([]*Teacher{
		{Name: scope + "_id", Age: 20}, {Name: scope + "_bson_id", Age: 20},
		{Name: scope + "_cond_one", Age: 20}, {Name: scope + "_cond_many", Age: 20}, {Name: scope + "_cond_many", Age: 21},
		{Name: scope + "_bson_one", Age: 20}, {Name: scope + "_bson_many", Age: 20}, {Name: scope + "_bson_many", Age: 21},
		{Name: scope + "_option_one", Age: 20}, {Name: scope + "_option_many", Age: 20}, {Name: scope + "_option_many", Age: 21},
	})
	if err != nil || len(ids) != 11 {
		t.Fatalf("准备更新数据失败: ids=%v err=%v", ids, err)
	}
	assertAffected := func(name string, expected, affected int64, err error) {
		t.Helper()
		if err != nil || affected != expected {
			t.Fatalf("%s 影响数量不正确: affected=%d expected=%d err=%v", name, affected, expected, err)
		}
	}
	result, err := teacherRepo.ModifyByID(&Teacher{Age: 30}, ids[0])
	assertAffected("ModifyByID", 1, result, err)
	result, err = teacherRepo.ModifyByIDWithBSON(bson.M{"age": uint(31)}, ids[1])
	assertAffected("ModifyByIDWithBSON", 1, result, err)
	result, err = teacherRepo.ModifyOneByCond(&Teacher{Age: 32}, Teacher{Name: scope + "_cond_one"})
	assertAffected("ModifyOneByCond", 1, result, err)
	result, err = teacherRepo.ModifyByCond(&Teacher{Age: 33}, Teacher{Name: scope + "_cond_many"})
	assertAffected("ModifyByCond", 2, result, err)
	result, err = teacherRepo.ModifyOneByBSON(bson.M{"age": uint(34)}, bson.M{"name": scope + "_bson_one"})
	assertAffected("ModifyOneByBSON", 1, result, err)
	result, err = teacherRepo.ModifyByBSON(bson.M{"age": uint(35)}, bson.M{"name": scope + "_bson_many"})
	assertAffected("ModifyByBSON", 2, result, err)
	result, err = teacherRepo.ModifyOneWithOptions(bson.M{"name": scope + "_option_one"}, bson.M{"$set": bson.M{"age": uint(36)}}, options.UpdateOne())
	assertAffected("ModifyOneWithOptions", 1, result, err)
	result, err = teacherRepo.ModifyWithOptions(bson.M{"name": scope + "_option_many"}, bson.M{"$set": bson.M{"age": uint(37)}}, options.UpdateMany())
	assertAffected("ModifyWithOptions", 2, result, err)

	removeIDs, err := teacherRepo.SaveBatch([]*Teacher{
		{Name: scope + "_remove_id"}, {Name: scope + "_remove_ids_1"}, {Name: scope + "_remove_ids_2"},
		{Name: scope + "_remove_cond_one"}, {Name: scope + "_remove_cond_many"}, {Name: scope + "_remove_cond_many"},
		{Name: scope + "_remove_bson_one"}, {Name: scope + "_remove_bson_many"}, {Name: scope + "_remove_bson_many"},
		{Name: scope + "_remove_option_one"}, {Name: scope + "_remove_option_many"}, {Name: scope + "_remove_option_many"},
	})
	if err != nil || len(removeIDs) != 12 {
		t.Fatalf("准备删除数据失败: ids=%v err=%v", removeIDs, err)
	}
	result, err = teacherRepo.RemoveByID(removeIDs[0])
	assertAffected("RemoveByID", 1, result, err)
	result, err = teacherRepo.RemoveByIDs([]any{removeIDs[1], removeIDs[2]})
	assertAffected("RemoveByIDs", 2, result, err)
	result, err = teacherRepo.RemoveOneByCond(Teacher{Name: scope + "_remove_cond_one"})
	assertAffected("RemoveOneByCond", 1, result, err)
	result, err = teacherRepo.RemoveByCond(Teacher{Name: scope + "_remove_cond_many"})
	assertAffected("RemoveByCond", 2, result, err)
	result, err = teacherRepo.RemoveOneByBSON(bson.M{"name": scope + "_remove_bson_one"})
	assertAffected("RemoveOneByBSON", 1, result, err)
	result, err = teacherRepo.RemoveByBSON(bson.M{"name": scope + "_remove_bson_many"})
	assertAffected("RemoveByBSON", 2, result, err)
	result, err = teacherRepo.RemoveOneWithOptions(bson.M{"name": scope + "_remove_option_one"}, options.DeleteOne())
	assertAffected("RemoveOneWithOptions", 1, result, err)
	result, err = teacherRepo.RemoveWithOptions(bson.M{"name": scope + "_remove_option_many"}, options.DeleteMany())
	assertAffected("RemoveWithOptions", 2, result, err)
	exists, err := teacherRepo.ExistsByID(removeIDs[0])
	if err != nil || exists {
		t.Fatalf("删除后的文档不应存在: exists=%v err=%v", exists, err)
	}
}

func TestSafetyValidation(t *testing.T) {
	if _, err := teacherRepo.ModifyByCond(&Teacher{Age: 1}, Teacher{}); !errors.Is(err, mongostarter.ErrEmptyCondition) {
		t.Fatalf("空更新条件应返回 ErrEmptyCondition，实际为 %v", err)
	}
	if _, err := teacherRepo.RemoveByBSON(bson.M{}); !errors.Is(err, mongostarter.ErrEmptyCondition) {
		t.Fatalf("空删除条件应返回 ErrEmptyCondition，实际为 %v", err)
	}
	if _, err := teacherRepo.RemoveByIDs(nil); !errors.Is(err, mongostarter.ErrEmptyIDs) {
		t.Fatalf("空 ID 列表应返回 ErrEmptyIDs，实际为 %v", err)
	}
	if _, err := teacherRepo.QueryByID("invalid-object-id"); err == nil {
		t.Fatal("非法 ObjectID 应返回错误")
	}
	if _, err := teacherRepo.QueryOneByBSON(databasecloudmongo.BSONQuery{Condition: bson.M{"name": "cloud_database_missing_record"}}); !errors.Is(err, drivermongo.ErrNoDocuments) {
		t.Fatalf("查询不存在文档应返回 mongo.ErrNoDocuments，实际为 %v", err)
	}
	if _, err := teacherRepo.QueryOneByCond(databasecloudmongo.NewCondQuery(Teacher{Name: "cloud_database_missing_record"})); !errors.Is(err, drivermongo.ErrNoDocuments) {
		t.Fatalf("实体条件查询不存在文档应返回 mongo.ErrNoDocuments，实际为 %v", err)
	}
	if _, err := teacherRepo.QueryOneWithOptions(bson.M{"name": "cloud_database_missing_record"}); !errors.Is(err, drivermongo.ErrNoDocuments) {
		t.Fatalf("原生查询不存在文档应返回 mongo.ErrNoDocuments，实际为 %v", err)
	}
	if _, err := teacherRepo.QueryPageByCond(databasecloudmongo.NewPageQuery(Teacher{}, 0, 10)); !errors.Is(err, mongostarter.ErrInvalidPage) {
		t.Fatalf("非法实体分页参数应返回 ErrInvalidPage，实际为 %v", err)
	}
	if _, err := teacherRepo.QueryPageByBSON(databasecloudmongo.NewBSONPageQuery(bson.M{}, 1, 0)); !errors.Is(err, mongostarter.ErrInvalidPage) {
		t.Fatalf("非法 BSON 分页参数应返回 ErrInvalidPage，实际为 %v", err)
	}
	if _, err := teacherRepo.QueryPageWithOptions(databasecloudmongo.NewFilterPageQuery(bson.M{}, -1, 10)); !errors.Is(err, mongostarter.ErrInvalidPage) {
		t.Fatalf("非法原生分页参数应返回 ErrInvalidPage，实际为 %v", err)
	}
}
