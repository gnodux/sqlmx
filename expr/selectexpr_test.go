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

func TestSelect(t *testing.T) {
	type args struct {
		expr Expr
	}

	tests := []struct {
		name      string
		args      args
		want      string
		wantError bool
	}{
		{
			name: "simple",
			args: args{
				expr: Select(All).From(Name("table")),
			},
			want: "SELECT * FROM `table`",
		}, {
			name: "simple with cols",
			args: args{
				expr: Select(Name("c1"), Name("c2")).From(Name("table")).Where(Eq(Name("id"), Raw(1))),
			},
			want: "SELECT `c1`,`c2` FROM `table` WHERE `id` = 1",
		},
		{
			name: "select with where",
			args: args{expr: Select(All).From(Name("table")).Where(Eq(Name("id"), Raw(1)))},
			want: "SELECT * FROM `table` WHERE `id` = 1",
		}, {
			name: "select with where and group by",
			args: args{expr: Select(All).From(Name("table")).Where(Eq(Name("id"), Raw(1))).GroupBy(Name("id"))},
			want: "SELECT * FROM `table` WHERE `id` = 1 GROUP BY `id`",
		}, {
			name: "order by test",
			args: args{expr: Select(All).From(Name("table")).Where(Eq(Name("id"), Raw(1))).GroupBy(Name("id")).OrderBy(Name("id"))},
			want: "SELECT * FROM `table` WHERE `id` = 1 GROUP BY `id` ORDER BY `id`",
		}, {
			name: "order by test(multi)",
			args: args{expr: Select(All).From(Name("table")).Where(Eq(Name("id"), Raw(1))).GroupBy(Name("id")).OrderBy(Name("id"), Name("name"))},
			want: "SELECT * FROM `table` WHERE `id` = 1 GROUP BY `id` ORDER BY `id`,`name`",
		},
		{
			name: "order by test(multi with sort direction)",
			args: args{expr: Select(All).From(N("table")).Where(Eq(N("id"), R(1))).GroupBy(N("id")).OrderBy(Asc(N("id")), Desc(N("name")))},
			want: "SELECT * FROM `table` WHERE `id` = 1 GROUP BY `id` ORDER BY `id` ASC,`name` DESC",
		},
		{
			name: "order by test(multi but with single sort direction)",
			args: args{expr: Select(All).From(N("table")).Where(Eq(N("id"), R(1))).GroupBy(N("id")).OrderBy(N("id"), Desc(N("name")))},
			want: "SELECT * FROM `table` WHERE `id` = 1 GROUP BY `id` ORDER BY `id`,`name` DESC",
		}, {
			name: "select with multiple where",
			args: args{
				expr: Select(All).From(Name("table")).Where(And(Eq(Name("id"), Const(1)), Eq(Name("name"), Const("test")))),
			},
			want: "SELECT * FROM `table` WHERE `id` = 1 AND `name` = 'test'",
		}, {
			name: "select with multiple where(OR)",
			args: args{
				expr: Select(All).From(Name("table")).Where(And(Eq(Name("user_id"), Const(3)), Or(Eq(Name("id"), Const(1)), Eq(Name("name"), Const("test"))))),
			},
			want: "SELECT * FROM `table` WHERE `user_id` = 3 AND `id` = 1 OR `name` = 'test'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := NewTracedBuffer(dialect.MySQL)
			tt.args.expr.Format(buffer)

			assert.Equal(t, tt.want, buffer.String())
			if buffer.String() != tt.want {
				t.Errorf("want %s, but got %s", tt.want, buffer.String())
			}
		})
	}
}
