package main

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/tnclong/go-que/pg"
	"github.com/w11th/goplayground/internal/cmd/workertests/v2/workflow"
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

	ctx := context.Background()

	worker := workflow.NewWorker(queue)
	err = worker.Start(ctx)
	if err != nil {
		panic(err)
	}
	defer worker.Stop()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}
