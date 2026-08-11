package mongo

import (
	"github.com/golang-acexy/cloud-database/databasecloud"
	"github.com/golang-acexy/starter-mongo/mongostarter"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository[M mongostarter.Mapper[T], T mongostarter.Model] struct {
	mapper M
}

// QueryOptions 是 MongoDB 查询选项在 Repository 层的门面别名。
type QueryOptions = mongostarter.QueryOptions

// PageOptions 是 MongoDB 分页选项在 Repository 层的门面别名。
type PageOptions = mongostarter.PageOptions

// CondQuery 是 MongoDB 实体条件查询在 Repository 层的门面别名。
type CondQuery[T mongostarter.Model] = mongostarter.CondQuery[T]

// BSONQuery 是 MongoDB BSON 条件查询在 Repository 层的门面别名。
type BSONQuery = mongostarter.BSONQuery

// BSONPageQuery 是 MongoDB BSON 条件分页查询在 Repository 层的门面别名。
type BSONPageQuery = mongostarter.BSONPageQuery

// FilterPageQuery 是 MongoDB 原生 Filter 分页查询在 Repository 层的门面别名。
type FilterPageQuery = mongostarter.FilterPageQuery

// PageQuery 是 MongoDB 实体条件分页查询在 Repository 层的门面别名。
type PageQuery[T mongostarter.Model] = mongostarter.PageQuery[T]

// NewRepository 创建 MongoDB Repository。
func NewRepository[M mongostarter.Mapper[T], T mongostarter.Model](mapper M) Repository[M, T] {
	return Repository[M, T]{mapper: mapper}
}

// RawMapper 获取具体 Mapper
func (r Repository[M, T]) RawMapper() M {
	return r.mapper
}

// Collection 获取当前 Mapper 对应的 MongoDB Collection
func (r Repository[M, T]) Collection() *mongo.Collection {
	return r.mapper.Collection()
}

// Save 保存数据并返回 ObjectID
func (r Repository[M, T]) Save(entity *T) (string, error) {
	return r.mapper.Insert(entity)
}

// SaveWithBSON 使用 BSON 文档保存数据并返回 ObjectID
func (r Repository[M, T]) SaveWithBSON(entity bson.M) (string, error) {
	return r.mapper.InsertWithBSON(entity)
}

// SaveWithOptions 使用原生选项插入数据并返回 ObjectID
func (r Repository[M, T]) SaveWithOptions(document any, opts ...options.Lister[options.InsertOneOptions]) (string, error) {
	return r.mapper.InsertWithOptions(document, opts...)
}

// SaveBatch 批量保存数据并返回 ObjectID
func (r Repository[M, T]) SaveBatch(entities []*T) ([]string, error) {
	return r.mapper.InsertBatch(entities)
}

// SaveBatchWithBSON 使用 BSON 文档批量保存数据并返回 ObjectID
func (r Repository[M, T]) SaveBatchWithBSON(entities bson.A) ([]string, error) {
	return r.mapper.InsertBatchWithBSON(entities)
}

// SaveBatchWithOptions 使用原生选项批量保存数据并返回 ObjectID
func (r Repository[M, T]) SaveBatchWithOptions(documents any, opts ...options.Lister[options.InsertManyOptions]) ([]string, error) {
	return r.mapper.InsertBatchWithOptions(documents, opts...)
}

// QueryByID 根据 ID 查询数据，普通字符串 ID 需要将 notObjectID 设置为 true
func (r Repository[M, T]) QueryByID(id any, notObjectID ...bool) (*T, error) {
	result := new(T); if err := r.mapper.SelectByID(id, result, notObjectID...); err != nil { return nil, err }; return result, nil
}

// QueryByIDs 根据多个 ID 查询数据，普通字符串 ID 需要将 notObjectID 设置为 true
func (r Repository[M, T]) QueryByIDs(ids []any, notObjectID ...bool) ([]*T, error) {
	result := make([]*T, 0); err := r.mapper.SelectByIDs(ids, &result, notObjectID...); return result, err
}

// ExistsByID 判断指定主键的数据是否存在。
func (r Repository[M, T]) ExistsByID(id any, notObjectID ...bool) (bool, error) {
	return r.mapper.ExistsByID(id, notObjectID...)
}

// QueryOneByCond 根据条件查询一条数据
func (r Repository[M, T]) QueryOneByCond(query CondQuery[T]) (*T, error) {
	result := new(T); if err := r.mapper.SelectOneByCond(query, result); err != nil { return nil, err }; return result, nil
}

// QueryByCond 根据条件查询数据
func (r Repository[M, T]) QueryByCond(query CondQuery[T]) ([]*T, error) {
	result := make([]*T, 0); err := r.mapper.SelectByCond(query, &result); return result, err
}

// QueryOneByBSON 根据条件查询一条数据
func (r Repository[M, T]) QueryOneByBSON(query BSONQuery) (*T, error) {
	result := new(T); if err := r.mapper.SelectOneByBSON(query, result); err != nil { return nil, err }; return result, nil
}

// QueryByBSON 根据条件查询数据
func (r Repository[M, T]) QueryByBSON(query BSONQuery) ([]*T, error) {
	result := make([]*T, 0); err := r.mapper.SelectByBSON(query, &result); return result, err
}

// QueryOneWithOptions 根据条件查询一条数据
func (r Repository[M, T]) QueryOneWithOptions(filter any, opts ...options.Lister[options.FindOneOptions]) (*T, error) {
	result := new(T); if err := r.mapper.SelectOneWithOptions(filter, result, opts...); err != nil { return nil, err }; return result, nil
}

// QueryWithOptions 根据条件查询数据
func (r Repository[M, T]) QueryWithOptions(filter any, opts ...options.Lister[options.FindOptions]) ([]*T, error) {
	result := make([]*T, 0); err := r.mapper.SelectWithOptions(filter, &result, opts...); return result, err
}

// QueryPageByCond 根据条件查询分页数据
func (r Repository[M, T]) QueryPageByCond(query PageQuery[T]) (databasecloud.Pager[T], error) {
	pager := databasecloud.Pager[T]{Number: query.Number, Size: query.Size}
	total, err := r.mapper.SelectPageByCond(query, &pager.Records)
	if err != nil { return pager, err }
	pager.Total = total
	return pager, nil
}

// QueryPageByBSON 根据条件查询分页数据
func (r Repository[M, T]) QueryPageByBSON(query BSONPageQuery) (databasecloud.Pager[T], error) {
	pager := databasecloud.Pager[T]{Number: query.Number, Size: query.Size}; total, err := r.mapper.SelectPageByBSON(query, &pager.Records)
	if err != nil { return pager, err }
	pager.Total = total
	return pager, nil
}

// QueryPageWithOptions 根据条件查询分页数据
func (r Repository[M, T]) QueryPageWithOptions(query FilterPageQuery) (databasecloud.Pager[T], error) {
	pager := databasecloud.Pager[T]{Number: query.Number, Size: query.Size}; total, err := r.mapper.SelectPageWithOptions(query, &pager.Records)
	if err != nil { return pager, err }
	pager.Total = total
	return pager, nil
}

// CountByCond 通过实体条件统计数据。
func (r Repository[M, T]) CountByCond(query CondQuery[T]) (int64, error) {
	return r.mapper.CountByCond(query)
}

// CountByBSON 通过 BSON 条件统计数据。
func (r Repository[M, T]) CountByBSON(query BSONQuery) (int64, error) {
	return r.mapper.CountByBSON(query)
}

// CountWithOptions 根据条件查询统计数据
func (r Repository[M, T]) CountWithOptions(filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	return r.mapper.CountWithOptions(filter, opts...)
}

// ModifyByID 根据 ID 修改数据
func (r Repository[M, T]) ModifyByID(update *T, id any, notObjectID ...bool) (int64, error) {
	return r.mapper.UpdateByID(update, id, notObjectID...)
}

// ModifyByIDWithBSON 根据 ID 使用 BSON 文档修改数据
func (r Repository[M, T]) ModifyByIDWithBSON(update bson.M, id any, notObjectID ...bool) (int64, error) {
	return r.mapper.UpdateByIDWithBSON(update, id, notObjectID...)
}

// ModifyOneByCond 根据条件修改一条数据
func (r Repository[M, T]) ModifyOneByCond(update *T, condition T) (int64, error) {
	return r.mapper.UpdateOneByCond(update, condition)
}

// ModifyByCond 根据条件修改数据
func (r Repository[M, T]) ModifyByCond(update *T, condition T) (int64, error) {
	return r.mapper.UpdateByCond(update, condition)
}

// ModifyOneByBSON 根据条件修改一条数据
func (r Repository[M, T]) ModifyOneByBSON(update, condition bson.M) (int64, error) {
	return r.mapper.UpdateOneByBSON(update, condition)
}

// ModifyByBSON 根据条件修改数据
func (r Repository[M, T]) ModifyByBSON(update, condition bson.M) (int64, error) {
	return r.mapper.UpdateByBSON(update, condition)
}

// ModifyOneWithOptions 使用原生 UpdateOneOptions 修改单条数据。
func (r Repository[M, T]) ModifyOneWithOptions(filter, update any, opts ...options.Lister[options.UpdateOneOptions]) (int64, error) {
	return r.mapper.UpdateOneWithOptions(filter, update, opts...)
}

// ModifyWithOptions 使用原生 UpdateManyOptions 修改多条数据。
func (r Repository[M, T]) ModifyWithOptions(filter, update any, opts ...options.Lister[options.UpdateManyOptions]) (int64, error) {
	return r.mapper.UpdateWithOptions(filter, update, opts...)
}

// RemoveByID 根据 ID 删除数据
func (r Repository[M, T]) RemoveByID(id any, notObjectID ...bool) (int64, error) {
	return r.mapper.DeleteByID(id, notObjectID...)
}

// RemoveByIDs 根据多个 ID 删除数据。
func (r Repository[M, T]) RemoveByIDs(ids []any, notObjectID ...bool) (int64, error) {
	return r.mapper.DeleteByIDs(ids, notObjectID...)
}

// RemoveOneByCond 根据条件删除一条数据
func (r Repository[M, T]) RemoveOneByCond(condition T) (int64, error) {
	return r.mapper.DeleteOneByCond(condition)
}

// RemoveByCond 根据条件删除数据
func (r Repository[M, T]) RemoveByCond(condition T) (int64, error) {
	return r.mapper.DeleteByCond(condition)
}

// RemoveOneByBSON 根据条件删除一条数据
func (r Repository[M, T]) RemoveOneByBSON(condition bson.M) (int64, error) {
	return r.mapper.DeleteOneByBSON(condition)
}

// RemoveByBSON 根据条件删除数据
func (r Repository[M, T]) RemoveByBSON(condition bson.M) (int64, error) {
	return r.mapper.DeleteByBSON(condition)
}

// RemoveOneWithOptions 使用原生 DeleteOneOptions 删除单条数据。
func (r Repository[M, T]) RemoveOneWithOptions(filter any, opts ...options.Lister[options.DeleteOneOptions]) (int64, error) {
	return r.mapper.DeleteOneWithOptions(filter, opts...)
}

// RemoveWithOptions 使用原生 DeleteManyOptions 删除多条数据。
func (r Repository[M, T]) RemoveWithOptions(filter any, opts ...options.Lister[options.DeleteManyOptions]) (int64, error) {
	return r.mapper.DeleteWithOptions(filter, opts...)
}
