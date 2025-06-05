package databasecloud

import "github.com/golang-acexy/starter-gorm/gormstarter"

type GormDatabaseService[B gormstarter.IBaseMapper[M, T], M gormstarter.BaseMapper[T], T gormstarter.IBaseModel] interface {
	WithTxMapper() M
}

type GormDBService[B gormstarter.IBaseMapper[M, T], M gormstarter.BaseMapper[T], T gormstarter.IBaseModel] struct {
	dbService B
}

func (s GormDBService[B, M, T]) QueryByID(id any, result *T) (int64, error) {
	return s.dbService.SelectById(id, result)
}
