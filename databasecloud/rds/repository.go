package rds

import (
	"database/sql"
	"errors"

	"github.com/acexy/golang-toolkit/logger"
	"github.com/golang-acexy/cloud-database/databasecloud"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"gorm.io/gorm"
)

// Mapper 在 GORM Mapper 基础上约束事务复制能力，确保事务仓库保留具体 Mapper 类型。
type Mapper[M any, T gormstarter.Model] interface {
	gormstarter.Mapper[T]
	WithTxMapper(tx *gorm.DB) M
}

// Repository 组合具体 Mapper，并通过 R 保留业务 Repository 类型。
type Repository[R any, M Mapper[M, T], T gormstarter.Model] struct {
	mapper  M
	factory func(Repository[R, M, T]) R
}

// PageQuery 定义关系型数据库分页查询参数。
type PageQuery struct {
	OrderBySQL     string
	SpecifyColumns []string
}

// NewRepository 创建基础 Repository，并注册业务 Repository 工厂函数。
func NewRepository[R any, M Mapper[M, T], T gormstarter.Model](mapper M, factory func(Repository[R, M, T]) R) Repository[R, M, T] {
	if factory == nil {
		panic(ErrNilRepositoryFactory)
	}
	return Repository[R, M, T]{mapper: mapper, factory: factory}
}

// NewTxRepo 创建绑定新事务的业务 Repository，调用方负责提交或回滚事务。
// 调用前必须完成 GORM Starter 和 Repository 初始化；事务启动错误由后续数据库操作返回。
// 若当前 Repository 已绑定事务，将记录错误日志但仍继续尝试创建新事务。
func (g Repository[R, M, T]) NewTxRepo(opts ...*sql.TxOptions) R {
	db, err := g.mapper.CurrentGorm()
	if err != nil {
		panic(err)
	}
	warnTransactionRepository(db, "NewTxRepo")
	tx := db.Begin(opts...)
	return g.repositoryWithTx(tx)
}

// WithTxRepo 创建绑定指定事务的业务 Repository，调用方负责保证事务有效并管理其生命周期。
// 调用前必须通过 NewRepository 完成 Repository 初始化。
// 若当前 Repository 已绑定事务，将记录错误日志但仍使用传入事务创建新 Repository。
func (g Repository[R, M, T]) WithTxRepo(tx *gorm.DB) R {
	if db, err := g.mapper.CurrentGorm(); err == nil {
		warnTransactionRepository(db, "WithTxRepo")
	}
	return g.repositoryWithTx(tx)
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
func (g Repository[R, M, T]) Transaction(fn func(R) error, opts ...*sql.TxOptions) (err error) {
	if fn == nil {
		return ErrNilTransactionFunc
	}
	if g.factory == nil {
		return ErrRepositoryNotInitialized
	}
	db, err := g.mapper.CurrentGorm()
	if err != nil {
		return err
	}
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

	if err = fn(g.repositoryWithTx(tx)); err != nil {
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			return errors.Join(err, rollbackErr)
		}
		return err
	}
	return tx.Commit().Error
}

func (g Repository[R, M, T]) repositoryWithTx(tx *gorm.DB) R {
	txRepository := g
	txRepository.mapper = g.mapper.WithTxMapper(tx)
	return g.factory(txRepository)
}

// >>>>>>>>>>>>>>> CRUD 操作API

// RawMapper 获取具体 Mapper
func (g Repository[R, M, T]) RawMapper() M {
	return g.mapper
}

// CurrentGORMDB 获取当前 GORM DB，如果已有事务则返回该事务，否则获取新的 GORM DB。
func (g Repository[R, M, T]) CurrentGORMDB() (*gorm.DB, error) {
	return g.mapper.CurrentGorm()
}

// TableGORMDB 获取已经限定当前 Mapper 表名的 GORM DB。
func (g Repository[R, M, T]) TableGORMDB() (*gorm.DB, error) {
	return g.mapper.GormWithTableName()
}

// Save 保存数据 默认零值数据也会进行存储 可通过设置excludeColumns排除零值数据
func (g Repository[R, M, T]) Save(entity *T, excludeColumns ...string) (int64, error) {
	return g.mapper.Insert(entity, excludeColumns...)
}

// SaveWithoutZeroFields 保存数据，零值字段不参与保存。
func (g Repository[R, M, T]) SaveWithoutZeroFields(entity *T) (int64, error) {
	return g.mapper.InsertWithoutZeroFields(entity)
}

// SaveWithMap 通过Map类型保存数据 key为列明 value为列值
func (g Repository[R, M, T]) SaveWithMap(entity map[string]any) (int64, error) {
	return g.mapper.InsertWithMap(entity)
}

// SaveOrModifyByPrimaryKey 保存/更新数据 (主键冲突则执行更新) 零值也将参与保存
func (g Repository[R, M, T]) SaveOrModifyByPrimaryKey(entity *T, excludeColumns ...string) (int64, error) {
	return g.mapper.InsertOrUpdateByPrimaryKey(entity, excludeColumns...)
}

// SaveBatch 批量保存数据 默认零值数据也会进行存储 可通过设置excludeColumns排除零值数据
func (g Repository[R, M, T]) SaveBatch(entities []*T, excludeColumns ...string) (int64, error) {
	return g.mapper.InsertBatch(entities, excludeColumns...)
}

// QueryByID 通过主键查询数据
func (g Repository[R, M, T]) QueryByID(id any, result *T) (int64, error) {
	return g.mapper.SelectByID(id, result)
}

// QueryByIDs 通过主键查询数据
func (g Repository[R, M, T]) QueryByIDs(ids []any, result *[]*T) (int64, error) {
	return g.mapper.SelectByIDs(ids, result)
}

// ExistsByID 判断指定主键的数据是否存在。
func (g Repository[R, M, T]) ExistsByID(id any) (bool, error) {
	return g.mapper.ExistsByID(id)
}

// QueryOneByCond 通过条件查询 查询条件零值字段将被自动忽略 specifyColumns 指定只需要查询的数据库字段
func (g Repository[R, M, T]) QueryOneByCond(condition *T, result *T, specifyColumns ...string) (int64, error) {
	return g.mapper.SelectOneByCond(condition, result, specifyColumns...)
}

// QueryByCond 通过条件查询 查询条件零值字段将被自动忽略 specifyColumns 指定只需要查询的数据库字段
func (g Repository[R, M, T]) QueryByCond(condition *T, orderBySQL string, result *[]*T, specifyColumns ...string) (int64, error) {
	return g.mapper.SelectByCond(condition, orderBySQL, result, specifyColumns...)
}

// QueryOneByMap 通过指定字段与值查询数据 解决零值条件问题 specifyColumns 指定只需要查询的数据库字段
func (g Repository[R, M, T]) QueryOneByMap(condition map[string]any, result *T, specifyColumns ...string) (int64, error) {
	return g.mapper.SelectOneByMap(condition, result, specifyColumns...)
}

// QueryByMap 通过指定字段与值查询数据 解决零值条件问题 specifyColumns 指定只需要查询的数据库字段
func (g Repository[R, M, T]) QueryByMap(condition map[string]any, orderBySQL string, result *[]*T, specifyColumns ...string) (int64, error) {
	return g.mapper.SelectByMap(condition, orderBySQL, result, specifyColumns...)
}

// QueryOneByWhere 通过原始Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql: "a = ?"  args = 1
func (g Repository[R, M, T]) QueryOneByWhere(rawWhereSQL string, result *T, args ...any) (int64, error) {
	return g.mapper.SelectOneByWhere(rawWhereSQL, result, args...)
}

// QueryByWhere 通过原始Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSq: "a = ?" args = 1
func (g Repository[R, M, T]) QueryByWhere(rawWhereSQL, orderBySQL string, result *[]*T, args ...any) (int64, error) {
	return g.mapper.SelectByWhere(rawWhereSQL, orderBySQL, result, args...)
}

// QueryOneByGORM 通过原始 GORM 查询单条数据。
func (g Repository[R, M, T]) QueryOneByGORM(result *T, rawDB func(*gorm.DB)) (int64, error) {
	return g.mapper.SelectOneByGorm(result, rawDB)
}

// QueryByGORM 通过原始 GORM 查询数据。
func (g Repository[R, M, T]) QueryByGORM(result *[]*T, rawDB func(*gorm.DB)) (int64, error) {
	return g.mapper.SelectByGorm(result, rawDB)
}

// QueryPageByCond 通过条件分页查询 零值字段将被自动忽略 specifyColumns 指定只需要查询的数据库字段
func (g Repository[R, M, T]) QueryPageByCond(condition *T, query PageQuery, pager *databasecloud.Pager[T]) error {
	total, err := g.mapper.SelectPageByCond(condition, gormstarter.PageQuery{
		PageNumber:     pager.Number,
		PageSize:       pager.Size,
		OrderBySQL:     query.OrderBySQL,
		SpecifyColumns: query.SpecifyColumns,
	}, &pager.Records)
	if err != nil {
		return err
	}
	pager.Total = total
	return nil
}

// QueryPageByMap 通过指定字段与值查询数据分页查询 解决零值条件问题 specifyColumns 指定只需要查询的数据库字段
func (g Repository[R, M, T]) QueryPageByMap(condition map[string]any, query PageQuery, pager *databasecloud.Pager[T]) error {
	total, err := g.mapper.SelectPageByMap(condition, gormstarter.PageQuery{
		PageNumber:     pager.Number,
		PageSize:       pager.Size,
		OrderBySQL:     query.OrderBySQL,
		SpecifyColumns: query.SpecifyColumns,
	}, &pager.Records)
	if err != nil {
		return err
	}
	pager.Total = total
	return nil
}

// QueryPageByWhere 通过原始 SQL 分页查询
func (g Repository[R, M, T]) QueryPageByWhere(rawWhereSQL string, query PageQuery, pager *databasecloud.Pager[T], args ...any) error {
	total, err := g.mapper.SelectPageByWhere(rawWhereSQL, gormstarter.PageQuery{
		PageNumber:     pager.Number,
		PageSize:       pager.Size,
		OrderBySQL:     query.OrderBySQL,
		SpecifyColumns: query.SpecifyColumns,
	}, &pager.Records, args...)
	if err != nil {
		return err
	}
	pager.Total = total
	return nil
}

// QueryPageByGORM 通过原始 GORM 查询分页数据。
func (g Repository[R, M, T]) QueryPageByGORM(countRawDB func(*gorm.DB), pageRawDB func(*gorm.DB), result *[]*T) (int64, error) {
	return g.mapper.SelectPageByGorm(countRawDB, pageRawDB, result)
}

// CountByCond 通过条件查询数据总数
func (g Repository[R, M, T]) CountByCond(condition *T) (int64, error) {
	return g.mapper.CountByCond(condition)
}

// CountByMap 通过指定字段与值查询数据总数 解决零值条件问题
func (g Repository[R, M, T]) CountByMap(condition map[string]any) (int64, error) {
	return g.mapper.CountByMap(condition)
}

// CountByWhere 通过原始SQL查询数据总数
func (g Repository[R, M, T]) CountByWhere(rawWhereSQL string, args ...any) (int64, error) {
	return g.mapper.CountByWhere(rawWhereSQL, args...)
}

// CountByGORM 通过原始 GORM 查询数据总数。
func (g Repository[R, M, T]) CountByGORM(rawDB func(*gorm.DB)) (int64, error) {
	return g.mapper.CountByGorm(rawDB)
}

// ModifyByID 通过ID更新含零值字段 updateColumns 手动指定需要更新的列
func (g Repository[R, M, T]) ModifyByID(updated *T, updateColumns ...string) (int64, error) {
	return g.mapper.UpdateByID(updated, updateColumns...)
}

// ModifyByIDWithoutZeroFields 通过 ID 更新非零值字段，includeZeroFieldColumns 额外指定需要更新的零值字段。
func (g Repository[R, M, T]) ModifyByIDWithoutZeroFields(updated *T, includeZeroFieldColumns ...string) (int64, error) {
	return g.mapper.UpdateByIDWithoutZeroFields(updated, includeZeroFieldColumns...)
}

// ModifyByIDWithMap 通过ID更新所有map中指定的列和值
func (g Repository[R, M, T]) ModifyByIDWithMap(updated map[string]any, id any) (int64, error) {
	return g.mapper.UpdateByIDWithMap(updated, id)
}

// ModifyByCond 通过条件更新 条件：零值将自动忽略，更新：零值字段将被自动忽略
// updateColumns 需要指定更新的数据库字段 更新指定字段(支持零值字段)
func (g Repository[R, M, T]) ModifyByCond(updated, condition *T, updateColumns ...string) (int64, error) {
	return g.mapper.UpdateByCond(updated, condition, updateColumns...)
}

// ModifyByCondWithZeroFields 通过条件更新，并指定可以更新的零值字段。
func (g Repository[R, M, T]) ModifyByCondWithZeroFields(updated, condition *T, includeZeroFieldColumns ...string) (int64, error) {
	return g.mapper.UpdateByCondWithZeroFields(updated, condition, includeZeroFieldColumns...)
}

// ModifyByMap 通过Map类型条件更新
func (g Repository[R, M, T]) ModifyByMap(updated, condition map[string]any) (int64, error) {
	return g.mapper.UpdateByMap(updated, condition)
}

// ModifyByWhere 通过原始SQL查询条件，更新非零实体字段 Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql: "a = ?" args = 1
func (g Repository[R, M, T]) ModifyByWhere(updated *T, rawWhereSQL string, args ...any) (int64, error) {
	return g.mapper.UpdateByWhere(updated, rawWhereSQL, args...)
}

// RemoveByID 通过ID删除
func (g Repository[R, M, T]) RemoveByID(id any) (int64, error) {
	return g.mapper.DeleteByID(id)
}

// RemoveByIDs 通过ID批量删除
func (g Repository[R, M, T]) RemoveByIDs(ids []any) (int64, error) {
	return g.mapper.DeleteByIDs(ids)
}

// RemoveByCond 通过条件删除 零值字段将被自动忽略
func (g Repository[R, M, T]) RemoveByCond(condition *T) (int64, error) {
	return g.mapper.DeleteByCond(condition)
}

// RemoveByMap 通过Map类型条件删除
func (g Repository[R, M, T]) RemoveByMap(condition map[string]any) (int64, error) {
	return g.mapper.DeleteByMap(condition)
}

// RemoveByWhere 通过原始SQL查询条件删除 Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql: "a = ?" args = 1
func (g Repository[R, M, T]) RemoveByWhere(rawWhereSQL string, args ...any) (int64, error) {
	return g.mapper.DeleteByWhere(rawWhereSQL, args...)
}
