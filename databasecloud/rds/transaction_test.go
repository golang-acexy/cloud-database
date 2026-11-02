package rds

import (
	"context"
	"database/sql"
	"errors"
	"runtime"
	"testing"

	"github.com/golang-acexy/starter-gorm/gormstarter"
	"gorm.io/gorm"
)

type transactionModel struct{}

// TableName supplies the test model contract.
func (transactionModel) TableName() string { return "transaction_test" }

type transactionMapper struct {
	gormstarter.BaseMapper[transactionModel]
	copyHook func()
}

// WithTxMapper returns a copy and optionally simulates abnormal mapper construction.
func (m transactionMapper) WithTxMapper(tx *gorm.DB) transactionMapper {
	if m.copyHook != nil {
		m.copyHook()
	}
	m.BaseMapper = m.GetBaseMapperWithTx(tx)
	return m
}

type transactionRepo struct {
	Repository[transactionRepo, transactionMapper, transactionModel]
}
type transactionPool struct {
	gorm.ConnPool
	commits, rollbacks     int
	commitErr, rollbackErr error
}

// Commit records transaction completion and returns the configured error.
func (p *transactionPool) Commit() error { p.commits++; return p.commitErr }

// Rollback records cleanup and returns the configured error.
func (p *transactionPool) Rollback() error { p.rollbacks++; return p.rollbackErr }

type beginPool struct {
	gorm.ConnPool
	tx  *transactionPool
	err error
}

// BeginTx returns the configured transaction or begin error.
func (p *beginPool) BeginTx(context.Context, *sql.TxOptions) (gorm.ConnPool, error) {
	return p.tx, p.err
}

func testTransactionRepo(t *testing.T, pool *beginPool, copyHook, factoryHook func()) transactionRepo {
	t.Helper()
	db, err := gorm.Open(nil, &gorm.Config{ConnPool: pool, DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	mapper := transactionMapper{BaseMapper: gormstarter.BaseMapper[transactionModel]{}.GetBaseMapperWithTx(db), copyHook: copyHook}
	return transactionRepo{NewRepository(mapper, func(base Repository[transactionRepo, transactionMapper, transactionModel]) transactionRepo {
		if factoryHook != nil {
			factoryHook()
		}
		return transactionRepo{base}
	})}
}

// TestTransactionCompletion preserves normal completion and error handling.
func TestTransactionCompletion(t *testing.T) {
	callbackErr := errors.New("callback")
	cleanupErr := errors.New("cleanup")
	commitErr := errors.New("commit")
	beginErr := errors.New("begin")
	for _, tc := range []struct {
		name                                          string
		callbackErr, rollbackErr, commitErr, beginErr error
		commits, rollbacks                            int
	}{
		{name: "success", commits: 1},
		{name: "callback error", callbackErr: callbackErr, rollbacks: 1},
		{name: "rollback error", callbackErr: callbackErr, rollbackErr: cleanupErr, rollbacks: 1},
		{name: "commit error", commitErr: commitErr, commits: 1},
		{name: "begin error", beginErr: beginErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx := &transactionPool{commitErr: tc.commitErr, rollbackErr: tc.rollbackErr}
			repo := testTransactionRepo(t, &beginPool{tx: tx, err: tc.beginErr}, nil, nil)
			called := false
			err := repo.Transaction(func(transactionRepo) error { called = true; return tc.callbackErr })
			for _, want := range []error{tc.callbackErr, tc.rollbackErr, tc.commitErr, tc.beginErr} {
				if want != nil && !errors.Is(err, want) {
					t.Fatalf("missing error %v: %v", want, err)
				}
			}
			if tc.callbackErr == nil && tc.commitErr == nil && tc.beginErr == nil && err != nil {
				t.Fatal(err)
			}
			if called != (tc.beginErr == nil) {
				t.Fatal("unexpected callback invocation")
			}
			if tx.commits != tc.commits || tx.rollbacks != tc.rollbacks {
				t.Fatalf("commit=%d rollback=%d", tx.commits, tx.rollbacks)
			}
		})
	}
}

// TestTransactionAbnormalExit verifies cleanup without changing panic or Goexit.
func TestTransactionAbnormalExit(t *testing.T) {
	for _, manual := range []bool{false, true} {
		for _, stage := range []string{"mapper", "factory", "callback"} {
			if manual && stage == "callback" {
				continue
			}
			for _, goexit := range []bool{false, true} {
				name := stage + "/panic"
				if goexit {
					name = stage + "/goexit"
				}
				if manual {
					name = "manual/" + name
				}
				t.Run(name, func(t *testing.T) {
					tx := &transactionPool{}
					marker := errors.New("exit")
					exit := func() {
						if goexit {
							runtime.Goexit()
						}
						panic(marker)
					}
					var copyHook, factoryHook func()
					if stage == "mapper" {
						copyHook = exit
					}
					if stage == "factory" {
						factoryHook = exit
					}
					repo := testTransactionRepo(t, &beginPool{tx: tx}, copyHook, factoryHook)
					done := make(chan struct{})
					var recovered any
					returned := false
					go func() {
						defer close(done)
						defer func() { recovered = recover() }()
						if manual {
							repo.NewTxRepo()
						} else {
							_ = repo.Transaction(func(transactionRepo) error {
								if stage == "callback" {
									exit()
								}
								return nil
							})
						}
						returned = true
					}()
					<-done
					if returned {
						t.Fatal("abnormal exit returned")
					}
					if !goexit && recovered != marker {
						t.Fatalf("panic changed: %v", recovered)
					}
					if goexit && recovered != nil {
						t.Fatalf("Goexit became panic: %v", recovered)
					}
					if tx.commits != 0 || tx.rollbacks != 1 {
						t.Fatalf("commit=%d rollback=%d", tx.commits, tx.rollbacks)
					}
				})
			}
		}
	}
}

// TestNewTxRepoHandoff leaves a successfully returned transaction caller-owned.
func TestNewTxRepoHandoff(t *testing.T) {
	tx := &transactionPool{}
	repo := testTransactionRepo(t, &beginPool{tx: tx}, nil, nil)
	handed := repo.NewTxRepo()
	if tx.commits != 0 || tx.rollbacks != 0 {
		t.Fatal("transaction closed before caller handoff")
	}
	if err := handed.CurrentGormDB().Rollback().Error; err != nil {
		t.Fatal(err)
	}
	if tx.rollbacks != 1 {
		t.Fatal("caller could not roll back")
	}
}
