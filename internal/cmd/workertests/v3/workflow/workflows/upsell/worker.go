package upsell

import (
	"context"
	"fmt"

	"github.com/tnclong/go-que"
	"github.com/w11th/goplayground/internal/cmd/workertests/v3/workflow"
)

type Worker struct {
	workflow.BaseWorker

	workflowQueue que.Queue
}

func NewWorker(
	workflowQueue que.Queue,
) *Worker {
	w := &Worker{}
	w.BaseWorker.NewQueWorkerFunc = w.newQueueWorker

	return w
}

func (w *Worker) newQueueWorker() (*que.Worker, error) {
	return que.NewWorker(que.WorkerOptions{
		Mutex:   w.workflowQueue.Mutex(),
		Queue:   "test",
		Perform: w.quePerform,
	})
}

func (w *Worker) quePerform(ctx context.Context, job que.Job) (err error) {
	fmt.Printf("perform job:#%d\n", job.ID())
	return job.Done(ctx)
}
