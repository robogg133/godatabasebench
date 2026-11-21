package stdlib

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/robogg133/godatabasebench/internal/bench"
)

func BenchmarkPointReadByID(b *testing.B) {
	ctx := context.Background()
	db, err := bench.NewStdlibPool(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var id int64
		var email, name string
		if err := db.QueryRowContext(ctx, "SELECT id, email, name FROM users WHERE id = $1", 1).Scan(&id, &email, &name); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPointReadByEmail(b *testing.B) {
	ctx := context.Background()
	db, _ := bench.NewStdlibPool(ctx)
	defer db.Close()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var id int64
		var email, name string
		if err := db.QueryRowContext(ctx, "SELECT id, email, name FROM users WHERE email = $1", "user1@example.com").Scan(&id, &email, &name); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkListPostsByUser(b *testing.B) {
	ctx := context.Background()
	db, _ := bench.NewStdlibPool(ctx)
	defer db.Close()
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		rows, err := db.QueryContext(ctx, "SELECT id, user_id, title FROM posts WHERE user_id = $1 LIMIT 50", 1)
		if err != nil {
			b.Fatal(err)
		}
		rows.Close()
	}
}

func BenchmarkInsertUser(b *testing.B) {
	ctx := context.Background()
	db, _ := bench.NewStdlibPool(ctx)
	defer db.Close()
	base := time.Now().UnixNano()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := db.ExecContext(ctx, "INSERT INTO users (email, name) VALUES ($1,$2)", fmt.Sprintf("bench-%d-%d@example.com", base, i), "Bench")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateUser(b *testing.B) {
	ctx := context.Background()
	db, _ := bench.NewStdlibPool(ctx)
	defer db.Close()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := db.ExecContext(ctx, "UPDATE users SET name=$2 WHERE id=$1", 1, "Updated"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBatchInsertUsers(b *testing.B) {
	ctx := context.Background()
	db, _ := bench.NewStdlibPool(ctx)
	defer db.Close()
	base := time.Now().UnixNano()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tx, _ := db.BeginTx(ctx, nil)
		for j := 0; j < 100; j++ {
			_, _ = tx.ExecContext(ctx, "INSERT INTO users (email,name) VALUES ($1,$2)", fmt.Sprintf("batch-%d-%d-%d@example.com", base, i, j), "Batch")
		}
		_ = tx.Commit()
	}
}

var _ = sql.DB{}
