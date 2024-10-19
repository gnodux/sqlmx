/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package expr

import (
	"github.com/gnodux/sqlmx/dialect"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdate(t *testing.T) {
	type args struct {
		expr *UpdateExpr
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "update",
			args: args{
				expr: Update(N("table")).Set(N("name").Eq(V("name", "gnodux")), N("age").Eq(V("age", 18))),
			},
			want: "UPDATE `table` SET `name` = :name, `age` = :age",
		}, {
			name: "update with where",
			args: args{
				expr: Update(N("table")).Set(N("name").Eq(V("name", "gnodux")), N("age").Eq(V("age", 18))).Where(N("id").Eq(V("id", 1))),
			},
			want: "UPDATE `table` SET `name` = :name, `age` = :age WHERE `id` = :id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewTracedBuffer(dialect.MySQL)
			sql, _, err := b.BuildNamed(tt.args.expr)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, sql)
		})
	}
}
