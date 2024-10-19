/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package expr

import "github.com/gnodux/sqlmx/expr/keywords"

type DeleteExpr struct {
	Table     Expr
	WhereExpr Expr
}

func (d *DeleteExpr) Delete(table Expr) *DeleteExpr {
	d.Table = table
	return d
}
func (d *DeleteExpr) Where(exp Expr) *DeleteExpr {
	d.WhereExpr = exp
	return d
}
func (d *DeleteExpr) Format(buf *TracedBuffer) {
	buf.AppendKeyword(keywords.Delete).AppendString(keywords.Space).AppendKeyword(keywords.From).AppendString(keywords.Space)
	d.Table.Format(buf)
	if d.WhereExpr != nil {
		buf.AppendKeywordWithSpace(keywords.Where)
		d.WhereExpr.Format(buf)
	}
}
func Delete(table Expr) *DeleteExpr {
	return &DeleteExpr{Table: table}
}

type DeleteExprFn func(*DeleteExpr)

func UseDeleteCondition(exp Expr) DeleteExprFn {
	return func(s *DeleteExpr) {
		s.WhereExpr = exp
	}
}
