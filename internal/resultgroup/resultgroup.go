package resultgroup

import (
	"context"
	"fmt"
	"sync"
)

type MutatingResultGroup[T any] struct {
	wg           sync.WaitGroup
	errOnce      sync.Once
	cancel       func(error)
	routines     int
	mutateTarget *T
	err          error
}

func New[T any](ctx context.Context, init T) (*MutatingResultGroup[T], context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	return &MutatingResultGroup[T]{
		cancel:       cancel,
		mutateTarget: &init,
	}, ctx
}

func (rg *MutatingResultGroup[T]) Wait() (T, error) {
	var zero T
	if rg.routines == 0 {
		return zero, nil
	}
	rg.wg.Wait()
	if rg.cancel != nil {
		rg.cancel(rg.err)
	}
	if rg.err != nil {
		return zero, fmt.Errorf("error in goroutine: %w", rg.err)
	}
	return *rg.mutateTarget, nil
}

func (rg *MutatingResultGroup[T]) Go(f func(t *T) error) {
	rg.routines++
	rg.wg.Add(1)
	go func() {
		defer rg.wg.Done()
		err := f(rg.mutateTarget)
		if err != nil {
			rg.errOnce.Do(func() {
				rg.err = err
				if rg.cancel != nil {
					rg.cancel(rg.err)
				}
			})
			return
		}
	}()
}
