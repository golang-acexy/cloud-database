package rds

import "errors"

var (
	// ErrNilRepositoryFactory 表示未提供业务 Repository 工厂函数。
	ErrNilRepositoryFactory = errors.New("repository factory function is nil")
	// ErrRepositoryNotInitialized 表示 Repository 未通过构造函数初始化。
	ErrRepositoryNotInitialized = errors.New("repository is not initialized")
	// ErrNilTransactionFunc 表示事务执行函数为空。
	ErrNilTransactionFunc = errors.New("transaction function is nil")
)
