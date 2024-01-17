package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var ErrStopped error = errors.New("err stopped")

type InnerWorker struct {
	ongoing int64
	stopped int64
}

func (w *InnerWorker) Run() error {
	if w.IsStopped() {
		return ErrStopped
	}

	defer func() {
		if !w.IsStopped() {
			w.Stop(context.Background())
		}
	}()

	for {
		if w.IsStopped() {
			return ErrStopped
		}

		atomic.AddInt64(&w.ongoing, 1)
		time.Sleep(2 * time.Second)
		fmt.Println(time.Now())
		atomic.AddInt64(&w.ongoing, -1)
	}
}

func (w *InnerWorker) Stop(ctx context.Context) error {
	atomic.StoreInt64(&w.stopped, 1)

	fmt.Println("inner worker waiting")
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
wait:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if atomic.LoadInt64(&w.ongoing) == 0 {
				break wait
			}
		}
	}

	fmt.Println("inner worker stopped")
	return nil
}

func (w *InnerWorker) IsStopped() bool {
	return atomic.LoadInt64(&w.stopped) == 1
}

// -> wrapped worker

type Worker struct {
	mu      sync.Mutex
	started bool
	stopped chan struct{}

	cancelFunc context.CancelFunc
}

func NewWorker() *Worker {
	return &Worker{}
}

func (w *Worker) Start(ctx context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.started {
		return
	}

	w.started = true
	w.stopped = make(chan struct{})
	ctx, w.cancelFunc = context.WithCancel(ctx)

	go func() {
		defer close(w.stopped)

		for {
			select {
			case <-ctx.Done():
				break
			default:
			}

			err := w.startInnerWorker(ctx)
			if errors.Is(err, context.Canceled) {
				break
			}

			fmt.Printf("inner worker err: %s\n", err)
			fmt.Println("trying to restart inner worker")
			time.Sleep(2 * time.Second)
		}
	}()
}

func (w *Worker) startInnerWorker(ctx context.Context) (err error) {
	defer func(err error) { fmt.Printf("returned err: %s\n", err) }(err)
	c := make(chan error, 1)

	iw := &InnerWorker{}
	go func() { c <- iw.Run() }()

	select {
	case <-ctx.Done():
		iw.Stop(context.Background())
		err := <-c // Wait for f to return.
		fmt.Printf("inner worker err: %s\n", err)
		return ctx.Err()
	case err := <-c:
		return err
	}
}

func (w *Worker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.stopped == nil {
		return
	}

	w.cancelFunc()

	<-w.stopped
	w.started = false
	w.stopped = nil
}

func main() {
	w := NewWorker()
	w.Start(context.Background())
	defer w.Stop()

	sigintOrTerm := make(chan os.Signal, 1)
	signal.Notify(sigintOrTerm, syscall.SIGINT, syscall.SIGTERM)

	<-sigintOrTerm
}
