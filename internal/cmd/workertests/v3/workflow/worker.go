package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tnclong/go-que"
	"github.com/w11th/goplayground/internal/cmd/workertests/v2/startstop"
)

type Worker struct {
	startstop.BaseStartStop
	workflowQueue que.Queue
}

func NewWorker(
	workflowQueue que.Queue,
) *Worker {
	return &Worker{
		workflowQueue: workflowQueue,
	}
}

func (w *Worker) Perform(ctx context.Context, job que.Job) (err error) {
	fmt.Printf("perform job:#%d\n", job.ID())
	return job.Done(ctx)
}

func (w *Worker) Start(ctx context.Context) error {
	ctx, shouldStart, stopped := w.StartInit(ctx)
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

			queWorker, err := que.NewWorker(que.WorkerOptions{
				Mutex:   w.workflowQueue.Mutex(),
				Queue:   "test",
				Perform: w.Perform,
			})
			if err != nil {
				panic(err)
			}

			err = w.runQueWorker(ctx, queWorker)
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

func (w *Worker) runQueWorker(ctx context.Context, queWorker *que.Worker) error {
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
