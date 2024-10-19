/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlmx

import "reflect"

type BeforeUpdate interface {
	BeforeUpdate() error
}
type AfterUpdate interface {
	AfterUpdate() error
}
type BeforeInsert interface {
	BeforeInsert() error
}
type AfterInsert interface {
	AfterInsert() error
}
type BeforeDelete interface {
	BeforeDelete() error
}
type AfterDelete interface {
	AfterDelete() error
}

func EvalBeforeHooks[T any](entities ...T) error {
	for idx, _ := range entities {
		if reflect.TypeOf(entities[idx]).Kind() == reflect.Ptr {
			if err := EvalBeforeHook(entities[idx]); err != nil {
				return err
			}
		} else {
			if err := EvalBeforeHook(&entities[idx]); err != nil {
				return err
			}
		}
	}
	return nil
}

// EvalBeforeHook eval before hook
// 如果entity是nil，那么不会调用任何hook
// 如果entity是struct 而非指针，那么在调用EvalBeforeHook之前，需要先将entity转换为指针
func EvalBeforeHook(entity any) error {
	switch e := entity.(type) {
	case nil:
		return nil
	case BeforeUpdate:
		return e.BeforeUpdate()
	case BeforeInsert:
		return e.BeforeInsert()
	case BeforeDelete:
		return e.BeforeDelete()
	}
	return nil
}

func EvalAfterHooks[T any](entities ...T) error {
	for idx, _ := range entities {
		if reflect.TypeOf(entities[idx]).Kind() == reflect.Ptr {
			if err := EvalAfterHook(entities[idx]); err != nil {
				return err
			}
		} else {
			if err := EvalAfterHook(&entities[idx]); err != nil {
				return err
			}
		}
	}
	return nil
}
func EvalAfterHook(entity any) error {
	switch e := entity.(type) {
	case nil:
		return nil
	case AfterUpdate:
		return e.AfterUpdate()
	case AfterInsert:
		return e.AfterInsert()
	case AfterDelete:
		return e.AfterDelete()
	}
	return nil
}
