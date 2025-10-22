package processor

import (
	"sync/atomic"
	"time"
)

type Processor interface {
	Start() error
	Stop() error
	Restart() error
	RunningTask()
}

type BaseProcessor struct {
	processor Processor
	running   atomic.Bool // true if running
}

func (b *BaseProcessor) Init(p Processor) {
	b.processor = p
}

func (b *BaseProcessor) Start() error {
	b.running.Store(true)

	go func() {
		for b.running.Load() {
			b.processor.RunningTask()
		}
	}()
	return nil
}

func (b *BaseProcessor) Stop() error {
	b.running.Store(false)
	return nil
}

func (b *BaseProcessor) Restart() error {
	if err := b.Stop(); err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	return b.Start()
}

func (b *BaseProcessor) IsRunning() bool {
	return b.running.Load()
}
