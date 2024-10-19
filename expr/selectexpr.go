/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package expr

import (
	"github.com/gnodux/sqlmx/expr/keywords"
)

var (
	Count    = &FuncExpr{Name: keywords.Count, Args: []Expr{Raw(1)}}
	CountAll = &FuncExpr{Name: keywords.Count, Args: []Expr{Raw("*")}}
)

type SelectExpr struct {
	Columns     Expr
	FromExpr    Expr
	WhereExpr   Expr
	GroupByExpr Expr
	HavingExpr  Expr
	OrderByExpr Expr
	limit       int
	offset      int
	withCount   bool
}

func (s *SelectExpr) UseCount() bool {
	return s.withCount
}
func (s *SelectExpr) WithCount() *SelectExpr {
	s.withCount = true
	return s
}
func (s *SelectExpr) WithoutCount() *SelectExpr {
	s.withCount = false
	return s
}

func (s *SelectExpr) BuildCountExpr() *SelectExpr {
	return Select(Count).
		From(s.FromExpr).
		Where(s.WhereExpr).GroupBy(s.GroupByExpr).Having(s.HavingExpr)
}
func (s *SelectExpr) Limit(limit int) *SelectExpr {
	s.limit = limit
	return s
}
func (s *SelectExpr) Offset(offset int) *SelectExpr {
	s.offset = offset
	return s
}
func (s *SelectExpr) Select(columns ...Expr) *SelectExpr {
	s.Columns = List(",", columns...)
	return s
}
func (s *SelectExpr) From(from Expr) *SelectExpr {
	s.FromExpr = from
	return s
}
func (s *SelectExpr) Where(exp Expr) *SelectExpr {
	s.WhereExpr = exp
	return s
}
func (s *SelectExpr) Having(exp Expr) *SelectExpr {
	s.HavingExpr = exp
	return s
}
func (s *SelectExpr) GroupBy(exp Expr) *SelectExpr {
	s.GroupByExpr = exp
	return s
}
func (s *SelectExpr) OrderBy(exps ...Expr) *SelectExpr {
	s.OrderByExpr = List(keywords.Comma, exps...)
	return s
}

func (s *SelectExpr) Format(buffer *TracedBuffer) {
	buffer.AppendString(buffer.Keyword(keywords.Select))
	buffer.AppendString(" ")
	if s.Columns == nil {
		All.Format(buffer)
	} else {
		s.Columns.Format(buffer)
	}
	buffer.AppendString(buffer.KeywordWithSpace(keywords.From))
	s.FromExpr.Format(buffer)
	if s.WhereExpr != nil {
		buffer.AppendString(buffer.KeywordWithSpace(keywords.Where))
		s.WhereExpr.Format(buffer)
	}
	if s.GroupByExpr != nil {
		buffer.AppendString(buffer.KeywordWithSpace(keywords.GroupBy))
		s.GroupByExpr.Format(buffer)
	}
	if s.HavingExpr != nil {
		buffer.AppendString(buffer.KeywordWithSpace(keywords.Having))
		s.HavingExpr.Format(buffer)
	}
	if s.OrderByExpr != nil {
		buffer.AppendString(buffer.KeywordWithSpace(keywords.OrderBy))
		s.OrderByExpr.Format(buffer)
	}
	if s.limit != 0 {

		buffer.AppendString(buffer.KeywordWithSpace(keywords.Limit))
		buffer.AppendString(keywords.Space)
		Var("limit", s.limit).Format(buffer)
		buffer.AppendString(keywords.Space)
		buffer.AppendString(buffer.KeywordWithSpace(keywords.Offset))
		buffer.AppendString(keywords.Space)
		Var("offset", s.offset).Format(buffer)
		buffer.AppendString(keywords.Space)
	}
}

func Select(columns ...Expr) *SelectExpr {
	return &SelectExpr{
		Columns: List(",", columns...),
	}
}
