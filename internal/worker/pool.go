package worker

import (
	"context"
	"log/slog"
	"sync"

	"github.com/BilalGunden-Insider/go-backend/internal/models"
)

type Task struct {
	Transaction *models.Transaction
}

type ProcessFunc func(ctx context.Context, tx *models.Transaction) error

type Pool struct {
	workers    int
	queue      chan Task
	processFn  ProcessFunc
	stats      *Stats
	log        *slog.Logger
	wg         sync.WaitGroup
	cancelFunc context.CancelFunc
}

func NewPool(workers, queueSize int, processFn ProcessFunc, log *slog.Logger) *Pool {
	return &Pool{
		workers:   workers,
		queue:     make(chan Task, queueSize),
		processFn: processFn,
		stats:     &Stats{},
		log:       log,
	}
}

func (p *Pool) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	p.cancelFunc = cancel

	for i := range p.workers {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}
	p.log.Info("worker pool started", slog.Int("workers", p.workers))
}

func (p *Pool) Submit(task Task) bool {
	select {
	case p.queue <- task:
		p.stats.Pending.Add(1)
		return true
	default:
		return false
	}
}

func (p *Pool) Stop() {
	p.cancelFunc()
	close(p.queue)
	p.wg.Wait()
	p.log.Info("worker pool stopped")
}

func (p *Pool) GetStats() *Stats {
	return p.stats
}

func (p *Pool) worker(ctx context.Context, id int) {
	defer p.wg.Done()
	p.log.Debug("worker started", slog.Int("worker_id", id))

	for task := range p.queue {
		p.stats.Pending.Add(-1)
		if err := p.processFn(ctx, task.Transaction); err != nil {
			p.stats.Failed.Add(1)
			p.log.Error("worker task failed",
				slog.Int("worker_id", id),
				slog.String("tx_id", task.Transaction.ID.String()),
				slog.String("error", err.Error()))
		} else {
			p.stats.Processed.Add(1)
		}
	}
}

func (p *Pool) BatchSubmit(tasks []Task) int {
	enqueued := 0
	for _, t := range tasks {
		if p.Submit(t) {
			enqueued++
		}
	}
	return enqueued
}
