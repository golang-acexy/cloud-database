package databasecloud

import (
	"github.com/golang-acexy/starter-mongo/mongostarter"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoRepository[B mongostarter.IBaseMapper[M, T], M mongostarter.BaseMapper[T], T mongostarter.IBaseModel] struct {
	Mapper B // mapper接口的声明类型
	mapper M // mapper的实际实现
}

// RawIMapper 获取原始基础Mapper
func (m MongoRepository[B, M, T]) RawIMapper() B {
	return m.Mapper
}

// CollWithTable 获取mongo.Collection并已限定当前mapper对应的表名
func (m MongoRepository[B, M, T]) CollWithTable() *mongo.Collection {
	return m.Mapper.CollWithTableName()
}

// Save 保存数据 并返回objectId
func (m MongoRepository[B, M, T]) Save(entity *T) (string, error) {
	return m.Mapper.Insert(entity)
}

// SaveUseBson 保存数据使用Bson返回objectId
func (m MongoRepository[B, M, T]) SaveUseBson(entity bson.M) (string, error) {
	return m.Mapper.InsertByBson(entity)
}

// InsertWithOption 插入数据并返回objectId
func (m MongoRepository[B, M, T]) InsertWithOption(document interface{}, opts ...options.Lister[options.InsertOneOptions]) (string, error) {
	return m.Mapper.InsertWithOption(document, opts...)
}

// SaveBatch 批量保存数据并返回objectId
func (m MongoRepository[B, M, T]) SaveBatch(entities *[]*T) ([]string, error) {
	return m.Mapper.InsertBatch(entities)
}

// SaveBatchUseBson 批量保存数据并返回objectId
func (m MongoRepository[B, M, T]) SaveBatchUseBson(entities bson.A) ([]string, error) {
	return m.Mapper.InsertBatchByBson(entities)
}

// SaveBatchWithOption 批量保存数据并返回objectId
func (m MongoRepository[B, M, T]) SaveBatchWithOption(documents interface{}, opts ...options.Lister[options.InsertManyOptions]) ([]string, error) {
	return m.Mapper.InsertBatchWithOption(documents, opts...)
}

// QueryByID 根据id查询数据 默认匹配_id字段 传入string的主键时，默认转换为mongo hex 如果不是该类型，需设置notObjectId为true
func (m MongoRepository[B, M, T]) QueryByID(id any, result *T, notObjectId ...bool) error {
	return m.Mapper.SelectById(id, result, notObjectId...)
}

// QueryByIDs 根据ids查询数据 默认匹配_id字段 传入string的主键时，默认转换为mongo hex 如果不是该类型，需设置notObjectId为true
func (m MongoRepository[B, M, T]) QueryByIDs(ids []any, result *[]*T, notObjectId ...bool) (err error) {
	return m.Mapper.SelectByIds(ids, result, notObjectId...)
}

// QueryOneByCond 根据条件查询一条数据
func (m MongoRepository[B, M, T]) QueryOneByCond(condition *T, result *T, specifyColumns ...string) error {
	return m.Mapper.SelectOneByCond(condition, result, specifyColumns...)
}

// QueryByCond 根据条件查询数据
func (m MongoRepository[B, M, T]) QueryByCond(condition *T, orderBy []*mongostarter.OrderBy, result *[]*T, specifyColumns ...string) error {
	return m.Mapper.SelectByCond(condition, orderBy, result, specifyColumns...)
}

// QueryOneByBson 根据条件查询一条数据
func (m MongoRepository[B, M, T]) QueryOneByBson(condition bson.M, result *T, specifyColumns ...string) error {
	return m.Mapper.SelectOneByBson(condition, result, specifyColumns...)
}

// QueryByBson 根据条件查询数据
func (m MongoRepository[B, M, T]) QueryByBson(condition bson.M, orderBy []*mongostarter.OrderBy, result *[]*T, specifyColumns ...string) error {
	return m.Mapper.SelectByBson(condition, orderBy, result, specifyColumns...)
}

// QueryOneByOption 根据条件查询一条数据
func (m MongoRepository[B, M, T]) QueryOneByOption(filter interface{}, result *T, opts ...options.Lister[options.FindOneOptions]) error {
	return m.Mapper.SelectOneByOption(filter, result, opts...)
}

// QueryByOption 根据条件查询数据
func (m MongoRepository[B, M, T]) QueryByOption(filter interface{}, result *[]*T, opts ...options.Lister[options.FindOptions]) error {
	return m.Mapper.SelectByOption(filter, result, opts...)
}

// QueryPageByCond 根据条件查询分页数据
func (m MongoRepository[B, M, T]) QueryPageByCond(condition *T, orderBy []*mongostarter.OrderBy, pager *Pager[T], specifyColumns ...string) error {
	total, err := m.Mapper.SelectPageByCond(condition, orderBy, pager.Number, pager.Size, &pager.Records, specifyColumns...)
	if err != nil {
		return err
	}
	pager.Total = total
	return nil
}

// QueryPageByBson 根据条件查询分页数据
func (m MongoRepository[B, M, T]) QueryPageByBson(condition bson.M, orderBy []*mongostarter.OrderBy, pager *Pager[T], specifyColumns ...string) error {
	total, err := m.Mapper.SelectPageByBson(condition, orderBy, pager.Number, pager.Size, &pager.Records, specifyColumns...)
	if err != nil {
		return err
	}
	pager.Total = total
	return nil
}

// QueryPageByOption 根据条件查询分页数据
func (m MongoRepository[B, M, T]) QueryPageByOption(filter interface{}, orderBy []*mongostarter.OrderBy, pager *Pager[T], opts ...options.Lister[options.FindOptions]) error {
	total, err := m.Mapper.SelectPageByOption(filter, orderBy, pager.Number, pager.Size, &pager.Records, opts...)
	if err != nil {
		return err
	}
	pager.Total = total
	return nil
}

// CountByCond 根据条件查询统计数据
func (m MongoRepository[B, M, T]) CountByCond(condition *T) (int64, error) {
	return m.Mapper.CountByCond(condition)
}

// CountByBson 根据条件查询统计数据
func (m MongoRepository[B, M, T]) CountByBson(condition bson.M) (int64, error) {
	return m.Mapper.CountByBson(condition)
}

// CountByOption 根据条件查询统计数据
func (m MongoRepository[B, M, T]) CountByOption(filter interface{}, opts ...options.Lister[options.CountOptions]) (int64, error) {
	return m.Mapper.CountByOption(filter, opts...)
}

// ModifyByID 根据id修改数据
func (m MongoRepository[B, M, T]) ModifyByID(update *T, id any, notObjectId ...bool) (bool, error) {
	return m.Mapper.UpdateById(update, id, notObjectId...)
}

// ModifyByIdUseBson 根据id修改数据
func (m MongoRepository[B, M, T]) ModifyByIdUseBson(update bson.M, id any, notObjectId ...bool) (bool, error) {
	return m.Mapper.UpdateByIdUseBson(update, id, notObjectId...)
}

// ModifyOneByCond 根据条件修改一条数据
func (m MongoRepository[B, M, T]) ModifyOneByCond(update, condition *T) (bool, error) {
	return m.Mapper.UpdateOneByCond(update, condition)
}

// ModifyByCond 根据条件修改数据
func (m MongoRepository[B, M, T]) ModifyByCond(update, condition *T) (bool, error) {
	return m.Mapper.UpdateByCond(update, condition)
}

// ModifyOneByCondUseBson 根据条件修改一条数据
func (m MongoRepository[B, M, T]) ModifyOneByCondUseBson(update, condition bson.M) (bool, error) {
	return m.Mapper.UpdateOneByCondUseBson(update, condition)
}

// ModifyByCondUseBson 根据条件修改数据
func (m MongoRepository[B, M, T]) ModifyByCondUseBson(update, condition bson.M) (bool, error) {
	return m.Mapper.UpdateByCondUseBson(update, condition)
}

// RemoveByID 根据id删除数据
func (m MongoRepository[B, M, T]) RemoveByID(id any, notObjectId ...bool) (bool, error) {
	return m.Mapper.DeleteById(id, notObjectId...)
}

// RemoveOneByCond 根据条件删除一条数据
func (m MongoRepository[B, M, T]) RemoveOneByCond(condition *T) (bool, error) {
	return m.Mapper.DeleteOneByCond(condition)
}

// RemoveByCond 根据条件删除数据
func (m MongoRepository[B, M, T]) RemoveByCond(condition *T) (bool, error) {
	return m.Mapper.DeleteByCond(condition)
}

// RemoveOneByCondUseBson 根据条件删除一条数据
func (m MongoRepository[B, M, T]) RemoveOneByCondUseBson(condition bson.M) (bool, error) {
	return m.Mapper.DeleteOneByCondUseBson(condition)
}

// RemoveByCondUseBson 根据条件删除数据
func (m MongoRepository[B, M, T]) RemoveByCondUseBson(condition bson.M) (bool, error) {
	return m.Mapper.DeleteByCondUseBson(condition)
}
