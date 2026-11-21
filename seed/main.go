package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	dsn := "postgres://bench:bench@localhost:5432/bench?sslmode=disable"
	if v := os.Getenv("DATABASE_URL"); v != "" {
		dsn = v
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	const n = 100000
	rows := make([][]any, 0, n)
	for i := 1; i <= n; i++ {
		rows = append(rows, []any{fmt.Sprintf("user%d@example.com", i), fmt.Sprintf("User %d", i)})
	}
	_, _ = pool.Exec(ctx, "TRUNCATE users, posts, comments, tags, post_tags CASCADE")
	copyCount, err := pool.CopyFrom(ctx, pgx.Identifier{"users"}, []string{"email", "name"}, pgx.CopyFromRows(rows))
	if err != nil {
		panic(err)
	}
	fmt.Printf("seeded %d users\n", copyCount)
}
