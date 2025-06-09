package databasecloud

import (
	"database/sql"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"gorm.io/gorm"
)

type GormRepository[B gormstarter.IBaseMapper[M, T], M gormstarter.BaseMapper[T], T gormstarter.IBaseModel] struct {
	Mapper B // mapper接口的声明类型
	mapper M // mapper的实际实现
}

// NewTxRepo 创建一个全新事务数据库服务
func (g GormRepository[B, M, T]) NewTxRepo(opts ...*sql.TxOptions) GormRepository[gormstarter.IBaseMapper[M, T], gormstarter.BaseMapper[T], T] {
	mapperWithTx := g.Mapper.NewBaseMapperWithTx(opts...)
	return GormRepository[gormstarter.IBaseMapper[M, T], gormstarter.BaseMapper[T], T]{mapper: mapperWithTx, Mapper: mapperWithTx}
}

// WithTxRepo 创建一个带有指定事务数据库服务
func (g GormRepository[B, M, T]) WithTxRepo(tx *gorm.DB) GormRepository[gormstarter.IBaseMapper[M, T], gormstarter.BaseMapper[T], T] {
	mapperWithTx := g.Mapper.GetBaseMapperWithTx(tx)
	return GormRepository[gormstarter.IBaseMapper[M, T], gormstarter.BaseMapper[T], T]{mapper: mapperWithTx, Mapper: mapperWithTx}
}

// >>>>>>>>>>>>>>> CRUD 操作API

// RawIMapper 获取原始基础Mapper
func (g GormRepository[B, M, T]) RawIMapper() B {
	return g.Mapper
}

// CurrentGormDB 获取当前的gorm.DB 如果已有事务则返回该事务，否则获取新的gorm.DB
func (g GormRepository[B, M, T]) CurrentGormDB() *gorm.DB {
	return g.Mapper.CurrentGorm()
}

// Save 保存数据 默认零值数据也会进行存储 可通过设置excludeColumns排除零值数据
func (g GormRepository[B, M, T]) Save(entity *T, excludeColumns ...string) (int64, error) {
	return g.Mapper.Insert(entity, excludeColumns...)
}

// SaveExcludeZeroField 保存数据 零值数据不参与保存
func (g GormRepository[B, M, T]) SaveExcludeZeroField(entity *T) (int64, error) {
	return g.Mapper.InsertWithoutZeroField(entity)
}

// SaveUseMap 通过Map类型保存数据 key为列明 value为列值
func (g GormRepository[B, M, T]) SaveUseMap(entity map[string]any) (int64, error) {
	return g.Mapper.InsertUseMap(entity)
}

// SaveOrModifyByPrimaryKey 保存/更新数据 (主键冲突则执行更新) 零值也将参与保存
func (g GormRepository[B, M, T]) SaveOrModifyByPrimaryKey(entity *T, excludeColumns ...string) (int64, error) {
	return g.Mapper.InsertOrUpdateByPrimaryKey(entity, excludeColumns...)
}

// SaveBatch 批量保存数据 默认零值数据也会进行存储 可通过设置excludeColumns排除零值数据
func (g GormRepository[B, M, T]) SaveBatch(entities *[]*T, excludeColumns ...string) (int64, error) {
	return g.Mapper.InsertBatch(entities, excludeColumns...)
}

// QueryByID 通过主键查询数据
func (g GormRepository[B, M, T]) QueryByID(id any, result *T) (int64, error) {
	return g.Mapper.SelectById(id, result)
}

// QueryByIDs 通过主键查询数据
func (g GormRepository[B, M, T]) QueryByIDs(id []any, result *[]*T) (int64, error) {
	return g.Mapper.SelectByIds(id, result)
}

// QueryOneByCond 通过条件查询 查询条件零值字段将被自动忽略 specifyColumns 指定只需要查询的数据库字段
func (g GormRepository[B, M, T]) QueryOneByCond(condition *T, result *T, specifyColumns ...string) (int64, error) {
	return g.Mapper.SelectOneByCond(condition, result, specifyColumns...)
}

// QueryByCond 通过条件查询 查询条件零值字段将被自动忽略 specifyColumns 指定只需要查询的数据库字段
func (g GormRepository[B, M, T]) QueryByCond(condition *T, sqlOrderBy string, result *[]*T, specifyColumns ...string) (int64, error) {
	return g.Mapper.SelectByCond(condition, sqlOrderBy, result, specifyColumns...)
}

// QueryOneByMap 通过指定字段与值查询数据 解决零值条件问题 specifyColumns 指定只需要查询的数据库字段
func (g GormRepository[B, M, T]) QueryOneByMap(condition map[string]any, result *T, specifyColumns ...string) (int64, error) {
	return g.Mapper.SelectOneByMap(condition, result, specifyColumns...)
}

// QueryByMap 通过指定字段与值查询数据 解决零值条件问题 specifyColumns 指定只需要查询的数据库字段
func (g GormRepository[B, M, T]) QueryByMap(condition map[string]any, sqlOrderBy string, result *[]*T, specifyColumns ...string) (int64, error) {
	return g.Mapper.SelectByMap(condition, sqlOrderBy, result, specifyColumns...)
}

// QueryOneByWhere 通过原始Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql: "a = ?"  args = 1
func (g GormRepository[B, M, T]) QueryOneByWhere(rawWhereSql string, result *T, args ...any) (int64, error) {
	return g.Mapper.SelectOneByWhere(rawWhereSql, result, args...)
}

// QueryByWhere 通过原始Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSq: "a = ?" args = 1
func (g GormRepository[B, M, T]) QueryByWhere(rawWhereSql, orderBy string, result *[]*T, args ...interface{}) (int64, error) {
	return g.Mapper.SelectByWhere(rawWhereSql, orderBy, result, args...)
}

// QueryOneByGorm 通过原始Gorm查询单条数据 构建Gorm查询条件
func (g GormRepository[B, M, T]) QueryOneByGorm(result *T, rawDb func(*gorm.DB)) (int64, error) {
	return g.Mapper.SelectOneByGorm(result, rawDb)
}

// QueryByGorm 通过原始Gorm查询数据
func (g GormRepository[B, M, T]) QueryByGorm(result *[]*T, rawDb func(*gorm.DB)) (int64, error) {
	return g.Mapper.SelectByGorm(result, rawDb)
}

// QueryPageByCond 通过条件分页查询 零值字段将被自动忽略 specifyColumns 指定只需要查询的数据库字段
func (g GormRepository[B, M, T]) QueryPageByCond(condition *T, sqlOrderBy string, pager *Pager[T], specifyColumns ...string) error {
	total, err := g.Mapper.SelectPageByCond(condition, sqlOrderBy, pager.Num, pager.Size, &pager.Records, specifyColumns...)
	if err != nil {
		return err
	}
	pager.Total = total
	return nil
}

// QueryPageByMap 通过指定字段与值查询数据分页查询 解决零值条件问题 specifyColumns 指定只需要查询的数据库字段
func (g GormRepository[B, M, T]) QueryPageByMap(condition map[string]any, sqlOrderBy string, pager *Pager[T], specifyColumns ...string) error {
	total, err := g.Mapper.SelectPageByMap(condition, sqlOrderBy, pager.Num, pager.Size, &pager.Records, specifyColumns...)
	if err != nil {
		return err
	}
	pager.Total = total
	return nil
}

// QueryPageByWhere 通过原始SQL分页查询 rawWhereSql 例如 where a = 1 则只需要rawWhereSq: "a = ?" args = 1
func (g GormRepository[B, M, T]) QueryPageByWhere(rawWhereSql, orderBy string, pager *Pager[T], args []any, specifyColumns ...string) error {
	total, err := g.Mapper.SelectPageByWhere(rawWhereSql, orderBy, pager.Num, pager.Size, &pager.Records, args, specifyColumns...)
	if err != nil {
		return err
	}
	pager.Total = total
	return nil
}

// QueryPageByGorm 通过原始Gorm查询分页数据
func (g GormRepository[B, M, T]) QueryPageByGorm(countRawDb func(*gorm.DB), pageRawDb func(*gorm.DB), result *[]*T) (int64, error) {
	return g.Mapper.SelectPageByGorm(countRawDb, pageRawDb, result)
}

// CountByCond 通过条件查询数据总数
func (g GormRepository[B, M, T]) CountByCond(condition *T) (int64, error) {
	return g.Mapper.CountByCond(condition)
}

// CountByMap 通过指定字段与值查询数据总数 解决零值条件问题
func (g GormRepository[B, M, T]) CountByMap(condition map[string]any) (int64, error) {
	return g.Mapper.CountByMap(condition)
}

// CountByWhere 通过原始SQL查询数据总数
func (g GormRepository[B, M, T]) CountByWhere(rawWhereSql string, args ...any) (int64, error) {
	return g.Mapper.CountByWhere(rawWhereSql, args...)
}

// CountByGorm 通过原始Gorm查询数据总数
func (g GormRepository[B, M, T]) CountByGorm(rawDb func(*gorm.DB)) (int64, error) {
	return g.Mapper.CountByGorm(rawDb)
}

// ModifyByID 通过ID更新含零值字段 updateColumns 手动指定需要更新的列
func (g GormRepository[B, M, T]) ModifyByID(updated *T, updateColumns ...string) (int64, error) {
	return g.Mapper.UpdateById(updated, updateColumns...)
}

// ModifyByIDExcludeZeroField 通过ID更新非零值字段 includeZeroFiledColumns 额外指定需要更新零值字段
func (g GormRepository[B, M, T]) ModifyByIDExcludeZeroField(updated *T, includeZeroFiledColumns ...string) (int64, error) {
	return g.Mapper.UpdateByIdWithoutZeroField(updated, includeZeroFiledColumns...)
}

// ModifyByIdUseMap 通过ID更新所有map中指定的列和值
func (g GormRepository[B, M, T]) ModifyByIdUseMap(updated map[string]any, id any) (int64, error) {
	return g.Mapper.UpdateByIdUseMap(updated, id)
}

// ModifyByCond 通过条件更新 条件：零值将自动忽略，更新：零值字段将被自动忽略
// updateColumns 需要指定更新的数据库字段 更新指定字段(支持零值字段)
func (g GormRepository[B, M, T]) ModifyByCond(updated, condition *T, updateColumns ...string) (int64, error) {
	return g.Mapper.UpdateByCond(updated, condition, updateColumns...)
}

// ModifyByCondIncludeZeroField 通过条件更新，并指定可以更新的零值字段 includeZeroFiledColumns 额外指定需要更新零值字段
func (g GormRepository[B, M, T]) ModifyByCondIncludeZeroField(updated, condition *T, includeZeroFiledColumns []string) (int64, error) {
	return g.Mapper.UpdateByCondWithZeroField(updated, condition, includeZeroFiledColumns)
}

// ModifyByMap 通过Map类型条件更新
func (g GormRepository[B, M, T]) ModifyByMap(updated, condition map[string]any) (int64, error) {
	return g.Mapper.UpdateByMap(updated, condition)
}

// ModifyByWhere 通过原始SQL查询条件，更新非零实体字段 Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql: "a = ?" args = 1
func (g GormRepository[B, M, T]) ModifyByWhere(updated *T, rawWhereSql string, args ...any) (int64, error) {
	return g.Mapper.UpdateByWhere(updated, rawWhereSql, args...)
}

// RemoveByID 通过ID删除
func (g GormRepository[B, M, T]) RemoveByID(id any) (int64, error) {
	return g.Mapper.DeleteById(id)
}

// RemoveByIDs 通过ID批量删除
func (g GormRepository[B, M, T]) RemoveByIDs(ids []any) (int64, error) {
	return g.Mapper.DeleteById(ids...)
}

// RemoveByCond 通过条件删除 零值字段将被自动忽略
func (g GormRepository[B, M, T]) RemoveByCond(condition *T) (int64, error) {
	return g.Mapper.DeleteByCond(condition)
}

// RemoveByMap 通过Map类型条件删除
func (g GormRepository[B, M, T]) RemoveByMap(condition map[string]any) (int64, error) {
	return g.Mapper.DeleteByMap(condition)
}

// RemoveByWhere 通过原始SQL查询条件删除 Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql: "a = ?" args = 1
func (g GormRepository[B, M, T]) RemoveByWhere(rawWhereSql string, args ...interface{}) (int64, error) {
	return g.Mapper.DeleteByWhere(rawWhereSql, args...)
}
