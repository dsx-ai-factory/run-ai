package runaipubsub

import (
	"context"
	"errors"
	"fmt"
	"sync"

	pkgerrors "github.com/pkg/errors"
)

type Subscriber interface {
	Resubscribe() error
}

type ResubscribeHandler struct {
	lock        sync.Locker
	subscribers map[string]Subscriber
}

func NewResubscribeHandler() *ResubscribeHandler {
	return &ResubscribeHandler{
		lock:        &sync.RWMutex{},
		subscribers: map[string]Subscriber{},
	}
}

func (r *ResubscribeHandler) Resubscribe() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	var err error
	for topic, i := range r.subscribers {
		if e := i.Resubscribe(); e != nil {
			err = errors.Join(err, fmt.Errorf("failed to resubscribe for topic %s: %w", topic, e))
		}
	}
	return err
}

func (r *ResubscribeHandler) Add(ctx context.Context, topic string, callback func(payloadJson []byte) error) {
	r.lock.Lock()
	r.subscribers[topic] = NewResubscribe(ctx, topic, callback)
	r.lock.Unlock()
}

type Resubscribe struct {
	ctx               context.Context
	topic             string
	subscribeCallback func(payloadJson []byte) error
}

func NewResubscribe(ctx context.Context, topic string, callback func(payloadJson []byte) error) Subscriber {
	return &Resubscribe{
		ctx:               ctx,
		topic:             topic,
		subscribeCallback: callback,
	}
}

func (r *Resubscribe) Resubscribe() error {
	if pubSuberInstance == nil {
		return nil
	}
	err := pubSuberInstance.Subscribe(r.ctx, r.topic, r.subscribeCallback)
	if err != nil {
		return err
	}

	if additionalPubSuberInstance == nil {
		return nil
	}

	if e := additionalPubSuberInstance.Subscribe(r.ctx, r.topic, r.subscribeCallback); e != nil {
		return pkgerrors.Wrap(e, fmt.Sprintf("additional pubsuber error: %+v", e))
	}

	return nil
}
