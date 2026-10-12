package mongo

import (
	"github.com/golang-acexy/cloud-database/databasecloud"
	"github.com/golang-acexy/starter-mongo/mongostarter"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Repository allocates caller-owned query results and delegates persistence to its mapper.
// Initialize with NewRepository; concurrency depends on the mapper and borrowed inputs.
// Database pools and clients remain starter-owned.
type Repository[M mongostarter.Mapper[T], T mongostarter.Model] struct {
	mapper M
}

// QueryOptions aliases the starter type; referenced inputs retain its ownership and mutation rules.
type QueryOptions = mongostarter.QueryOptions

// PageOptions aliases the starter type; referenced inputs retain its ownership and mutation rules.
type PageOptions = mongostarter.PageOptions

// CondQuery aliases the starter type; referenced inputs retain its ownership and mutation rules.
type CondQuery[T mongostarter.Model] = mongostarter.CondQuery[T]

// BSONQuery aliases the starter type; referenced inputs retain its ownership and mutation rules.
type BSONQuery = mongostarter.BSONQuery

// BSONPageQuery aliases the starter type; referenced inputs retain its ownership and mutation rules.
type BSONPageQuery = mongostarter.BSONPageQuery

// FilterPageQuery aliases the starter type; referenced inputs retain its ownership and mutation rules.
type FilterPageQuery = mongostarter.FilterPageQuery

// PageQuery aliases the starter type; referenced inputs retain its ownership and mutation rules.
type PageQuery[T mongostarter.Model] = mongostarter.PageQuery[T]

// NewRepository stores the mapper without validating it; start MongoStarter before operations.
func NewRepository[M mongostarter.Mapper[T], T mongostarter.Model](mapper M) Repository[M, T] {
	return Repository[M, T]{mapper: mapper}
}

// RawMapper returns the concrete mapper; underlying resources remain starter-owned.
func (r Repository[M, T]) RawMapper() M {
	return r.mapper
}

// Collection returns the native collection; the standard mapper returns nil without an active client.
func (r Repository[M, T]) Collection() *mongo.Collection {
	return r.mapper.Collection()
}

// Save inserts a document and returns a formatted ID; ObjectIDs become hexadecimal strings.
func (r Repository[M, T]) Save(entity *T) (string, error) {
	return r.mapper.Insert(entity)
}

// SaveWithBSON inserts a document and returns a formatted ID; ObjectIDs become hexadecimal strings.
func (r Repository[M, T]) SaveWithBSON(entity bson.M) (string, error) {
	return r.mapper.InsertWithBSON(entity)
}

// SaveWithOptions inserts a document and returns a formatted ID; ObjectIDs become hexadecimal strings.
func (r Repository[M, T]) SaveWithOptions(document any, opts ...options.Lister[options.InsertOneOptions]) (string, error) {
	return r.mapper.InsertWithOptions(document, opts...)
}

// SaveBatch inserts documents and returns formatted IDs; errors can follow partial writes without rollback or returned IDs.
func (r Repository[M, T]) SaveBatch(entities []*T) ([]string, error) {
	return r.mapper.InsertBatch(entities)
}

// SaveBatchWithBSON inserts documents and returns formatted IDs; errors can follow partial writes without rollback or returned IDs.
func (r Repository[M, T]) SaveBatchWithBSON(entities bson.A) ([]string, error) {
	return r.mapper.InsertBatchWithBSON(entities)
}

// SaveBatchWithOptions inserts documents and returns formatted IDs; errors can follow partial writes without rollback or returned IDs.
func (r Repository[M, T]) SaveBatchWithOptions(documents any, opts ...options.Lister[options.InsertManyOptions]) ([]string, error) {
	return r.mapper.InsertBatchWithOptions(documents, opts...)
}

// QueryByID allocates one result and preserves errors, including mongo.ErrNoDocuments.
// Only the first notObjectID flag is used; true preserves ordinary string IDs, otherwise strings must be ObjectID hex.
func (r Repository[M, T]) QueryByID(id any, notObjectID ...bool) (*T, error) {
	result := new(T)
	if err := r.mapper.SelectByID(id, result, notObjectID...); err != nil {
		return nil, err
	}
	return result, nil
}

// QueryByIDs allocates a result slice; errors may accompany partially populated results.
// Only the first notObjectID flag is used; true preserves ordinary string IDs, otherwise strings must be ObjectID hex.
func (r Repository[M, T]) QueryByIDs(ids []any, notObjectID ...bool) ([]*T, error) {
	result := make([]*T, 0)
	err := r.mapper.SelectByIDs(ids, &result, notObjectID...)
	return result, err
}

// ExistsByID checks existence and returns database errors separately from absence.
// Only the first notObjectID flag is used; true preserves ordinary string IDs, otherwise strings must be ObjectID hex.
func (r Repository[M, T]) ExistsByID(id any, notObjectID ...bool) (bool, error) {
	return r.mapper.ExistsByID(id, notObjectID...)
}

// QueryOneByCond allocates one result and preserves errors, including mongo.ErrNoDocuments.
func (r Repository[M, T]) QueryOneByCond(query CondQuery[T]) (*T, error) {
	result := new(T)
	if err := r.mapper.SelectOneByCond(query, result); err != nil {
		return nil, err
	}
	return result, nil
}

// QueryByCond allocates a result slice; errors may accompany partially populated results.
func (r Repository[M, T]) QueryByCond(query CondQuery[T]) ([]*T, error) {
	result := make([]*T, 0)
	err := r.mapper.SelectByCond(query, &result)
	return result, err
}

// QueryOneByBSON allocates one result and preserves errors, including mongo.ErrNoDocuments.
func (r Repository[M, T]) QueryOneByBSON(query BSONQuery) (*T, error) {
	result := new(T)
	if err := r.mapper.SelectOneByBSON(query, result); err != nil {
		return nil, err
	}
	return result, nil
}

// QueryByBSON allocates a result slice; errors may accompany partially populated results.
func (r Repository[M, T]) QueryByBSON(query BSONQuery) ([]*T, error) {
	result := make([]*T, 0)
	err := r.mapper.SelectByBSON(query, &result)
	return result, err
}

// QueryOneWithOptions allocates one result and preserves errors, including mongo.ErrNoDocuments.
func (r Repository[M, T]) QueryOneWithOptions(filter any, opts ...options.Lister[options.FindOneOptions]) (*T, error) {
	result := new(T)
	if err := r.mapper.SelectOneWithOptions(filter, result, opts...); err != nil {
		return nil, err
	}
	return result, nil
}

// QueryWithOptions allocates a result slice; errors may accompany partially populated results.
func (r Repository[M, T]) QueryWithOptions(filter any, opts ...options.Lister[options.FindOptions]) ([]*T, error) {
	result := make([]*T, 0)
	err := r.mapper.SelectWithOptions(filter, &result, opts...)
	return result, err
}

// QueryPageByCond allocates page output and delegates separate count/data operations without an automatic snapshot.
// Records may be nil; check errors before using partial records or metadata.
func (r Repository[M, T]) QueryPageByCond(query PageQuery[T]) (databasecloud.Pager[T], error) {
	pager := databasecloud.Pager[T]{Number: query.Number, Size: query.Size}
	total, err := r.mapper.SelectPageByCond(query, &pager.Records)
	if err != nil {
		return pager, err
	}
	pager.Total = total
	return pager, nil
}

// QueryPageByBSON allocates page output and delegates separate count/data operations without an automatic snapshot.
// Records may be nil; check errors before using partial records or metadata.
func (r Repository[M, T]) QueryPageByBSON(query BSONPageQuery) (databasecloud.Pager[T], error) {
	pager := databasecloud.Pager[T]{Number: query.Number, Size: query.Size}
	total, err := r.mapper.SelectPageByBSON(query, &pager.Records)
	if err != nil {
		return pager, err
	}
	pager.Total = total
	return pager, nil
}

// QueryPageWithOptions allocates page output and delegates separate count/data operations without an automatic snapshot.
// Records may be nil; check errors before using partial records or metadata.
func (r Repository[M, T]) QueryPageWithOptions(query FilterPageQuery) (databasecloud.Pager[T], error) {
	pager := databasecloud.Pager[T]{Number: query.Number, Size: query.Size}
	total, err := r.mapper.SelectPageWithOptions(query, &pager.Records)
	if err != nil {
		return pager, err
	}
	pager.Total = total
	return pager, nil
}

// CountByCond delegates counting to the mapper and preserves errors.
func (r Repository[M, T]) CountByCond(query CondQuery[T]) (int64, error) {
	return r.mapper.CountByCond(query)
}

// CountByBSON delegates counting to the mapper and preserves errors.
func (r Repository[M, T]) CountByBSON(query BSONQuery) (int64, error) {
	return r.mapper.CountByBSON(query)
}

// CountWithOptions delegates counting to the mapper and preserves errors.
func (r Repository[M, T]) CountWithOptions(filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	return r.mapper.CountWithOptions(filter, opts...)
}

// ModifyByID delegates updates and returns ModifiedCount, which can be zero for unchanged matches.
// Unchanged matches can return zero; filter guards use default BSON encoding, not configured BSONOptions.
// Only the first notObjectID flag is used; true preserves ordinary string IDs, otherwise strings must be ObjectID hex.
func (r Repository[M, T]) ModifyByID(update *T, id any, notObjectID ...bool) (int64, error) {
	return r.mapper.UpdateByID(update, id, notObjectID...)
}

// ModifyByIDWithBSON delegates updates and returns ModifiedCount, which can be zero for unchanged matches.
// Unchanged matches can return zero; filter guards use default BSON encoding, not configured BSONOptions.
// Only the first notObjectID flag is used; true preserves ordinary string IDs, otherwise strings must be ObjectID hex.
func (r Repository[M, T]) ModifyByIDWithBSON(update bson.M, id any, notObjectID ...bool) (int64, error) {
	return r.mapper.UpdateByIDWithBSON(update, id, notObjectID...)
}

// ModifyOneByCond updates one typed-filter match with $set and returns ModifiedCount.
// Unchanged matches can return zero; filter guards use default BSON encoding, not configured BSONOptions.
func (r Repository[M, T]) ModifyOneByCond(update *T, condition T) (int64, error) {
	return r.mapper.UpdateOneByCond(update, condition)
}

// ModifyByCond updates typed-filter matches with $set and returns ModifiedCount.
// Unchanged matches can return zero; filter guards use default BSON encoding, not configured BSONOptions.
func (r Repository[M, T]) ModifyByCond(update *T, condition T) (int64, error) {
	return r.mapper.UpdateByCond(update, condition)
}

// ModifyOneByBSON updates one BSON-filter match; plain assignments use $set and operator documents pass through.
// Unchanged matches can return zero; filter guards use default BSON encoding, not configured BSONOptions.
func (r Repository[M, T]) ModifyOneByBSON(update, condition bson.M) (int64, error) {
	return r.mapper.UpdateOneByBSON(update, condition)
}

// ModifyByBSON updates BSON-filter matches; plain assignments use $set and operator documents pass through.
// Unchanged matches can return zero; filter guards use default BSON encoding, not configured BSONOptions.
func (r Repository[M, T]) ModifyByBSON(update, condition bson.M) (int64, error) {
	return r.mapper.UpdateByBSON(update, condition)
}

// ModifyOneWithOptions updates one native-filter match using driver options and returns ModifiedCount.
// Unchanged matches can return zero; filter guards use default BSON encoding, not configured BSONOptions.
func (r Repository[M, T]) ModifyOneWithOptions(filter, update any, opts ...options.Lister[options.UpdateOneOptions]) (int64, error) {
	return r.mapper.UpdateOneWithOptions(filter, update, opts...)
}

// ModifyWithOptions updates native-filter matches using driver options and returns ModifiedCount.
// Unchanged matches can return zero; filter guards use default BSON encoding, not configured BSONOptions.
func (r Repository[M, T]) ModifyWithOptions(filter, update any, opts ...options.Lister[options.UpdateManyOptions]) (int64, error) {
	return r.mapper.UpdateWithOptions(filter, update, opts...)
}

// RemoveByID deletes one ID and returns DeletedCount.
// Only the first notObjectID flag is used; true preserves ordinary string IDs, otherwise strings must be ObjectID hex.
func (r Repository[M, T]) RemoveByID(id any, notObjectID ...bool) (int64, error) {
	return r.mapper.DeleteByID(id, notObjectID...)
}

// RemoveByIDs deletes an ID batch and returns DeletedCount; empty input is rejected.
// Only the first notObjectID flag is used; true preserves ordinary string IDs, otherwise strings must be ObjectID hex.
func (r Repository[M, T]) RemoveByIDs(ids []any, notObjectID ...bool) (int64, error) {
	return r.mapper.DeleteByIDs(ids, notObjectID...)
}

// RemoveOneByCond deletes one typed-filter match; custom BSON encoding can affect the empty-filter guard.
func (r Repository[M, T]) RemoveOneByCond(condition T) (int64, error) {
	return r.mapper.DeleteOneByCond(condition)
}

// RemoveByCond deletes typed-filter matches; custom BSON encoding can affect the empty-filter guard.
func (r Repository[M, T]) RemoveByCond(condition T) (int64, error) {
	return r.mapper.DeleteByCond(condition)
}

// RemoveOneByBSON deletes one BSON-filter match; an empty map is rejected.
func (r Repository[M, T]) RemoveOneByBSON(condition bson.M) (int64, error) {
	return r.mapper.DeleteOneByBSON(condition)
}

// RemoveByBSON deletes BSON-filter matches; an empty map is rejected.
func (r Repository[M, T]) RemoveByBSON(condition bson.M) (int64, error) {
	return r.mapper.DeleteByBSON(condition)
}

// RemoveOneWithOptions deletes one native-filter match using driver options and returns DeletedCount.
func (r Repository[M, T]) RemoveOneWithOptions(filter any, opts ...options.Lister[options.DeleteOneOptions]) (int64, error) {
	return r.mapper.DeleteOneWithOptions(filter, opts...)
}

// RemoveWithOptions deletes native-filter matches using driver options and returns DeletedCount.
func (r Repository[M, T]) RemoveWithOptions(filter any, opts ...options.Lister[options.DeleteManyOptions]) (int64, error) {
	return r.mapper.DeleteWithOptions(filter, opts...)
}
