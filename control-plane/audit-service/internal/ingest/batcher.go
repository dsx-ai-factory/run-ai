/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ingest

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/run-ai/runai/control-plane/audit-service/internal/config"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	"github.com/run-ai/runai/runai-common-packages/logging"
)

type BatchHandler[T any] func(ctx context.Context, b []*T) error

type Batcherer[T any] interface {
	Add(item *T) error
	WaitDone() (exitedGracefully bool)
}

type Batcher[T any] struct {
	batchCh chan *T
	conf    *config.BatcherConfig
	ctx     context.Context
	handler BatchHandler[T]
	wg      *sync.WaitGroup
}

func NewBatcher[T any](ctx context.Context, handler BatchHandler[T]) *Batcher[T] {
	conf := runai_config.Get[*config.BatcherConfig]()
	batcher := &Batcher[T]{
		batchCh: make(chan *T, conf.MaxSize*10),
		conf:    conf,
		handler: handler,
		ctx:     ctx,
		wg:      &sync.WaitGroup{},
	}

	batcher.wg.Add(1)
	go batcher.flushTask()

	return batcher
}

func (b *Batcher[T]) Add(item *T) error {
	select {
	case b.batchCh <- item:
	default:
		return errors.New("failed to batch item in item channel")
	}
	return nil
}

func (b *Batcher[T]) flushTask() {
	defer b.wg.Done()

	var itemsToFlush []*T

	maxInterval := time.Duration(b.conf.IntervalSec) * time.Second
	for {
		select {
		case <-time.After(maxInterval):
			itemsToFlush = b.flush(itemsToFlush)
		case <-b.ctx.Done():
			logging.Infof("batcher context was canceled. Flushing whatever left and shutting down")
			close(b.batchCh)
			b.flush(itemsToFlush)
			return
		case item := <-b.batchCh:
			itemsToFlush = append(itemsToFlush, item)
			if len(itemsToFlush) >= b.conf.MaxSize {
				itemsToFlush = b.flush(itemsToFlush)
			}
		}
	}
}

func (b *Batcher[T]) flush(itemsToFlush []*T) []*T {
	for len(itemsToFlush) > 0 {
		c := b.conf.MaxSize
		if len(itemsToFlush) < c {
			c = len(itemsToFlush)
		}

		batchFlush := itemsToFlush[:c]
		itemsToFlush = itemsToFlush[c:]
		if err := b.handler(b.ctx, batchFlush); err != nil {
			logging.Errorf("failed handling a batch of %d items. All items are lost! %v", len(itemsToFlush), err)
		}
	}
	return nil
}

func (b *Batcher[T]) WaitDone() (exitedGracefully bool) {
	c := make(chan struct{})
	go func() {
		defer close(c)
		b.wg.Wait()
	}()

	maxInterval := time.Duration(b.conf.IntervalSec) * time.Second
	select {
	case <-c:
		logging.Infof("batcher exited gracefully (managed to handle all batched items)")
		return true
	case <-time.After(maxInterval * 2):
		logging.Warnf("batcher exited by timeout. Some batched items may have lost")
		return false
	}
}
