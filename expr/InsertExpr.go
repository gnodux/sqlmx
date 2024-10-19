/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package expr

import (
	"github.com/gnodux/sqlmx/expr/keywords"
)

// InsertExpr is a struct for insert expression
type InsertExpr struct {
	Table      Expr
	ValueExprs []*BinaryExpr
}

// Into is a function to set table
func (i *InsertExpr) Into(table Expr) *InsertExpr {
	i.Table = table
	return i
}

func (i *InsertExpr) Values(values ...*BinaryExpr) *InsertExpr {
	i.ValueExprs = append(i.ValueExprs, values...)
	return i
}
func (i *InsertExpr) Set(colName string, value any) *InsertExpr {
	i.Values(N(colName).Eq(value))
	return i
}
func (i *InsertExpr) SetExpr(name Expr, value Expr) *InsertExpr {
	i.Values(Binary(name, keywords.Equal, value))
	return i
}
func (i *InsertExpr) SetMap(data map[string]any) *InsertExpr {
	for k, v := range data {
		i.Set(k, v)
	}
	return i
}

func (i *InsertExpr) Format(buf *TracedBuffer) {
	var cols []Expr
	var values []Expr
	for _, exp := range i.ValueExprs {
		cols = append(cols, exp.Left)
		values = append(values, exp.Right)
	}
	buf.AppendKeyword(keywords.InsertInto)
	buf.AppendString(" ")
	i.Table.Format(buf)
	buf.AppendString(" ")
	Paren(List(keywords.Comma, cols...)).Format(buf)
	buf.AppendKeywordWithSpace(keywords.Values)
	Paren(List(keywords.Comma, values...)).Format(buf)
}

// InsertInto 创建一个InsertExpr并设置表名
func InsertInto(table Expr, values ...*BinaryExpr) *InsertExpr {
	return &InsertExpr{Table: table, ValueExprs: values}
}
