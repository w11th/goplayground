package main

import (
	"context"
	"database/sql"

	"github.com/tnclong/go-que"
	"github.com/tnclong/go-que/pg"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	db, err := sql.Open("pgx", "postgres://test:123@localhost:5412/test_dev")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	queue, err := pg.New(db)
	if err != nil {
		panic(err)
	}

	queue.Enqueue(context.Background(), nil, que.Plan{
		Queue: "test",
	})
}
