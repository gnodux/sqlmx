/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlmx

import "testing"

type HookedUser struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

func (u *HookedUser) BeforeInsert() error {
	u.ID = 0
	u.Name = "gnodux"
	return nil
}

func TestEvalAfterHooks(t *testing.T) {
	users := []*HookedUser{
		{
			ID:   1,
			Name: "gnodux1",
		}, {
			ID:   2,
			Name: "gnodux2",
		},
	}
	if err := EvalBeforeHooks(users...); err != nil {
		t.Error(err)
	}
	encoder.Encode(users)

	userss := []HookedUser{
		{
			ID:   1,
			Name: "gnodux1",
		}, {
			ID:   2,
			Name: "gnodux2",
		},
	}
	if err := EvalBeforeHooks(userss...); err != nil {
		t.Error(err)
	}
	encoder.Encode(userss)
}
