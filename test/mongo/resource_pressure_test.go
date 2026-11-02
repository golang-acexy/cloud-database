package mongo

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	repository "github.com/golang-acexy/cloud-database/databasecloud/mongo"
	"github.com/golang-acexy/starter-mongo/mongostarter"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var auditCollection = fmt.Sprintf("cloud_database_audit_%d", time.Now().UnixNano())

type auditRow struct {
	ID      int64  `bson:"_id"`
	Payload string `bson:"payload"`
}

// CollectionName isolates all audit operations from application collections.
func (auditRow) CollectionName() string { return auditCollection }

// TestRepositoryResourceAudit uses only a unique temporary collection and must run alone.
func TestRepositoryResourceAudit(t *testing.T) {
	if os.Getenv("CLOUD_DATABASE_AUDIT") != "1" {
		t.Skip("set CLOUD_DATABASE_AUDIT=1")
	}
	repo := repository.NewRepository[mongostarter.BaseMapper[auditRow], auditRow](mongostarter.BaseMapper[auditRow]{})
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := repo.Collection().Drop(ctx); err != nil {
			t.Error(err)
		}
	})
	rows := make([]*auditRow, 128)
	for i := range rows {
		rows[i] = &auditRow{ID: int64(i + 1), Payload: strings.Repeat("x", 8192)}
	}
	if _, err := repo.SaveBatch(rows); err != nil {
		t.Fatal(err)
	}
	rows = nil
	sample := func(label string) runtime.MemStats {
		runtime.GC()
		runtime.GC()
		var s runtime.MemStats
		runtime.ReadMemStats(&s)
		t.Logf("%s heap=%d goroutines=%d", label, s.HeapAlloc, runtime.NumGoroutine())
		return s
	}
	for i := 0; i < 20; i++ {
		if _, err := repo.QueryByID(1); err != nil {
			t.Fatal(err)
		}
	}
	sample("warm")
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
		start := time.Now()
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
					page, err := repo.QueryPageByBSON(repository.BSONPageQuery{Condition: bson.M{}, PageOptions: repository.PageOptions{Number: 1, Size: 16}})
					if err != nil || len(page.Records) != 16 || page.Total != 128 {
						failures <- fmt.Errorf("page: %v", err)
						return
					}
				}
			}()
		}
		wg.Wait()
		close(failures)
		for err := range failures {
			t.Fatal(err)
		}
		elapsed := time.Since(start)
		after := sample(fmt.Sprintf("round%d_after", round))
		t.Logf("round=%d workflows=2000 duration=%s allocated_bytes/workflow=%d", round, elapsed, (after.TotalAlloc-before.TotalAlloc)/2000)
	}
	for _, columns := range [][]string{nil, {"_id"}} {
		before := sample("projection_before")
		started := time.Now()
		for i := 0; i < 100; i++ {
			p, err := repo.QueryPageByBSON(repository.BSONPageQuery{Condition: bson.M{}, PageOptions: repository.PageOptions{Number: 1, Size: 16, SelectColumns: columns}})
			if err != nil || len(p.Records) != 16 {
				t.Fatalf("projection: %v", err)
			}
		}
		elapsed := time.Since(started)
		after := sample("projection_after")
		t.Logf("projection=%v bytes/page=%d duration/100pages=%s", columns, (after.TotalAlloc-before.TotalAlloc)/100, elapsed)
	}
	// Invalid decoding must also release a cursor, including a multi-batch result.

	for i := 0; i < 50; i++ {
		var invalid []int
		cursor, err := repo.Collection().Find(context.Background(), bson.M{}, options.Find().SetBatchSize(1))
		if err != nil {
			t.Fatal(err)
		}
		err = cursor.All(context.Background(), &invalid)
		if err == nil {
			t.Fatal("expected decode error")
		}
		if cursor.ID() != 0 {
			t.Fatal("cursor retained after decode failure")
		}
	}
	full, err := repo.QueryByBSON(repository.BSONQuery{})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("unlimited list returned=%d", len(full))
	if len(full) != 128 {
		t.Fatal("wrong list size")
	}
	full = nil
	sample("after_decode_errors")
}
