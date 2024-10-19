/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package expr

import (
	"github.com/gnodux/sqlmx/expr/keywords"
	"strings"
)

type FilterFn func(Expr)
type SelectFilterFn func(s *SelectExpr)
type DeleteFilterFn func(s *DeleteExpr)
type InsertFilterFn func(s *InsertExpr)
type UpdateFilterFn func(s *UpdateExpr)

type Filters []FilterFn

func (f *Filters) Apply(exp Expr) {
	for _, fn := range *f {
		fn(exp)
	}
}
func (f *Filters) ApplyAll(exps []Expr) {
	for _, exp := range exps {
		f.Apply(exp)
	}
}
func (f *Filters) Append(filters ...FilterFn) *Filters {
	*f = append(*f, filters...)
	return f
}
func (f *Filters) Select(filters ...FilterFn) *Filters {
	f.Append(SelectFilter(func(s *SelectExpr) {
		filters[0](s)
	}))
	return f
}
func (f *Filters) Delete(filters ...FilterFn) *Filters {
	f.Append(DeleteFilter(func(s *DeleteExpr) {
		filters[0](s)
	}))
	return f
}
func (f *Filters) Update(filters ...FilterFn) {
	f.Append(UpdateFilter(func(s *UpdateExpr) {
		filters[0](s)
	}))
}

func SelectFilter(fn SelectFilterFn) FilterFn {
	return func(exp Expr) {
		if s, ok := exp.(*SelectExpr); ok {
			fn(s)
		}
	}
}
func DeleteFilter(fn DeleteFilterFn) FilterFn {
	return func(exp Expr) {
		if s, ok := exp.(*DeleteExpr); ok {
			fn(s)
		}
	}
}
func UpdateFilter(fn UpdateFilterFn) FilterFn {
	return func(exp Expr) {
		if s, ok := exp.(*UpdateExpr); ok {
			fn(s)
		}
	}
}

func InsertFilter(fn InsertFilterFn) FilterFn {
	return func(exp Expr) {
		if s, ok := exp.(*InsertExpr); ok {
			fn(s)
		}
	}
}

func UseLimit(limit int) FilterFn {
	return SelectFilter(func(s *SelectExpr) {
		s.limit = limit
	})
}
func UseLimits(limit int, offset int) FilterFn {
	return SelectFilter(func(s *SelectExpr) {
		s.limit = limit
		s.offset = offset
	})
}
func UseOffset(offset int) FilterFn {
	return SelectFilter(func(s *SelectExpr) {
		s.offset = offset
	})
}

func WithCount(exp Expr) {
	if s, ok := exp.(*SelectExpr); ok {
		s.withCount = true
	}
}

// AllToOr 将所有的条件转换为or连接
func AllToOr(exp Expr) {
	var where Expr
	switch ex := exp.(type) {
	case *SelectExpr:
		where = ex.WhereExpr
	case *DeleteExpr:
		where = ex.WhereExpr
	case *UpdateExpr:
		where = ex.WhereExpr
	}
	if where != nil {
		if lst, ok := where.(*ListExpr); ok {
			lst.Placeholder = keywords.Space
			lst.Separator = keywords.Or
		}
	}
}

// AllToAnd 将所有的条件转换为and连接
func AllToAnd(exp Expr) {
	var where Expr
	switch ex := exp.(type) {
	case *SelectExpr:
		where = ex.WhereExpr
	case *DeleteExpr:
		where = ex.WhereExpr
	case *UpdateExpr:
		where = ex.WhereExpr
	}
	if where != nil {
		if lst, ok := where.(*ListExpr); ok {
			lst.Placeholder = keywords.Space
			lst.Separator = keywords.And
		}
	}
}

// UseCondition 使用条件
func UseCondition(exp Expr) FilterFn {
	return func(s Expr) {
		switch ss := s.(type) {
		case *SelectExpr:
			ss.WhereExpr = exp
		case *DeleteExpr:
			ss.WhereExpr = exp
		case *UpdateExpr:
			ss.WhereExpr = exp
		}
	}
}

// AutoFuzzy 自动模糊查询，对于字符类型且包含%.?*等字符的，自动转换为like查询
func AutoFuzzy(exp Expr) {
	switch s := exp.(type) {
	case *SelectExpr:
		fuzzy(s.WhereExpr)
	case *DeleteExpr:
		fuzzy(s.WhereExpr)
	case *UpdateExpr:
		fuzzy(s.WhereExpr)
	}
}

func fuzzy(exp Expr) {
	if exp == nil {
		return
	}
	switch n := exp.(type) {
	case *ListExpr:
		for _, v := range n.ExprList {
			fuzzy(v)
		}
	case *BinaryExpr:
		switch right := n.Right.(type) {
		case *ValueExpr:
			if isStringAndFuzzy(right.Value) {
				n.Space = keywords.Space
				n.Operator = keywords.Like
			}
		case *ConstantExpr:
			if isStringAndFuzzy(right.Value) {
				n.Space = keywords.Space
				n.Operator = keywords.Like
			}
		}
	case *AroundExpr:
		if n.Prefix == LeftBracket && n.Suffix == RightBracket {
			fuzzy(n.Expr)
		}
	}
}
func isStringAndFuzzy(v any) bool {
	switch vv := v.(type) {
	case string:
		if strings.IndexAny(vv, "%?.") >= 0 {
			return true
		}
	}
	return false
}

func Set(exps ...Expr) FilterFn {
	return UpdateFilter(func(s *UpdateExpr) {
		s.Values = exps
	})
}

func UseSort(direct string, exprs ...Expr) FilterFn {
	return SelectFilter(func(s *SelectExpr) {
		s.OrderByExpr = Sorts(direct, exprs...)
	})
}
func UseOrderBy(exp Expr) FilterFn {
	return SelectFilter(func(s *SelectExpr) {
		s.OrderByExpr = exp
	})
}
