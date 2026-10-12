package rds

import "errors"

var (
	// ErrNilRepositoryFactory indicates that NewRepository received a nil factory.
	ErrNilRepositoryFactory = errors.New("repository factory function is nil")
	// ErrRepositoryNotInitialized indicates that Transaction has no initialized factory.
	ErrRepositoryNotInitialized = errors.New("repository is not initialized")
	// ErrNilTransactionFunc indicates that Transaction received a nil callback.
	ErrNilTransactionFunc = errors.New("transaction function is nil")
)
