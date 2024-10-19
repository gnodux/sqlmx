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

func TestBetween(t *testing.T) {
	type args struct {
		left  Expr
		start Expr
		end   Expr
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "simple",
			args: args{
				left:  Name("id"),
				start: Raw(1),
				end:   Raw(10),
			},
			want: "`id` BETWEEN 1 AND 10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := NewTracedBuffer(dialect.MySQL)
			b := Between(tt.args.left, tt.args.start, tt.args.end)
			b.Format(buf)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func TestIn(t *testing.T) {
	type args struct {
		left   Expr
		name   string
		values []any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "named in",
			args: args{
				left:   Name("id"),
				name:   "id",
				values: []any{1, 2, 3},
			},
			want: "`id` IN ( :id_0,:id_1,:id_2 )",
		},
		{
			name: "unnamed in",
			args: args{
				left:   Name("id"),
				name:   "",
				values: []any{1, 2, 3},
			},
			want: "`id` IN ( ?,?,? )",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := NewTracedBuffer(dialect.MySQL)
			if tt.args.name == "" {
				buf.NamedVar = false
			}
			expr := In(tt.args.left, tt.args.name, tt.args.values...)
			expr.Format(buf)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func TestInsertInto(t *testing.T) {
	tests := []struct {
		name  string
		table Expr
		cols  []*BinaryExpr
		want  string
	}{
		{
			name:  "simple",
			table: Name("user"),
			cols: []*BinaryExpr{
				Name("name").Eq(Const("gnodux")),
				Name("age").Eq(Const(18)),
			},
			want: "INSERT INTO `user` ( `name`,`age` ) VALUES ( 'gnodux',18 )",
		}, {
			name:  "set	map",
			table: Name("user"),
			cols: []*BinaryExpr{
				Name("name").Eq(Var("name", "gnodux")),
				Name("age").Eq(Var("age", 18)),
			},
			want: "INSERT INTO `user` ( `name`,`age` ) VALUES ( :name,:age )",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := NewTracedBuffer(dialect.MySQL)
			InsertInto(tt.table, tt.cols...).Format(buf)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}
