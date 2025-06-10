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
	return m.Mapper.InsertByColl(document, opts...)
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
	return m.Mapper.InsertBatchByColl(documents, opts...)
}

// QueryByID 根据id查询数据 默认匹配_id字段如果重写了_id notObjectId应当设置为true
func (m MongoRepository[B, M, T]) QueryByID(id string, result *T, notObjectId ...bool) error {
	return m.Mapper.SelectById(id, result, notObjectId...)
}

func (m MongoRepository[B, M, T]) QueryByIDs(ids []string, result *[]*T) (err error) {
	return m.Mapper.SelectByIds(ids, result)
}
