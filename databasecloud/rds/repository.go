package rds

import (
	"database/sql"
	"errors"

	"github.com/acexy/golang-toolkit/logger"
	"github.com/golang-acexy/cloud-database/databasecloud"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"gorm.io/gorm"
)

// Mapper combines starter operations with concrete-type transaction copying.
type Mapper[M any, T gormstarter.Model] interface {
	gormstarter.RawMapper
	gormstarter.QueryMapper[T]
	gormstarter.InsertMapper[T]
	gormstarter.UpdateMapper[T]
	gormstarter.DeleteMapper[T]
	// Wrapper creates a query builder.
	Wrapper() *QueryWrapper[T]
	// PageWrapper creates a page builder.
	PageWrapper(number, size int) *PageWrapper[T]
	// UpdateWrapper creates an update builder.
	UpdateWrapper() *ModifyWrapper[T]
	// WithTxMapper copies the mapper onto the supplied transaction.
	WithTxMapper(tx *gorm.DB) M
}

// Repository allocates caller-owned query results and delegates persistence to its mapper.
// Initialize with NewRepository; concurrency depends on the mapper and borrowed inputs.
// Database pools and clients remain starter-owned.
type Repository[R any, M Mapper[M, T], T gormstarter.Model] struct {
	mapper  M
	factory func(Repository[R, M, T]) R
}

// QueryOptions aliases the starter type; referenced inputs retain its ownership and mutation rules.
type QueryOptions = gormstarter.QueryOptions

// PageOptions aliases the starter type; referenced inputs retain its ownership and mutation rules.
type PageOptions = gormstarter.PageOptions

// TimeRange aliases the starter type; referenced inputs retain its ownership and mutation rules.
type TimeRange = gormstarter.TimeRange

// CondQuery aliases the starter type; referenced inputs retain its ownership and mutation rules.
type CondQuery[T gormstarter.Model] = gormstarter.CondQuery[T]

// MapQuery aliases the starter type; referenced inputs retain its ownership and mutation rules.
type MapQuery = gormstarter.MapQuery

// WhereQuery aliases the starter type; referenced inputs retain its ownership and mutation rules.
type WhereQuery = gormstarter.WhereQuery

// PageQuery aliases the starter type; referenced inputs retain its ownership and mutation rules.
type PageQuery[T gormstarter.Model] = gormstarter.PageQuery[T]

// MapPageQuery aliases the starter type; referenced inputs retain its ownership and mutation rules.
type MapPageQuery = gormstarter.MapPageQuery

// WherePageQuery aliases the starter type; referenced inputs retain its ownership and mutation rules.
type WherePageQuery = gormstarter.WherePageQuery

// QueryWrapper aliases the starter type; referenced inputs retain its ownership and mutation rules.
type QueryWrapper[T gormstarter.Model] = gormstarter.QueryWrapper[T]

// PageWrapper aliases the starter type; referenced inputs retain its ownership and mutation rules.
type PageWrapper[T gormstarter.Model] = gormstarter.PageWrapper[T]

// ModifyWrapper aliases the starter type; referenced inputs retain its ownership and mutation rules.
type ModifyWrapper[T gormstarter.Model] = gormstarter.UpdateWrapper[T]

// GormPageQuery combines callbacks with output metadata; pagination is not applied automatically.
type GormPageQuery struct {
	// CountRawDB builds count conditions and must be non-nil.
	CountRawDB func(*gorm.DB)
	// PageRawDB builds data conditions and limit/offset; required when count is positive.
	PageRawDB func(*gorm.DB)
	// PageOptions provides output Number and Size; other options are ignored in this callback path.
	PageOptions
}

// NewRepository stores the mapper and concrete factory; nil factories panic with ErrNilRepositoryFactory.
func NewRepository[R any, M Mapper[M, T], T gormstarter.Model](mapper M, factory func(Repository[R, M, T]) R) Repository[R, M, T] {
	if factory == nil {
		panic(ErrNilRepositoryFactory)
	}
	return Repository[R, M, T]{mapper: mapper, factory: factory}
}

// NewTxRepo begins a caller-owned transaction; check CurrentGormDB().Error immediately.
// Requires an initialized repository and available database; unavailable routing may panic.
// Existing transactions are logged but not rejected.
func (r Repository[R, M, T]) NewTxRepo(opts ...*sql.TxOptions) R {
	db := r.mapper.CurrentGormDB()
	warnTransactionRepository(db, "NewTxRepo")
	tx := db.Begin(opts...)
	return r.repositoryWithTx(tx)
}

// WithTxRepo copies an initialized repository onto a caller-owned transaction; nil restores ordinary routing with the standard mapper.
func (r Repository[R, M, T]) WithTxRepo(tx *gorm.DB) R {
	warnTransactionRepository(r.mapper.CurrentGormDB(), "WithTxRepo")
	return r.repositoryWithTx(tx)
}

func warnTransactionRepository(db *gorm.DB, operation string) {
	if db == nil || db.Statement == nil {
		return
	}
	if _, ok := db.Statement.ConnPool.(interface {
		Commit() error
		Rollback() error
	}); ok {
		logger.Logrus().WithField("operation", operation).Errorln("transaction repository creates another transaction repository")
	}
}

// Transaction returns ErrNilTransactionFunc for nil callbacks and ErrRepositoryNotInitialized for missing factories.
// It commits on callback success; callback errors trigger rollback and are joined with rollback errors.
// Panics trigger rollback and are rethrown, discarding rollback errors; commit errors return without another rollback.
// Requires an available database and initialized factory; use the callback repository for transactional work. Nested transactions are unsupported.
func (r Repository[R, M, T]) Transaction(fn func(R) error, opts ...*sql.TxOptions) (err error) {
	if fn == nil {
		return ErrNilTransactionFunc
	}
	if r.factory == nil {
		return ErrRepositoryNotInitialized
	}
	db := r.mapper.CurrentGormDB()
	tx := db.Begin(opts...)
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			tx.Rollback()
			panic(recovered)
		}
	}()

	if err = fn(r.repositoryWithTx(tx)); err != nil {
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			return errors.Join(err, rollbackErr)
		}
		return err
	}
	return tx.Commit().Error
}

func (r Repository[R, M, T]) repositoryWithTx(tx *gorm.DB) R {
	txRepository := r
	txRepository.mapper = r.mapper.WithTxMapper(tx)
	return r.factory(txRepository)
}

// RawMapper returns the concrete mapper; underlying resources remain starter-owned.
func (r Repository[R, M, T]) RawMapper() M {
	return r.mapper
}

// CurrentGormDB returns the bound transaction or shared routed DB, possibly nil; it does not create an isolated session.
func (r Repository[R, M, T]) CurrentGormDB() *gorm.DB {
	return r.mapper.CurrentGormDB()
}

// TableGormDB returns a table-scoped DB, possibly nil on routing failure; do not close its shared pool.
func (r Repository[R, M, T]) TableGormDB() *gorm.DB {
	return r.mapper.TableGormDB()
}

// Wrapper creates a mutable model-specific query builder.
func (r Repository[R, M, T]) Wrapper() *QueryWrapper[T] {
	return r.mapper.Wrapper()
}

// PageWrapper creates a mutable page builder; validation occurs during execution.
func (r Repository[R, M, T]) PageWrapper(number, size int) *PageWrapper[T] {
	return r.mapper.PageWrapper(number, size)
}

// ModifyWrapper creates a mutable update builder requiring predicates and assignments.
func (r Repository[R, M, T]) ModifyWrapper() *ModifyWrapper[T] {
	return r.mapper.UpdateWrapper()
}

// Save inserts using GORM Create semantics, including model defaults and optional omitted columns.
func (r Repository[R, M, T]) Save(entity *T, excludeColumns ...string) (int64, error) {
	return r.mapper.Insert(entity, excludeColumns...)
}

// SaveWithoutZeroFields inserts selected nonzero fields; empty effective data is rejected.
func (r Repository[R, M, T]) SaveWithoutZeroFields(entity *T) (int64, error) {
	return r.mapper.InsertWithoutZeroFields(entity)
}

// SaveWithMap inserts explicit column assignments, preserving zeros.
func (r Repository[R, M, T]) SaveWithMap(entity map[string]any) (int64, error) {
	return r.mapper.InsertWithMap(entity)
}

// SaveOrModifyByPrimaryKey delegates to GORM Save; this is not an arbitrary-conflict upsert.
func (r Repository[R, M, T]) SaveOrModifyByPrimaryKey(entity *T, excludeColumns ...string) (int64, error) {
	return r.mapper.InsertOrUpdateByPrimaryKey(entity, excludeColumns...)
}

// SaveBatch inserts entities using GORM Create semantics; generated values may mutate the inputs.
func (r Repository[R, M, T]) SaveBatch(entities []*T, excludeColumns ...string) (int64, error) {
	return r.mapper.InsertBatch(entities, excludeColumns...)
}

// QueryByID allocates one result; zero rows return nil, nil and mapper errors are preserved.
func (r Repository[R, M, T]) QueryByID(id any) (*T, error) {
	result := new(T)
	rows, err := r.mapper.SelectByID(id, result)
	if err != nil || rows == 0 {
		return nil, err
	}
	return result, nil
}

// QueryByIDs queries an ID batch into a new slice; empty input is rejected and input ordering is not guaranteed.
func (r Repository[R, M, T]) QueryByIDs(ids []any) ([]*T, error) {
	result := make([]*T, 0)
	_, err := r.mapper.SelectByIDs(ids, &result)
	return result, err
}

// ExistsByID checks existence and returns database errors separately from absence.
func (r Repository[R, M, T]) ExistsByID(id any) (bool, error) {
	return r.mapper.ExistsByID(id)
}

// QueryOneByCond queries nonzero struct conditions; no rows return nil, nil.
func (r Repository[R, M, T]) QueryOneByCond(query CondQuery[T]) (*T, error) {
	result := new(T)
	rows, err := r.mapper.SelectOneByCond(query, result)
	if err != nil || rows == 0 {
		return nil, err
	}
	return result, nil
}

// QueryByCond queries nonzero struct conditions; zero Limit is unrestricted and errors may accompany partial results.
func (r Repository[R, M, T]) QueryByCond(query CondQuery[T]) ([]*T, error) {
	result := make([]*T, 0)
	_, err := r.mapper.SelectByCond(query, &result)
	return result, err
}

// QueryOneByMap queries column conditions including explicit zeros; no rows return nil, nil.
func (r Repository[R, M, T]) QueryOneByMap(query MapQuery) (*T, error) {
	result := new(T)
	rows, err := r.mapper.SelectOneByMap(query, result)
	if err != nil || rows == 0 {
		return nil, err
	}
	return result, nil
}

// QueryByMap queries column conditions including explicit zeros; zero Limit is unrestricted.
func (r Repository[R, M, T]) QueryByMap(query MapQuery) ([]*T, error) {
	result := make([]*T, 0)
	_, err := r.mapper.SelectByMap(query, &result)
	return result, err
}

// QueryOneByWhere queries trusted SQL with bound arguments; no rows return nil, nil and LIMIT 1 is not added.
func (r Repository[R, M, T]) QueryOneByWhere(query WhereQuery) (*T, error) {
	result := new(T)
	rows, err := r.mapper.SelectOneByWhere(query, result)
	if err != nil || rows == 0 {
		return nil, err
	}
	return result, nil
}

// QueryByWhere queries trusted SQL with bound arguments into a new slice; zero Limit is unrestricted.
func (r Repository[R, M, T]) QueryByWhere(query WhereQuery) ([]*T, error) {
	result := make([]*T, 0)
	_, err := r.mapper.SelectByWhere(query, &result)
	return result, err
}

// QueryOneByGorm allocates one result; zero rows return nil, nil and mapper errors are preserved.
// Supply a non-nil callback that mutates the provided session; nil panics when invoked.
func (r Repository[R, M, T]) QueryOneByGorm(rawDB func(*gorm.DB)) (*T, error) {
	result := new(T)
	rows, err := r.mapper.SelectOneByGorm(result, rawDB)
	if err != nil || rows == 0 {
		return nil, err
	}
	return result, nil
}

// QueryByGorm allocates a result slice; errors may accompany partially populated results.
// Supply a non-nil callback that mutates the provided session; nil panics when invoked.
func (r Repository[R, M, T]) QueryByGorm(rawDB func(*gorm.DB)) ([]*T, error) {
	result := make([]*T, 0)
	_, err := r.mapper.SelectByGorm(&result, rawDB)
	return result, err
}

// QueryOneByWrapper allocates one result; zero rows return nil, nil and mapper errors are preserved.
func (r Repository[R, M, T]) QueryOneByWrapper(query *QueryWrapper[T]) (*T, error) {
	result := new(T)
	rows, err := r.mapper.SelectOneByWrapper(query, result)
	if err != nil || rows == 0 {
		return nil, err
	}
	return result, nil
}

// QueryByWrapper allocates a result slice; errors may accompany partially populated results.
func (r Repository[R, M, T]) QueryByWrapper(query *QueryWrapper[T]) ([]*T, error) {
	result := make([]*T, 0)
	_, err := r.mapper.SelectByWrapper(query, &result)
	return result, err
}

// QueryPageByCond allocates page output and delegates separate count/data operations without an automatic snapshot.
// Records may be nil; check errors before using partial records or metadata.
func (r Repository[R, M, T]) QueryPageByCond(query PageQuery[T]) (databasecloud.Pager[T], error) {
	pager := databasecloud.Pager[T]{Number: query.Number, Size: query.Size}
	total, err := r.mapper.SelectPageByCond(query, &pager.Records)
	if err != nil {
		return pager, err
	}
	pager.Total = total
	return pager, nil
}

// QueryPageByMap allocates page output and delegates separate count/data operations without an automatic snapshot.
// Records may be nil; check errors before using partial records or metadata.
func (r Repository[R, M, T]) QueryPageByMap(query MapPageQuery) (databasecloud.Pager[T], error) {
	pager := databasecloud.Pager[T]{Number: query.Number, Size: query.Size}
	total, err := r.mapper.SelectPageByMap(query, &pager.Records)
	if err != nil {
		return pager, err
	}
	pager.Total = total
	return pager, nil
}

// QueryPageByWhere allocates page output and delegates separate count/data operations without an automatic snapshot.
// Records may be nil; check errors before using partial records or metadata.
func (r Repository[R, M, T]) QueryPageByWhere(query WherePageQuery) (databasecloud.Pager[T], error) {
	pager := databasecloud.Pager[T]{Number: query.Number, Size: query.Size}
	total, err := r.mapper.SelectPageByWhere(query, &pager.Records)
	if err != nil {
		return pager, err
	}
	pager.Total = total
	return pager, nil
}

// QueryPageByGorm executes count and data callbacks; metadata does not apply or validate pagination.
// Callbacks must be non-nil when invoked and supply matching filters and data limit/offset.
func (r Repository[R, M, T]) QueryPageByGorm(query GormPageQuery) (databasecloud.Pager[T], error) {
	pager := databasecloud.Pager[T]{Number: query.Number, Size: query.Size}
	total, err := r.mapper.SelectPageByGorm(query.CountRawDB, query.PageRawDB, &pager.Records)
	pager.Total = total
	return pager, err
}

// QueryPageByWrapper allocates page output and delegates separate count/data operations without an automatic snapshot.
// Records may be nil; check errors before using partial records or metadata.
func (r Repository[R, M, T]) QueryPageByWrapper(query *PageWrapper[T]) (databasecloud.Pager[T], error) {
	pager := databasecloud.Pager[T]{}
	if query != nil {
		pager.Number = query.Number()
		pager.Size = query.Size()
	}
	total, err := r.mapper.SelectPageByWrapper(query, &pager.Records)
	pager.Total = total
	return pager, err
}

// CountByCond counts nonzero struct predicates and time ranges; list projection, order, and limit are ignored.
func (r Repository[R, M, T]) CountByCond(query CondQuery[T]) (int64, error) {
	return r.mapper.CountByCond(query)
}

// CountByMap counts column predicates including zeros; list projection, order, and limit are ignored.
func (r Repository[R, M, T]) CountByMap(query MapQuery) (int64, error) {
	return r.mapper.CountByMap(query)
}

// CountByWhere counts trusted SQL predicates with bound arguments; list limits are ignored.
func (r Repository[R, M, T]) CountByWhere(query WhereQuery) (int64, error) {
	return r.mapper.CountByWhere(query)
}

// CountByGorm delegates counting to the mapper and preserves errors.
// Supply a non-nil callback that mutates the provided session; nil panics when invoked.
func (r Repository[R, M, T]) CountByGorm(rawDB func(*gorm.DB)) (int64, error) {
	return r.mapper.CountByGorm(rawDB)
}

// CountByWrapper counts wrapper predicates while ignoring projection, order, limit, and offset.
func (r Repository[R, M, T]) CountByWrapper(query *QueryWrapper[T]) (int64, error) {
	return r.mapper.CountByWrapper(query)
}

// ModifyByID updates an ID; zero struct fields are omitted unless explicitly selected by updateColumns.
func (r Repository[R, M, T]) ModifyByID(updated *T, id any, updateColumns ...string) (int64, error) {
	return r.mapper.UpdateByID(updated, id, updateColumns...)
}

// ModifyByIDWithoutZeroFields updates selected nonzero fields plus explicitly included zero columns.
func (r Repository[R, M, T]) ModifyByIDWithoutZeroFields(updated *T, id any, includeZeroFieldColumns ...string) (int64, error) {
	return r.mapper.UpdateByIDWithoutZeroFields(updated, id, includeZeroFieldColumns...)
}

// ModifyByIDWithMap updates an ID with explicit column assignments, including zero and nil values.
func (r Repository[R, M, T]) ModifyByIDWithMap(updated map[string]any, id any) (int64, error) {
	return r.mapper.UpdateByIDWithMap(updated, id)
}

// ModifyByCond updates nonzero struct conditions; explicit updateColumns can select zero assignments.
func (r Repository[R, M, T]) ModifyByCond(updated *T, condition T, updateColumns ...string) (int64, error) {
	return r.mapper.UpdateByCond(updated, condition, updateColumns...)
}

// ModifyByCondWithZeroFields updates nonzero struct conditions using nonzero fields plus selected zero columns.
func (r Repository[R, M, T]) ModifyByCondWithZeroFields(updated *T, condition T, includeZeroFieldColumns ...string) (int64, error) {
	return r.mapper.UpdateByCondWithZeroFields(updated, condition, includeZeroFieldColumns...)
}

// ModifyByMap updates nonempty map conditions with explicit column assignments, including zero and nil values.
func (r Repository[R, M, T]) ModifyByMap(updated, condition map[string]any) (int64, error) {
	return r.mapper.UpdateByMap(updated, condition)
}

// ModifyByWhere updates nonzero struct fields using trusted SQL and bound arguments; blank conditions are rejected.
func (r Repository[R, M, T]) ModifyByWhere(updated *T, rawWhereSQL string, args ...any) (int64, error) {
	return r.mapper.UpdateByWhere(updated, rawWhereSQL, args...)
}

// ModifyByWrapper updates explicit assignments; empty predicates or assignments are rejected.
func (r Repository[R, M, T]) ModifyByWrapper(wrapper *ModifyWrapper[T]) (int64, error) {
	return r.mapper.UpdateByWrapper(wrapper)
}

// RemoveByID uses GORM inline-condition deletion.
// Nonnumeric strings may become SQL clauses; use RemoveByMap or bound RemoveByWhere for string keys.
func (r Repository[R, M, T]) RemoveByID(id any) (int64, error) {
	return r.mapper.DeleteByID(id)
}

// RemoveByIDs deletes an ID batch; empty input is rejected but elements are not individually validated.
func (r Repository[R, M, T]) RemoveByIDs(ids []any) (int64, error) {
	return r.mapper.DeleteByIDs(ids)
}

// RemoveByCond delegates deletion with mapper empty-condition guards; struct conditions omit zero fields.
func (r Repository[R, M, T]) RemoveByCond(condition T) (int64, error) {
	return r.mapper.DeleteByCond(condition)
}

// RemoveByMap deletes nonempty column conditions, including explicit zero values.
func (r Repository[R, M, T]) RemoveByMap(condition map[string]any) (int64, error) {
	return r.mapper.DeleteByMap(condition)
}

// RemoveByWhere deletes using trusted SQL and bound arguments; blank conditions are rejected.
func (r Repository[R, M, T]) RemoveByWhere(rawWhereSQL string, args ...any) (int64, error) {
	return r.mapper.DeleteByWhere(rawWhereSQL, args...)
}
