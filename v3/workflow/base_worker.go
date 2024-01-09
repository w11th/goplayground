package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tnclong/go-que"
	"github.com/w11th/goplayground/v2/startstop"
)

type BaseWorker struct {
	startstop.BaseStartStop

	NewQueWorkerFunc func() (*que.Worker, error)
}

func NewBaseWorker(
	NewQueWorkerFunc func() (*que.Worker, error),
) *BaseWorker {
	return &BaseWorker{
		NewQueWorkerFunc: NewQueWorkerFunc,
	}
}

func (b *BaseWorker) Start(ctx context.Context) error {
	ctx, shouldStart, stopped := b.StartInit(ctx)
	if !shouldStart {
		return nil
	}

	go func() {
		defer close(stopped)

		for {
			select {
			case <-ctx.Done():
				break
			default:
			}

			queWorker, err := b.NewQueWorkerFunc()
			if err != nil {
				panic(err)
			}

			err = b.runQueWorker(ctx, queWorker)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					break
				}

				fmt.Printf("error notifier: %s\n", err)
			}

			fmt.Printf("restart worker after 5 sec...")
			time.Sleep(5 * time.Second)
		}
	}()

	return nil
}

func (b *BaseWorker) runQueWorker(ctx context.Context, queWorker *que.Worker) error {
	c := make(chan error, 1)
	go func() { c <- queWorker.Run() }()

	select {
	case <-ctx.Done():
		queWorker.Stop(context.Background()) // Can add timeout for Stop
		err := <-c                           // Wait for queWorker.Run to return.
		fmt.Printf("que worker err: %s\n", err)
		return ctx.Err()
	case err := <-c:
		return err
	}
}
