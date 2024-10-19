/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package expr

import (
	"fmt"
	"github.com/gnodux/sqlmx/expr/keywords"
	"github.com/gnodux/sqlmx/utils"
	"sync/atomic"
	"time"
)

var (
	All          = &RawExpr{Value: "*"}
	NULL         = &ConstantExpr{Value: nil}
	LeftBracket  = &RawExpr{Value: "("}
	RightBracket = &RawExpr{Value: ")"}

	//start shortcut for expr

	//N = Name
	N = Name
	//V = Value
	V = Var
	//R = Raw
	R = Raw
	F = Fn
	C = Const

	autoNum int64 = 0
)

func autoName(p string) string {
	n := atomic.AddInt64(&autoNum, 1)
	return fmt.Sprintf("%s_%d", p, n)
}
func autoParam() string {
	return autoName("param")
}

type RawExpr struct {
	Value any
}

func (r *RawExpr) Format(buffer *TracedBuffer) {
	switch vv := r.Value.(type) {
	case nil:
		buffer.AppendString(buffer.Keyword("NULL"))
	case string:
		buffer.AppendString(vv)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		buffer.AppendString(fmt.Sprintf("%d", vv))
	case float32, float64:
		buffer.AppendString(fmt.Sprintf("%f", vv))
	case time.Time:
		buffer.AppendString(vv.Format(buffer.DateFormat))
	case bool:
		if vv {
			buffer.AppendString(buffer.Keyword("TRUE"))
		} else {
			buffer.AppendString(buffer.Keyword("FALSE"))
		}
	case []byte:
		buffer.AppendString(string(vv))
	default:
		buffer.AppendString(fmt.Sprintf("%v", r.Value))
	}
}

type ConstantExpr struct {
	Value any
}

func (c *ConstantExpr) Format(buffer *TracedBuffer) {
	switch vv := c.Value.(type) {
	case nil:
		buffer.AppendString(buffer.Keyword("NULL"))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		buffer.AppendString(fmt.Sprintf("%d", vv))
	case float32, float64:
		buffer.AppendString(fmt.Sprintf("%f", vv))
	case time.Time:
		buffer.AppendString(vv.Format(buffer.DateFormat))
	case bool:
		if vv {
			buffer.AppendString(buffer.Keyword("TRUE"))
		} else {
			buffer.AppendString(buffer.Keyword("FALSE"))
		}
	case []byte:
		buffer.AppendString("'").AppendString(utils.Escape(string(vv))).AppendString("'")
	case string:
		buffer.AppendString("'").AppendString(utils.Escape(vv)).AppendString("'")
	default:
		buffer.AppendString("'").AppendString(fmt.Sprintf("%v", c.Value)).AppendString("'")
	}
}

type NameExpr struct {
	Qualifier []string
	Name      string
}

func (n *NameExpr) Format(buffer *TracedBuffer) {
	if len(n.Qualifier) > 0 {
		for _, qualifier := range n.Qualifier {
			if len(qualifier) == 0 {
				buffer.AppendString(buffer.SQLNameFunc(qualifier)).AppendString(".")
			}
		}
		buffer.AppendString(buffer.SQLNameFunc(n.Qualifier))
		buffer.AppendString(".")
	}
	buffer.AppendString(buffer.SQLNameFunc(n.Name))
}

func (n *NameExpr) Eq(value any) *BinaryExpr {
	return Eq(n, value)
}
func (n *NameExpr) Ne(value any) *BinaryExpr {
	return Ne(n, value)
}
func (n *NameExpr) Gt(value any) *BinaryExpr {
	return Gt(n, value)
}
func (n *NameExpr) Ge(value any) *BinaryExpr {
	return Ge(n, value)
}
func (n *NameExpr) Lt(value any) *BinaryExpr {
	return Lt(n, value)
}
func (n *NameExpr) Le(value any) *BinaryExpr {
	return Le(n, value)
}
func (n *NameExpr) Like(value any) *BinaryExpr {
	return Like(n, value)
}

func (n *NameExpr) In(values ...any) *BinaryExpr {
	return In(n, n.Name, values...)
}
func (n *NameExpr) NotIn(values ...any) *BinaryExpr {
	return NotIn(n, n.Name, values...)
}
func (n *NameExpr) Between(min, max any) *BetweenExpr {
	minExp, minOk := min.(Expr)
	maxExp, maxOk := max.(Expr)
	if !minOk {
		minExp = Var(autoParam(), min)
	}
	if !maxOk {
		maxExp = Var(autoParam(), max)
	}
	return Between(n, minExp, maxExp)
}

type AroundExpr struct {
	Prefix Expr
	Expr   Expr
	Suffix Expr
}

func (s *AroundExpr) Format(buffer *TracedBuffer) {
	if s.Prefix != nil {
		s.Prefix.Format(buffer)
		buffer.AppendString(" ")
	}

	s.Expr.Format(buffer)
	if s.Suffix != nil {
		buffer.AppendString(" ")
		s.Suffix.Format(buffer)
	}
}

type Tuple struct {
	Name  Expr
	Value Expr
}

// ValueExpr 值表达式
type ValueExpr struct {
	Name  string
	Value any
}

// Format 格式化, 如果是命名参数, 则使用命名参数，否则使用占位符
func (n *ValueExpr) Format(buffer *TracedBuffer) {
	if buffer.NamedVar {
		buffer.AppendNamedArg(n.Name, n.Value)
		buffer.AppendString(buffer.NamedPrefix)
		buffer.AppendString(n.Name)
	} else {
		buffer.AppendArg(n.Value)
		buffer.AppendString(buffer.PlaceHolder)
	}
}

// AliasExpr 别名表达式
type AliasExpr struct {
	Expr  Expr
	Alias string
}

func (a *AliasExpr) Format(buffer *TracedBuffer) {
	a.Expr.Format(buffer)
	buffer.AppendKeywordWithSpace(keywords.AS)
	buffer.AppendString(buffer.SQLNameFunc(a.Alias))
}

// BinaryExpr 二元表达式
type BinaryExpr struct {
	Left     Expr
	Operator string
	Space    string
	Right    Expr
}

// UseSpace 使用空格
func (b *BinaryExpr) UseSpace() *BinaryExpr {
	b.Space = " "
	return b
}

func (b *BinaryExpr) Format(buffer *TracedBuffer) {
	b.Left.Format(buffer)
	buffer.AppendString(b.Space)
	buffer.AppendString(b.Operator)
	buffer.AppendString(b.Space)
	b.Right.Format(buffer)
}

// UnaryExpr 一元表达式
type UnaryExpr struct {
	Operator string
	Expr     Expr
}

func (u *UnaryExpr) Format(buffer *TracedBuffer) {
	buffer.AppendString(u.Operator)
	buffer.AppendString(" ")
	u.Expr.Format(buffer)
}

type ListExpr struct {
	Prefix      Expr
	Separator   string
	Placeholder string
	ExprList    []Expr
	Suffix      Expr
}

func (l *ListExpr) Format(buffer *TracedBuffer) {
	if l.Prefix != nil {
		l.Prefix.Format(buffer)
		buffer.AppendString(" ")
	}
	for idx, expr := range l.ExprList {
		if idx > 0 {
			buffer.AppendString(l.Placeholder).AppendString(l.Separator).AppendString(l.Placeholder)
		}
		expr.Format(buffer)
	}
	if l.Suffix != nil {
		buffer.AppendString(" ")
		l.Suffix.Format(buffer)
	}
}

type FuncExpr struct {
	Name string
	Args []Expr
}

func (f *FuncExpr) Format(buffer *TracedBuffer) {
	buffer.AppendString(buffer.Keyword(f.Name))
	buffer.AppendString("(")
	for idx, arg := range f.Args {
		if idx > 0 {
			buffer.AppendString(",")
		}
		arg.Format(buffer)
	}
	buffer.AppendString(")")
}

type BetweenExpr struct {
	Left  Expr
	Start Expr
	End   Expr
}

func (b *BetweenExpr) Format(buffer *TracedBuffer) {
	b.Left.Format(buffer)
	buffer.AppendString(buffer.KeywordWithSpace(keywords.Between))
	b.Start.Format(buffer)
	buffer.AppendString(buffer.KeywordWithSpace(keywords.And))
	b.End.Format(buffer)
}

// And 与
func And(exprList ...Expr) *ListExpr {
	return &ListExpr{Separator: keywords.And, Placeholder: keywords.Space, ExprList: exprList}
}

// Paren 括号
func Paren(expr Expr) *AroundExpr {
	return &AroundExpr{Prefix: LeftBracket, Expr: expr, Suffix: RightBracket}
}
func Or(exprList ...Expr) *ListExpr {
	return &ListExpr{Separator: keywords.Or, Placeholder: keywords.Space, ExprList: exprList}
}

// Name 字段名、表名、别名等
// 例如：Name("id"):mysql 驱动下则会被格式化为： `id`
// 例如：如果有限定名称Name("id", "user")，则会被格式化为： `user`.`id`，如果有多个限定名称Name("id", "user", "t")，则会被格式化为： `user`.`t`.`id`
func Name(name string, qualifiers ...string) *NameExpr {
	return &NameExpr{Name: name, Qualifier: qualifiers}
}

func Var(name string, value any) *ValueExpr {
	return &ValueExpr{Name: name, Value: value}
}

// Const 常量, 例如：Const(1), Const("a")，和Raw不同的是，Const会自动将值转换为SQL语句中的常量，例如：Const(1)会被格式化为：1，Const("a")会被格式化为：'a'
func Const(value any) *ConstantExpr {
	return &ConstantExpr{Value: value}
}
func List(operator string, exprList ...Expr) *ListExpr {
	return &ListExpr{Separator: operator, ExprList: exprList}
}

func Unary(operator string, expr Expr) *UnaryExpr {
	return &UnaryExpr{Operator: operator, Expr: expr}
}

func Binary(left Expr, op string, right any) *BinaryExpr {
	switch r := right.(type) {
	case Expr:
		return &BinaryExpr{Left: left, Space: " ", Operator: op, Right: r}
	case nil:
		if op == keywords.Equal {
			return &BinaryExpr{Left: left, Space: " ", Operator: "IS", Right: NULL}
		} else if op == keywords.NotEqual {
			return &BinaryExpr{Left: left, Space: " ", Operator: "IS NOT", Right: NULL}
		} else {
			return &BinaryExpr{Left: left, Space: " ", Operator: op, Right: NULL}
		}
	default:
		return &BinaryExpr{Left: left, Space: " ", Operator: op, Right: Const(right)}
	}
}
func Eq(left Expr, right any) *BinaryExpr {
	return Binary(left, keywords.Equal, right)
}
func Ne(left Expr, right any) *BinaryExpr {
	return Binary(left, keywords.NotEqual, right)
}
func Gt(left Expr, right any) *BinaryExpr {
	return Binary(left, keywords.Greater, right)
}
func Ge(left Expr, right any) *BinaryExpr {
	return Binary(left, keywords.GreaterEqual, right)
}
func Lt(left Expr, right any) *BinaryExpr {
	return Binary(left, keywords.Less, right)
}
func Le(left Expr, right any) *BinaryExpr {
	return Binary(left, keywords.LessEqual, right)
}
func Like(left Expr, right any) *BinaryExpr {
	return Binary(left, keywords.Like, right)
}

func Raw(value any) *RawExpr {
	return &RawExpr{Value: value}
}
func Fn(name string, args ...Expr) *FuncExpr {
	return &FuncExpr{Name: name, Args: args}
}

func Between(left Expr, start Expr, end Expr) *BetweenExpr {
	return &BetweenExpr{Left: left, Start: start, End: end}
}

func Alias(expr Expr, alias string) *AliasExpr {
	return &AliasExpr{Expr: expr, Alias: alias}
}

func InValues(left Expr, values ...Expr) *BinaryExpr {
	return &BinaryExpr{Left: left, Space: " ", Operator: keywords.In, Right: Paren(List(",", values...))}
}
func NotInValues(left Expr, values ...Expr) *BinaryExpr {
	return &BinaryExpr{Left: left, Space: " ", Operator: keywords.NotIn, Right: Paren(List(",", values...))}
}

func AutoNamedValues(name string, values ...any) []Expr {
	var exprs []Expr
	const nameFmt = "%s_%d"
	for idx, value := range values {
		switch vv := value.(type) {
		case nil:
			exprs = append(exprs, NULL)
		case Expr:
			exprs = append(exprs, vv)
		default:
			exprs = append(exprs, Var(fmt.Sprintf(nameFmt, name, idx), value))
		}
	}
	return exprs
}

func In(left Expr, name string, values ...any) *BinaryExpr {
	exprs := AutoNamedValues(name, values...)
	return InValues(left, exprs...)
}
func NotIn(left Expr, name string, values ...any) *BinaryExpr {
	exprs := AutoNamedValues(name, values...)
	return NotInValues(left, exprs...)
}

// Not 非, 例如：Not(In(Name("id"), 1,2))
func Not(expr Expr) *UnaryExpr {
	return &UnaryExpr{Operator: keywords.Not, Expr: expr}
}

func Sort(exp Expr, direction string) Expr {
	return List(keywords.Space, exp, Raw(direction))
}
func Sorts(direction string, exps ...Expr) Expr {
	var exprs []Expr
	for _, exp := range exps {
		exprs = append(exprs, Sort(exp, direction))
	}
	return List(",", exprs...)
}
func Desc(exp Expr) Expr {
	return Sort(exp, keywords.Desc)
}
func Asc(exp Expr) Expr {
	return Sort(exp, keywords.Asc)
}
