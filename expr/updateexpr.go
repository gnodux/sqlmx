/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package expr

import "github.com/gnodux/sqlmx/expr/keywords"

type UpdateExpr struct {
	Table     Expr
	Values    []Expr
	WhereExpr Expr
}

func (u *UpdateExpr) Update(table Expr) *UpdateExpr {
	u.Table = table
	return u
}
func (u *UpdateExpr) Set(values ...Expr) *UpdateExpr {
	u.Values = values
	return u
}
func (u *UpdateExpr) Where(exp Expr) *UpdateExpr {
	u.WhereExpr = exp
	return u
}

func Update(table Expr) *UpdateExpr {
	return &UpdateExpr{Table: table}
}

func (u *UpdateExpr) Format(buf *TracedBuffer) {
	buf.AppendKeyword(keywords.Update).AppendString(keywords.Space)
	u.Table.Format(buf)
	buf.AppendKeywordWithSpace(keywords.Set)
	for i, v := range u.Values {
		if i != 0 {
			buf.WriteString(", ")
		}
		v.Format(buf)
	}
	if u.WhereExpr != nil {
		buf.AppendKeywordWithSpace(keywords.Where)
		u.WhereExpr.Format(buf)
	}
}
