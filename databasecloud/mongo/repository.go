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

// PageQuery 定义 MongoDB 分页、排序、投影和原生查询选项。
type PageQuery struct {
	OrderBy        []*mongostarter.OrderBy
	SpecifyColumns []string
	FindOptions    []options.Lister[options.FindOptions]
	CountOptions   []options.Lister[options.CountOptions]
}

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
func (r Repository[M, T]) QueryByID(id any, result *T, notObjectID ...bool) error {
	return r.mapper.SelectByID(id, result, notObjectID...)
}

// QueryByIDs 根据多个 ID 查询数据，普通字符串 ID 需要将 notObjectID 设置为 true
func (r Repository[M, T]) QueryByIDs(ids []any, result *[]*T, notObjectID ...bool) (err error) {
	return r.mapper.SelectByIDs(ids, result, notObjectID...)
}

// ExistsByID 判断指定主键的数据是否存在。
func (r Repository[M, T]) ExistsByID(id any, notObjectID ...bool) (bool, error) {
	return r.mapper.ExistsByID(id, notObjectID...)
}

// QueryOneByCond 根据条件查询一条数据
func (r Repository[M, T]) QueryOneByCond(condition *T, result *T, specifyColumns ...string) error {
	return r.mapper.SelectOneByCond(condition, result, specifyColumns...)
}

// QueryByCond 根据条件查询数据
func (r Repository[M, T]) QueryByCond(condition *T, orderBy []*mongostarter.OrderBy, result *[]*T, specifyColumns ...string) error {
	return r.mapper.SelectByCond(condition, orderBy, result, specifyColumns...)
}

// QueryOneByBSON 根据条件查询一条数据
func (r Repository[M, T]) QueryOneByBSON(condition bson.M, result *T, specifyColumns ...string) error {
	return r.mapper.SelectOneByBSON(condition, result, specifyColumns...)
}

// QueryByBSON 根据条件查询数据
func (r Repository[M, T]) QueryByBSON(condition bson.M, orderBy []*mongostarter.OrderBy, result *[]*T, specifyColumns ...string) error {
	return r.mapper.SelectByBSON(condition, orderBy, result, specifyColumns...)
}

// QueryOneWithOptions 根据条件查询一条数据
func (r Repository[M, T]) QueryOneWithOptions(filter any, result *T, opts ...options.Lister[options.FindOneOptions]) error {
	return r.mapper.SelectOneWithOptions(filter, result, opts...)
}

// QueryWithOptions 根据条件查询数据
func (r Repository[M, T]) QueryWithOptions(filter any, result *[]*T, opts ...options.Lister[options.FindOptions]) error {
	return r.mapper.SelectWithOptions(filter, result, opts...)
}

// QueryPageByCond 根据条件查询分页数据
func (r Repository[M, T]) QueryPageByCond(condition *T, query PageQuery, pager *databasecloud.Pager[T]) error {
	total, err := r.mapper.SelectPageByCond(condition, mongostarter.PageQuery{PageNumber: pager.Number, PageSize: pager.Size, OrderBy: query.OrderBy, SpecifyColumns: query.SpecifyColumns, FindOptions: query.FindOptions, CountOptions: query.CountOptions}, &pager.Records)
	if err != nil {
		return err
	}
	pager.Total = total
	return nil
}

// QueryPageByBSON 根据条件查询分页数据
func (r Repository[M, T]) QueryPageByBSON(condition bson.M, query PageQuery, pager *databasecloud.Pager[T]) error {
	total, err := r.mapper.SelectPageByBSON(condition, mongostarter.PageQuery{PageNumber: pager.Number, PageSize: pager.Size, OrderBy: query.OrderBy, SpecifyColumns: query.SpecifyColumns, FindOptions: query.FindOptions, CountOptions: query.CountOptions}, &pager.Records)
	if err != nil {
		return err
	}
	pager.Total = total
	return nil
}

// QueryPageWithOptions 根据条件查询分页数据
func (r Repository[M, T]) QueryPageWithOptions(filter any, query PageQuery, pager *databasecloud.Pager[T]) error {
	total, err := r.mapper.SelectPageWithOptions(filter, mongostarter.PageQuery{PageNumber: pager.Number, PageSize: pager.Size, OrderBy: query.OrderBy, SpecifyColumns: query.SpecifyColumns, FindOptions: query.FindOptions, CountOptions: query.CountOptions}, &pager.Records)
	if err != nil {
		return err
	}
	pager.Total = total
	return nil
}

// CountByCond 根据条件查询统计数据
func (r Repository[M, T]) CountByCond(condition *T) (int64, error) {
	return r.mapper.CountByCond(condition)
}

// CountByBSON 根据条件查询统计数据
func (r Repository[M, T]) CountByBSON(condition bson.M) (int64, error) {
	return r.mapper.CountByBSON(condition)
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
func (r Repository[M, T]) ModifyOneByCond(update, condition *T) (int64, error) {
	return r.mapper.UpdateOneByCond(update, condition)
}

// ModifyByCond 根据条件修改数据
func (r Repository[M, T]) ModifyByCond(update, condition *T) (int64, error) {
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
func (r Repository[M, T]) RemoveOneByCond(condition *T) (int64, error) {
	return r.mapper.DeleteOneByCond(condition)
}

// RemoveByCond 根据条件删除数据
func (r Repository[M, T]) RemoveByCond(condition *T) (int64, error) {
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
