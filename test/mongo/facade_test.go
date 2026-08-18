package mongo

import (
	"reflect"
	"testing"

	databasecloudmongo "github.com/golang-acexy/cloud-database/databasecloud/mongo"
	"github.com/golang-acexy/starter-mongo/mongostarter"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestQueryFacade(t *testing.T) {
	orderBy := mongostarter.OrderBy{Column: "created_at", Desc: true}

	cond := databasecloudmongo.NewCondQuery(Teacher{Sex: 1}).WithOrderBy(orderBy).Select("name").WithLimit(10)
	if cond.Condition.Sex != 1 || cond.Limit != 10 || len(cond.OrderBy) != 1 || !reflect.DeepEqual(cond.SelectColumns, []string{"name"}) {
		t.Fatalf("CondQuery 门面转发异常: %+v", cond)
	}

	bsonQuery := databasecloudmongo.NewBSONQuery(bson.M{"sex": 0}).WithOrderBy(orderBy).Select("name").WithLimit(5)
	if bsonQuery.Condition["sex"] != 0 || bsonQuery.Limit != 5 || len(bsonQuery.OrderBy) != 1 {
		t.Fatalf("BSONQuery 门面转发异常: %+v", bsonQuery)
	}

	findOption := options.Find().SetAllowDiskUse(true)
	countOption := options.Count().SetHint("sex_1")
	page := databasecloudmongo.NewPageQuery(Teacher{Sex: 1}, 2, 20).
		WithOrderBy(orderBy).
		Select("name").
		WithFindOptions(findOption).
		WithCountOptions(countOption)
	bsonPage := databasecloudmongo.NewBSONPageQuery(bson.M{"sex": 0}, 3, 30).WithOrderBy(orderBy)
	filterPage := databasecloudmongo.NewFilterPageQuery(bson.M{"sex": 1}, 4, 40).Select("name")
	if page.Number != 2 || page.Size != 20 || len(page.FindOptions) != 1 || len(page.CountOptions) != 1 || bsonPage.Number != 3 || bsonPage.Size != 30 || filterPage.Number != 4 || filterPage.Size != 40 {
		t.Fatalf("分页 Query 门面转发异常: page=%+v bson=%+v filter=%+v", page, bsonPage, filterPage)
	}
}
