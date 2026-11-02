package gorm

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-acexy/cloud-database/databasecloud/rds"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var auditTable = fmt.Sprintf("cloud_database_audit_%d", time.Now().UnixNano())

type auditRow struct {
	ID      int64
	Payload string `gorm:"type:text"`
}

// TableName isolates all audit operations from application tables.
func (auditRow) TableName() string { return auditTable }

type auditMapper struct {
	gormstarter.BaseMapper[auditRow]
}

// WithTxMapper returns a transaction-bound copy.
func (m auditMapper) WithTxMapper(tx *gorm.DB) auditMapper {
	return auditMapper{m.GetBaseMapperWithTx(tx)}
}

type auditRepo struct {
	rds.Repository[auditRepo, auditMapper, auditRow]
}

func newAuditRepo() auditRepo {
	return auditRepo{rds.NewRepository(auditMapper{}, func(r rds.Repository[auditRepo, auditMapper, auditRow]) auditRepo { return auditRepo{r} })}
}

// TestRepositoryResourceAudit uses only a unique temporary table and must run alone.
func TestRepositoryResourceAudit(t *testing.T) {
	if os.Getenv("CLOUD_DATABASE_AUDIT") != "1" {
		t.Skip("set CLOUD_DATABASE_AUDIT=1")
	}
	repo := newAuditRepo()
	db := repo.CurrentGormDB()
	// Bind a quiet independent session without changing the starter's shared configuration.
	repo = repo.WithTxRepo(db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}))
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Migrator().CreateTable(&auditRow{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Migrator().DropTable(&auditRow{}); err != nil {
			t.Error(err)
		}
	})
	for i := int64(1); i <= 128; i++ {
		if _, err := repo.Save(&auditRow{ID: i, Payload: strings.Repeat("x", 8192)}); err != nil {
			t.Fatal(err)
		}
	}
	sample := func(label string) runtime.MemStats {
		runtime.GC()
		runtime.GC()
		var s runtime.MemStats
		runtime.ReadMemStats(&s)
		t.Logf("%s heap=%d goroutines=%d in_use=%d idle=%d", label, s.HeapAlloc, runtime.NumGoroutine(), pool.Stats().InUse, pool.Stats().Idle)
		return s
	}
	for i := 0; i < 20; i++ {
		if _, err := repo.QueryByID(1); err != nil {
			t.Fatal(err)
		}
	}
	sample("warm")
	rollback := errors.New("audit rollback")
	rounds := 6
	if value := os.Getenv("CLOUD_DATABASE_AUDIT_ROUNDS"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			t.Fatal("CLOUD_DATABASE_AUDIT_ROUNDS must be between 1 and 100")
		}
		rounds = parsed
	}
	for round := 0; round < rounds; round++ {

		before := sample(fmt.Sprintf("round%d_before", round))
		started := time.Now()
		var wg sync.WaitGroup
		failures := make(chan error, 8)
		for worker := 0; worker < 8; worker++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 250; i++ {
					row, err := repo.QueryByID(i%128 + 1)
					if err != nil || row == nil {
						failures <- fmt.Errorf("query: %v", err)
						return
					}
					page, err := repo.QueryPageByMap(rds.MapPageQuery{PageOptions: rds.PageOptions{Number: 1, Size: 16}})
					if err != nil || len(page.Records) != 16 || page.Total != 128 {
						failures <- fmt.Errorf("page: %v", err)
						return
					}
					if i%8 == 0 {
						err = repo.Transaction(func(tx auditRepo) error {
							_, e := tx.QueryByID(1)
							if e != nil {
								return e
							}
							return rollback
						})
						if !errors.Is(err, rollback) {
							failures <- fmt.Errorf("transaction: %v", err)
							return
						}
					}
				}
			}()
		}
		wg.Wait()
		close(failures)
		for err := range failures {
			t.Fatal(err)
		}
		elapsed := time.Since(started)
		after := sample(fmt.Sprintf("round%d_after", round))
		t.Logf("round=%d workflows=2000 duration=%s allocated_bytes/workflow=%d", round, elapsed, (after.TotalAlloc-before.TotalAlloc)/2000)
		if pool.Stats().InUse != 0 {
			t.Fatal("connections remain checked out")
		}
	}
	if err := repo.Transaction(func(tx auditRepo) error { _, err := tx.QueryByID(1); return err }); err != nil {
		t.Fatal(err)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("panic not propagated")
			}
		}()
		_ = repo.Transaction(func(auditRepo) error { panic("audit") })
	}()
	if pool.Stats().InUse != 0 {
		t.Fatal("normal transaction paths retained connections")
	}
	// Verify abnormal exits release their connections before returning control.
	var abandoned *gorm.DB
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = repo.Transaction(func(tx auditRepo) error { abandoned = tx.CurrentGormDB(); runtime.Goexit(); return nil })
	}()
	<-done
	t.Logf("Goexit checked_out=%d", pool.Stats().InUse)
	if pool.Stats().InUse != 0 {
		t.Error("Goexit retained a connection")
	}
	if abandoned == nil {
		t.Fatal("missing Goexit transaction")
	}
	if err := abandoned.Rollback().Error; !errors.Is(err, sql.ErrTxDone) {
		t.Fatal(err)
	}
	broken := rds.NewRepository(auditMapper{}, func(base rds.Repository[auditRepo, auditMapper, auditRow]) auditRepo {
		abandoned = base.CurrentGormDB()
		panic("factory")
	})
	func() {
		defer func() {
			if recover() == nil {
				t.Error("factory panic not propagated")
			}
		}()
		_ = broken.NewTxRepo()
	}()
	t.Logf("NewTxRepo factory panic checked_out=%d", pool.Stats().InUse)
	if pool.Stats().InUse != 0 {
		t.Error("factory panic retained a connection")
	}
	if err := abandoned.Rollback().Error; !errors.Is(err, sql.ErrTxDone) {
		t.Fatal(err)
	}
	sample("after_abnormal_cleanup")
	nativeDone := make(chan struct{})
	go func() {
		defer close(nativeDone)
		_ = db.Transaction(func(*gorm.DB) error { runtime.Goexit(); return nil })
	}()
	<-nativeDone
	t.Logf("native GORM Goexit checked_out=%d", pool.Stats().InUse)
	if pool.Stats().InUse != 0 {
		t.Fatal("native transaction retained connection")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("managed factory panic not propagated")
			}
		}()
		_ = broken.Transaction(func(auditRepo) error { return nil })
	}()
	if pool.Stats().InUse != 0 {
		t.Fatal("managed factory panic retained connection")
	}

	if pool.Stats().InUse != 0 {
		t.Fatal("audit cleanup retained connection")
	}
	for _, columns := range [][]string{nil, {"id"}} {
		before := sample("projection_before")
		started := time.Now()
		for i := 0; i < 100; i++ {
			p, err := repo.QueryPageByMap(rds.MapPageQuery{PageOptions: rds.PageOptions{Number: 1, Size: 16, SelectColumns: columns}})
			if err != nil || len(p.Records) != 16 {
				t.Fatalf("projection: %v", err)
			}
		}
		elapsed := time.Since(started)
		after := sample("projection_after")
		t.Logf("projection=%v bytes/page=%d duration/100pages=%s", columns, (after.TotalAlloc-before.TotalAlloc)/100, elapsed)
	}
	// Callback pagination intentionally demonstrates the current metadata-only contract.

	page, err := repo.QueryPageByGorm(rds.GormPageQuery{PageOptions: rds.PageOptions{Number: 1, Size: 16}, CountRawDB: func(*gorm.DB) {}, PageRawDB: func(*gorm.DB) {}})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("callback page size=%d returned=%d", page.Size, len(page.Records))
	if len(page.Records) != 128 {
		t.Fatal("unexpected callback pagination behavior")
	}
}
