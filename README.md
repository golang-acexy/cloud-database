# cloud-database

`cloud-database` provides repository-layer abstractions for the golang-acexy cloud ecosystem. It adapts `starter-gorm` and `starter-mongo` mappers to business-oriented repository APIs while preserving access to the underlying mapper and native database capabilities.

Repository methods use business verbs such as `Save`, `Query`, `Modify`, and `Remove`. Mapper methods remain aligned with database operations such as `Insert`, `Select`, `Update`, and `Delete`.

## Requirements

Current module Go version: `1.25.8`.

```bash
go get github.com/golang-acexy/cloud-database
```

Use the matching starter for the selected database:

```bash
go get github.com/golang-acexy/starter-gorm
go get github.com/golang-acexy/starter-mongo
```

The database starter must be running before repository methods are used. `cloud-database` is a code package and does not own database lifecycle.

## Shared Pagination

RDS and Mongo repositories use the same page result:

```go
type Pager[T any] struct {
	Records []*T `json:"records"`
	Total   int64 `json:"total"`
	Size    int   `json:"size"`
	Number  int   `json:"number"`
}
```

Page numbers start at 1. Callers are responsible for validating page number and size before invoking a repository.

## RDS Repository

The RDS package wraps `starter-gorm` and supports MySQL or PostgreSQL according to the model and starter configuration.

### Model and Mapper

Define a GORM model and embed `gormstarter.BaseMapper[T]` in the concrete mapper:

```go
type Teacher struct {
	ID   uint64 `gorm:"primaryKey"`
	Name string
	Age  uint
}

func (Teacher) TableName() string {
	return "teacher"
}

type TeacherMapper struct {
	gormstarter.BaseMapper[Teacher]
}

func (m TeacherMapper) WithTxMapper(tx *gorm.DB) TeacherMapper {
	return TeacherMapper{
		BaseMapper: m.GetBaseMapperWithTx(tx),
	}
}
```

`WithTxMapper` preserves the concrete mapper type when a transaction repository is created.

### Business Repository

Embed the generic RDS repository and provide a factory that rebuilds the concrete business repository:

```go
type TeacherRepo struct {
	rds.Repository[TeacherRepo, TeacherMapper, Teacher]
}

func NewTeacherRepo() TeacherRepo {
	base := rds.NewRepository[TeacherRepo](
		TeacherMapper{},
		func(repository rds.Repository[TeacherRepo, TeacherMapper, Teacher]) TeacherRepo {
			return TeacherRepo{Repository: repository}
		},
	)
	return TeacherRepo{Repository: base}
}
```

The self type `TeacherRepo` ensures transaction methods return the concrete business repository instead of only the embedded generic repository. Custom business methods remain available inside transactions.

### Save Operations

```go
rows, err := repo.Save(&Teacher{Name: "Alice", Age: 30})

rows, err = repo.SaveWithoutZeroFields(&Teacher{Name: "Bob"})

rows, err = repo.SaveWithMap(map[string]any{
	"name": "Carol",
	"age":  28,
})

rows, err = repo.SaveBatch([]*Teacher{
	{Name: "Dave"},
	{Name: "Eve"},
})
```

`Save` includes zero values unless excluded explicitly. `SaveWithoutZeroFields` omits zero-value fields. `SaveOrModifyByPrimaryKey` performs insert-or-update behavior through the underlying mapper.

### Query Operations

Typed conditions ignore zero-value fields:

```go
var teacher Teacher
rows, err := repo.QueryOneByCond(
	&Teacher{Name: "Alice"},
	&teacher,
)
```

Use Map conditions when zero is a meaningful query value:

```go
var teachers []*Teacher
rows, err := repo.QueryByMap(
	map[string]any{"age": 0},
	"id desc",
	&teachers,
)
```

Other query forms include:

- `QueryByID` and `QueryByIDs`
- `ExistsByID`
- `QueryOneByWhere` and `QueryByWhere`
- `QueryOneByGORM` and `QueryByGORM`
- `CountByCond`, `CountByMap`, `CountByWhere`, and `CountByGORM`

Raw GORM callbacks receive the current session and mutate that session directly:

```go
rows, err := repo.QueryByGORM(&teachers, func(db *gorm.DB) {
	db.Where("age >= ?", 18).Order("id desc")
})
```

### Pagination

```go
pager := databasecloud.Pager[Teacher]{
	Number: 1,
	Size:   20,
}

err := repo.QueryPageByMap(
	map[string]any{"age": 18},
	rds.PageQuery{
		OrderBySQL:     "id desc",
		SpecifyColumns: []string{"id", "name", "age"},
	},
	&pager,
)
```

RDS pagination supports typed conditions, Map conditions, raw Where SQL, and custom GORM callbacks. Count and page queries use separate GORM sessions in the underlying mapper.

### Modify and Remove

```go
rows, err := repo.ModifyByIDWithMap(
	map[string]any{"name": "Updated"},
	teacherID,
)

rows, err = repo.ModifyByMap(
	map[string]any{"age": 20},
	map[string]any{"class_no": 1},
)

rows, err = repo.RemoveByIDs([]any{1, 2, 3})
```

Use Map variants when zero-value updates or conditions must be preserved. Repository methods return the affected row count and the underlying error.

### Transactions

Use `Transaction` for automatic commit and rollback:

```go
err := repo.Transaction(func(txRepo TeacherRepo) error {
	if _, err := txRepo.Save(&Teacher{Name: "Alice"}); err != nil {
		return err
	}
	_, err := txRepo.ModifyByIDWithMap(
		map[string]any{"age": 31},
		teacherID,
	)
	return err
})
```

Behavior:

- A nil callback returns `ErrNilTransactionFunc`.
- Returning nil commits the transaction.
- Returning an error rolls back and returns the error.
- A panic rolls back and is rethrown.
- Commit and rollback errors are preserved.

Use `NewTxRepo` when transaction lifecycle must be controlled manually:

```go
txRepo := repo.NewTxRepo()
tx, err := txRepo.CurrentGORMDB()
if err != nil {
	return err
}

if _, err = txRepo.Save(&Teacher{Name: "Alice"}); err != nil {
	tx.Rollback()
	return err
}
return tx.Commit().Error
```

Use `WithTxRepo(tx)` to bind an existing `*gorm.DB` transaction. Creating another transaction repository from an already transaction-bound repository logs an error warning but does not stop the operation.

### Raw Access

```go
mapper := repo.RawMapper()
db, err := repo.CurrentGORMDB()
tableDB, err := repo.TableGORMDB()
```

Prefer repository methods for normal business code. Raw access is intended for integrations and queries not represented by the common API.

## Mongo Repository

The Mongo package wraps a concrete `starter-mongo` mapper. It does not require the RDS self-type factory because Mongo transactions are not part of this repository abstraction.

### Model, Mapper, and Repository

```go
type Teacher struct {
	ID   string `bson:"_id,omitempty" json:"id"`
	Name string `bson:"name" json:"name"`
	Age  uint   `bson:"age,omitempty" json:"age"`
}

func (Teacher) CollectionName() string {
	return "teacher"
}

type TeacherMapper struct {
	mongostarter.BaseMapper[Teacher]
}

type TeacherRepo struct {
	mongo.Repository[TeacherMapper, Teacher]
}

func NewTeacherRepo() TeacherRepo {
	return TeacherRepo{
		Repository: mongo.NewRepository[TeacherMapper, Teacher](TeacherMapper{}),
	}
}
```

### Save Operations

Mongo save methods return inserted ObjectID strings:

```go
id, err := repo.Save(&Teacher{Name: "Alice", Age: 30})

id, err = repo.SaveWithBSON(bson.M{
	"name": "Bob",
	"age":  28,
})

ids, err := repo.SaveBatch([]*Teacher{
	{Name: "Carol"},
	{Name: "Dave"},
})
```

Every single and batch save operation also has BSON or native Options variants.

### Query and Pagination

Mongo queries are available in three forms:

- `ByCond` for typed model conditions
- `ByBSON` for `bson.M` filters
- `WithOptions` for native driver filters and options

```go
var teachers []*Teacher
err := repo.QueryByBSON(
	bson.M{"age": bson.M{"$gte": 18}},
	mongostarter.NewOrderBy("age", true),
	&teachers,
)
```

Pagination uses the shared result and Mongo-specific query options:

```go
pager := databasecloud.Pager[Teacher]{Number: 1, Size: 20}

err := repo.QueryPageByBSON(
	bson.M{"age": bson.M{"$gte": 18}},
	mongo.PageQuery{
		OrderBy:        mongostarter.NewOrderBy("age", true),
		SpecifyColumns: []string{"name", "age"},
	},
	&pager,
)
```

`PageQuery` also accepts native `FindOptions` and `CountOptions`.

### ID Handling

String IDs are interpreted as MongoDB ObjectID hex strings by default:

```go
var teacher Teacher
err := repo.QueryByID("507f1f77bcf86cd799439011", &teacher)
```

Pass `true` when the collection uses ordinary string IDs:

```go
err := repo.QueryByID("teacher-1001", &teacher, true)
```

The same option applies to ID query, update, existence, and remove methods.

### Modify and Remove

```go
rows, err := repo.ModifyByIDWithBSON(
	bson.M{"age": 31},
	teacherID,
)

rows, err = repo.ModifyWithOptions(
	bson.M{"status": "pending"},
	bson.M{"$set": bson.M{"status": "active"}},
)

rows, err = repo.RemoveByBSON(bson.M{"status": "inactive"})
```

Single and multi-document variants are explicit: `ModifyOne*`/`Modify*` and `RemoveOne*`/`Remove*`.

Empty update and delete conditions are rejected by `starter-mongo` to prevent accidental full-collection operations.

### Raw Access

```go
mapper := repo.RawMapper()
collection, err := repo.Collection()
```

## Lifecycle

Start `starter-gorm` or `starter-mongo` through the parent loader before constructing application traffic that uses repositories:

```go
loader := parent.InitStarterLoader([]parent.Starter{
	&gormstarter.GormStarter{Config: gormConfig},
	&mongostarter.MongoStarter{Config: mongoConfig},
})

if err := loader.Start(); err != nil {
	panic(err)
}
```

Stop starters through the same loader during application shutdown:

```go
_, err := loader.StopAllBySetting(10 * time.Second)
if err != nil {
	panic(err)
}
```

## Important Errors

RDS repository construction and transaction helpers define these errors:

| Error | Meaning |
| --- | --- |
| `rds.ErrNilRepositoryFactory` | `NewRepository` received a nil concrete repository factory. |
| `rds.ErrRepositoryNotInitialized` | A repository requiring its factory was not initialized through `NewRepository`. |
| `rds.ErrNilTransactionFunc` | `Transaction` received a nil callback. |

Mapper and starter errors are returned without being hidden. Use `errors.Is` for exported sentinel errors.

## Testing

Run tests from the module directory:

```bash
GOMODCACHE=/Users/acexy/Repository/cache/golang go test ./...
```

The RDS and Mongo integration suites require the local database configurations defined in their test packages.
