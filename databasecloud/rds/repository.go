package rds

import (
	"database/sql"
	"errors"

	"github.com/acexy/golang-toolkit/logger"
	"github.com/golang-acexy/cloud-database/databasecloud"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"gorm.io/gorm"
)

// Mapper 组合 GORM Mapper 的各项基础能力，并约束事务复制能力。
// 直接展开各项接口可避免嵌套泛型接口给跨包类型检查带来的额外复杂度。
type Mapper[M any, T gormstarter.Model] interface {
	gormstarter.RawMapper
	gormstarter.QueryMapper[T]
	gormstarter.InsertMapper[T]
	gormstarter.UpdateMapper[T]
	gormstarter.DeleteMapper[T]
	Wrapper() *QueryWrapper[T]
	PageWrapper(number, size int) *PageWrapper[T]
	UpdateWrapper() *UpdateWrapper[T]
	WithTxMapper(tx *gorm.DB) M
}

// Repository 组合具体 Mapper，并通过 R 保留业务 Repository 类型。
type Repository[R any, M Mapper[M, T], T gormstarter.Model] struct {
	mapper  M
	factory func(Repository[R, M, T]) R
}

// QueryOptions 是 GORM 查询选项在 RDS Repository 层的门面别名。
type QueryOptions = gormstarter.QueryOptions

// PageOptions 是 GORM 分页选项在 RDS Repository 层的门面别名。
type PageOptions = gormstarter.PageOptions

// TimeRange 是 GORM 时间范围条件在 RDS Repository 层的门面别名。
type TimeRange = gormstarter.TimeRange

// CondQuery 是 GORM 实体条件查询在 RDS Repository 层的门面别名。
type CondQuery[T gormstarter.Model] = gormstarter.CondQuery[T]

// MapQuery 是 GORM Map 条件查询在 RDS Repository 层的门面别名。
type MapQuery = gormstarter.MapQuery

// WhereQuery 是 GORM 原始 SQL 条件查询在 RDS Repository 层的门面别名。
type WhereQuery = gormstarter.WhereQuery

// PageQuery 是 GORM 实体条件分页查询在 RDS Repository 层的门面别名。
type PageQuery[T gormstarter.Model] = gormstarter.PageQuery[T]

// MapPageQuery 是 GORM Map 条件分页查询在 RDS Repository 层的门面别名。
type MapPageQuery = gormstarter.MapPageQuery

// WherePageQuery 是 GORM 原始 SQL 条件分页查询在 RDS Repository 层的门面别名。
type WherePageQuery = gormstarter.WherePageQuery

// QueryWrapper 是 GORM 类型安全查询 Wrapper 在 RDS Repository 层的门面别名。
type QueryWrapper[T gormstarter.Model] = gormstarter.QueryWrapper[T]

// PageWrapper 是 GORM 类型安全分页 Wrapper 在 RDS Repository 层的门面别名。
type PageWrapper[T gormstarter.Model] = gormstarter.PageWrapper[T]

// UpdateWrapper 是 GORM 类型安全更新 Wrapper 在 RDS Repository 层的门面别名。
type UpdateWrapper[T gormstarter.Model] = gormstarter.UpdateWrapper[T]

// GormPageQuery 定义分别构建统计和分页数据的原始 GORM 查询。
type GormPageQuery struct {
	CountRawDB func(*gorm.DB)
	PageRawDB  func(*gorm.DB)
	PageOptions
}

// NewRepository 创建基础 Repository，并注册业务 Repository 工厂函数。
func NewRepository[R any, M Mapper[M, T], T gormstarter.Model](mapper M, factory func(Repository[R, M, T]) R) Repository[R, M, T] {
	if factory == nil {
		panic(ErrNilRepositoryFactory)
	}
	return Repository[R, M, T]{mapper: mapper, factory: factory}
}

// -------------------- 事务 --------------------

// NewTxRepo 创建绑定新事务的业务 Repository，调用方负责提交或回滚事务。
// 调用前必须完成 GORM Starter 和 Repository 初始化；事务启动错误由后续数据库操作返回。
// 若当前 Repository 已绑定事务，将记录错误日志但仍继续尝试创建新事务。
func (r Repository[R, M, T]) NewTxRepo(opts ...*sql.TxOptions) R {
	db := r.mapper.CurrentGormDB()
	warnTransactionRepository(db, "NewTxRepo")
	tx := db.Begin(opts...)
	return r.repositoryWithTx(tx)
}

// WithTxRepo 创建绑定指定事务的业务 Repository，调用方负责保证事务有效并管理其生命周期。
// 调用前必须通过 NewRepository 完成 Repository 初始化。
// 若当前 Repository 已绑定事务，将记录错误日志但仍使用传入事务创建新 Repository。
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

// Transaction 在新事务中执行业务函数，并自动提交；返回错误或发生 panic 时自动回滚。
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

// -------------------- 原生访问与 Wrapper 构造 --------------------

// RawMapper 获取具体 Mapper
func (r Repository[R, M, T]) RawMapper() M {
	return r.mapper
}

// CurrentGormDB 获取当前 Gorm DB，如果已有事务则返回该事务，否则获取新的 Gorm DB。
func (r Repository[R, M, T]) CurrentGormDB() *gorm.DB {
	return r.mapper.CurrentGormDB()
}

// TableGormDB 获取已经限定当前 Mapper 表名的 Gorm DB。
func (r Repository[R, M, T]) TableGormDB() *gorm.DB {
	return r.mapper.TableGormDB()
}

// Wrapper 创建与当前 Repository 模型类型绑定的查询 Wrapper。
func (r Repository[R, M, T]) Wrapper() *QueryWrapper[T] {
	return r.mapper.Wrapper()
}

// PageWrapper 创建与当前 Repository 模型类型绑定的分页查询 Wrapper。
func (r Repository[R, M, T]) PageWrapper(number, size int) *PageWrapper[T] {
	return r.mapper.PageWrapper(number, size)
}

// UpdateWrapper 创建与当前 Repository 模型类型绑定的更新 Wrapper。
func (r Repository[R, M, T]) UpdateWrapper() *UpdateWrapper[T] {
	return r.mapper.UpdateWrapper()
}

// -------------------- 新增 --------------------

// Save 保存数据 默认零值数据也会进行存储 可通过设置excludeColumns排除零值数据
func (r Repository[R, M, T]) Save(entity *T, excludeColumns ...string) (int64, error) {
	return r.mapper.Insert(entity, excludeColumns...)
}

// SaveWithoutZeroFields 保存数据，零值字段不参与保存。
func (r Repository[R, M, T]) SaveWithoutZeroFields(entity *T) (int64, error) {
	return r.mapper.InsertWithoutZeroFields(entity)
}

// SaveWithMap 通过Map类型保存数据 key为列明 value为列值
func (r Repository[R, M, T]) SaveWithMap(entity map[string]any) (int64, error) {
	return r.mapper.InsertWithMap(entity)
}

// SaveOrModifyByPrimaryKey 保存/更新数据 (主键冲突则执行更新) 零值也将参与保存
func (r Repository[R, M, T]) SaveOrModifyByPrimaryKey(entity *T, excludeColumns ...string) (int64, error) {
	return r.mapper.InsertOrUpdateByPrimaryKey(entity, excludeColumns...)
}

// SaveBatch 批量保存数据 默认零值数据也会进行存储 可通过设置excludeColumns排除零值数据
func (r Repository[R, M, T]) SaveBatch(entities []*T, excludeColumns ...string) (int64, error) {
	return r.mapper.InsertBatch(entities, excludeColumns...)
}

// -------------------- 查询 --------------------

// QueryByID 通过主键查询数据
func (r Repository[R, M, T]) QueryByID(id any) (*T, error) {
	result := new(T)
	rows, err := r.mapper.SelectByID(id, result)
	if err != nil || rows == 0 {
		return nil, err
	}
	return result, nil
}

// QueryByIDs 通过主键查询数据
func (r Repository[R, M, T]) QueryByIDs(ids []any) ([]*T, error) {
	result := make([]*T, 0)
	_, err := r.mapper.SelectByIDs(ids, &result)
	return result, err
}

// ExistsByID 判断指定主键的数据是否存在。
func (r Repository[R, M, T]) ExistsByID(id any) (bool, error) {
	return r.mapper.ExistsByID(id)
}

// QueryOneByCond 通过条件查询 查询条件零值字段将被自动忽略 specifyColumns 指定只需要查询的数据库字段
func (r Repository[R, M, T]) QueryOneByCond(query CondQuery[T]) (*T, error) {
	result := new(T)
	rows, err := r.mapper.SelectOneByCond(query, result)
	if err != nil || rows == 0 {
		return nil, err
	}
	return result, nil
}

// QueryByCond 通过条件查询 查询条件零值字段将被自动忽略 specifyColumns 指定只需要查询的数据库字段
func (r Repository[R, M, T]) QueryByCond(query CondQuery[T]) ([]*T, error) {
	result := make([]*T, 0)
	_, err := r.mapper.SelectByCond(query, &result)
	return result, err
}

// QueryOneByMap 通过指定字段与值查询数据 解决零值条件问题 specifyColumns 指定只需要查询的数据库字段
func (r Repository[R, M, T]) QueryOneByMap(query MapQuery) (*T, error) {
	result := new(T)
	rows, err := r.mapper.SelectOneByMap(query, result)
	if err != nil || rows == 0 {
		return nil, err
	}
	return result, nil
}

// QueryByMap 通过指定字段与值查询数据 解决零值条件问题 specifyColumns 指定只需要查询的数据库字段
func (r Repository[R, M, T]) QueryByMap(query MapQuery) ([]*T, error) {
	result := make([]*T, 0)
	_, err := r.mapper.SelectByMap(query, &result)
	return result, err
}

// QueryOneByWhere 通过原始Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql: "a = ?"  args = 1
func (r Repository[R, M, T]) QueryOneByWhere(query WhereQuery) (*T, error) {
	result := new(T)
	rows, err := r.mapper.SelectOneByWhere(query, result)
	if err != nil || rows == 0 {
		return nil, err
	}
	return result, nil
}

// QueryByWhere 通过原始Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSq: "a = ?" args = 1
func (r Repository[R, M, T]) QueryByWhere(query WhereQuery) ([]*T, error) {
	result := make([]*T, 0)
	_, err := r.mapper.SelectByWhere(query, &result)
	return result, err
}

// QueryOneByGorm 通过原始 Gorm 查询单条数据。
func (r Repository[R, M, T]) QueryOneByGorm(rawDB func(*gorm.DB)) (*T, error) {
	result := new(T)
	rows, err := r.mapper.SelectOneByGorm(result, rawDB)
	if err != nil || rows == 0 {
		return nil, err
	}
	return result, nil
}

// QueryByGorm 通过原始 Gorm 查询数据。
func (r Repository[R, M, T]) QueryByGorm(rawDB func(*gorm.DB)) ([]*T, error) {
	result := make([]*T, 0)
	_, err := r.mapper.SelectByGorm(&result, rawDB)
	return result, err
}

// QueryOneByWrapper 通过类型安全的 Wrapper 查询一条数据。
func (r Repository[R, M, T]) QueryOneByWrapper(query *QueryWrapper[T]) (*T, error) {
	result := new(T)
	rows, err := r.mapper.SelectOneByWrapper(query, result)
	if err != nil || rows == 0 {
		return nil, err
	}
	return result, nil
}

// QueryByWrapper 通过类型安全的 Wrapper 查询数据。
func (r Repository[R, M, T]) QueryByWrapper(query *QueryWrapper[T]) ([]*T, error) {
	result := make([]*T, 0)
	_, err := r.mapper.SelectByWrapper(query, &result)
	return result, err
}

// -------------------- 分页 --------------------

// QueryPageByCond 通过实体条件分页查询，条件中的零值字段将被自动忽略。
func (r Repository[R, M, T]) QueryPageByCond(query PageQuery[T]) (databasecloud.Pager[T], error) {
	pager := databasecloud.Pager[T]{Number: query.Number, Size: query.Size}
	total, err := r.mapper.SelectPageByCond(query, &pager.Records)
	if err != nil {
		return pager, err
	}
	pager.Total = total
	return pager, nil
}

// QueryPageByMap 通过 Map 条件分页查询，支持显式查询零值字段。
func (r Repository[R, M, T]) QueryPageByMap(query MapPageQuery) (databasecloud.Pager[T], error) {
	pager := databasecloud.Pager[T]{Number: query.Number, Size: query.Size}
	total, err := r.mapper.SelectPageByMap(query, &pager.Records)
	if err != nil {
		return pager, err
	}
	pager.Total = total
	return pager, nil
}

// QueryPageByWhere 通过原始 SQL 分页查询
func (r Repository[R, M, T]) QueryPageByWhere(query WherePageQuery) (databasecloud.Pager[T], error) {
	pager := databasecloud.Pager[T]{Number: query.Number, Size: query.Size}
	total, err := r.mapper.SelectPageByWhere(query, &pager.Records)
	if err != nil {
		return pager, err
	}
	pager.Total = total
	return pager, nil
}

// QueryPageByGorm 通过原始 Gorm 查询分页数据。
func (r Repository[R, M, T]) QueryPageByGorm(query GormPageQuery) (databasecloud.Pager[T], error) {
	pager := databasecloud.Pager[T]{Number: query.Number, Size: query.Size}
	total, err := r.mapper.SelectPageByGorm(query.CountRawDB, query.PageRawDB, &pager.Records)
	pager.Total = total
	return pager, err
}

// QueryPageByWrapper 通过类型安全的分页 Wrapper 查询数据。
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

// -------------------- 统计 --------------------

// CountByCond 通过实体条件统计数据。
func (r Repository[R, M, T]) CountByCond(query CondQuery[T]) (int64, error) {
	return r.mapper.CountByCond(query)
}

// CountByMap 通过 Map 条件统计数据，支持显式零值条件。
func (r Repository[R, M, T]) CountByMap(query MapQuery) (int64, error) {
	return r.mapper.CountByMap(query)
}

// CountByWhere 通过原始 SQL 条件统计数据。
func (r Repository[R, M, T]) CountByWhere(query WhereQuery) (int64, error) {
	return r.mapper.CountByWhere(query)
}

// CountByGorm 通过原始 Gorm 查询数据总数。
func (r Repository[R, M, T]) CountByGorm(rawDB func(*gorm.DB)) (int64, error) {
	return r.mapper.CountByGorm(rawDB)
}

// CountByWrapper 通过 Wrapper 条件统计数据总数。
func (r Repository[R, M, T]) CountByWrapper(query *QueryWrapper[T]) (int64, error) {
	return r.mapper.CountByWrapper(query)
}

// -------------------- 更新 --------------------

// ModifyByID 通过ID更新含零值字段 updateColumns 手动指定需要更新的列
func (r Repository[R, M, T]) ModifyByID(updated *T, id any, updateColumns ...string) (int64, error) {
	return r.mapper.UpdateByID(updated, id, updateColumns...)
}

// ModifyByIDWithoutZeroFields 通过 ID 更新非零值字段，includeZeroFieldColumns 额外指定需要更新的零值字段。
func (r Repository[R, M, T]) ModifyByIDWithoutZeroFields(updated *T, id any, includeZeroFieldColumns ...string) (int64, error) {
	return r.mapper.UpdateByIDWithoutZeroFields(updated, id, includeZeroFieldColumns...)
}

// ModifyByIDWithMap 通过ID更新所有map中指定的列和值
func (r Repository[R, M, T]) ModifyByIDWithMap(updated map[string]any, id any) (int64, error) {
	return r.mapper.UpdateByIDWithMap(updated, id)
}

// ModifyByCond 通过条件更新 条件：零值将自动忽略，更新：零值字段将被自动忽略
// updateColumns 需要指定更新的数据库字段 更新指定字段(支持零值字段)
func (r Repository[R, M, T]) ModifyByCond(updated *T, condition T, updateColumns ...string) (int64, error) {
	return r.mapper.UpdateByCond(updated, condition, updateColumns...)
}

// ModifyByCondWithZeroFields 通过条件更新，并指定可以更新的零值字段。
func (r Repository[R, M, T]) ModifyByCondWithZeroFields(updated *T, condition T, includeZeroFieldColumns ...string) (int64, error) {
	return r.mapper.UpdateByCondWithZeroFields(updated, condition, includeZeroFieldColumns...)
}

// ModifyByMap 通过Map类型条件更新
func (r Repository[R, M, T]) ModifyByMap(updated, condition map[string]any) (int64, error) {
	return r.mapper.UpdateByMap(updated, condition)
}

// ModifyByWhere 通过原始SQL查询条件，更新非零实体字段 Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql: "a = ?" args = 1
func (r Repository[R, M, T]) ModifyByWhere(updated *T, rawWhereSQL string, args ...any) (int64, error) {
	return r.mapper.UpdateByWhere(updated, rawWhereSQL, args...)
}

// ModifyByWrapper 通过 UpdateWrapper 的条件和 Set 赋值更新数据。
func (r Repository[R, M, T]) ModifyByWrapper(wrapper *UpdateWrapper[T]) (int64, error) {
	return r.mapper.UpdateByWrapper(wrapper)
}

// -------------------- 删除 --------------------

// RemoveByID 通过ID删除
func (r Repository[R, M, T]) RemoveByID(id any) (int64, error) {
	return r.mapper.DeleteByID(id)
}

// RemoveByIDs 通过ID批量删除
func (r Repository[R, M, T]) RemoveByIDs(ids []any) (int64, error) {
	return r.mapper.DeleteByIDs(ids)
}

// RemoveByCond 通过条件删除 零值字段将被自动忽略
func (r Repository[R, M, T]) RemoveByCond(condition T) (int64, error) {
	return r.mapper.DeleteByCond(condition)
}

// RemoveByMap 通过Map类型条件删除
func (r Repository[R, M, T]) RemoveByMap(condition map[string]any) (int64, error) {
	return r.mapper.DeleteByMap(condition)
}

// RemoveByWhere 通过原始SQL查询条件删除 Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql: "a = ?" args = 1
func (r Repository[R, M, T]) RemoveByWhere(rawWhereSQL string, args ...any) (int64, error) {
	return r.mapper.DeleteByWhere(rawWhereSQL, args...)
}
