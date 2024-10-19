/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

// Package expr package 用于生成sql语句
package expr

//go:generate go run genindx.go

import (
	"fmt"
	"github.com/gnodux/sqlmx/dialect"
	"reflect"
	"strings"
)

type valueHandler func(interface{}, *TracedBuffer)

func handleString(value interface{}, buffer *TracedBuffer) {
	buffer.AppendString(value.(string))
}

func handleInt(value interface{}, buffer *TracedBuffer) {
	buffer.AppendString(fmt.Sprintf("%d", value))
}

func handleFloat(value interface{}, buffer *TracedBuffer) {
	buffer.AppendString(fmt.Sprintf("%f", value))
}

func handleBool(value interface{}, buffer *TracedBuffer) {
	if value.(bool) {
		buffer.AppendString(buffer.Keyword("TRUE"))
	} else {
		buffer.AppendString(buffer.Keyword("FALSE"))
	}
}

func handleSliceByte(value interface{}, buffer *TracedBuffer) {
	buffer.AppendString(string(value.([]byte)))
}

// 时间格式化字符串定义为常量，避免外部可控制
const dateFormat = "2006-01-02"

// valueHandler定义了处理函数的签名

// handleMap存储了每种类型的处理函数
var handleMap = map[reflect.Kind]valueHandler{
	reflect.String:  handleString,
	reflect.Int:     handleInt,
	reflect.Float64: handleFloat,
	reflect.Bool:    handleBool,
	reflect.Slice:   handleSliceByte,
	// 可以继续添加更多类型的处理函数
}

// DefaultHandler 处理未知类型的逻辑
func DefaultHandler(value interface{}, buffer *TracedBuffer) bool {
	// 使用fmt.Sprintf处理未知类型，避免panic，但性能可能不佳
	buffer.AppendString(fmt.Sprintf("%v", value))
	return false
}

// TracedBuffer is a buffer that can be used to trace the
type TracedBuffer struct {
	//namedArgs 命名参数
	namedArgs map[string]any
	//args 位置参数
	args     []any
	NamedVar bool
	*dialect.Dialect
	strings.Builder
}

func NewTracedBuffer(driver *dialect.Dialect) *TracedBuffer {
	return &TracedBuffer{Dialect: driver, NamedVar: true}
}

func (t *TracedBuffer) AppendNamedArg(name string, value any) *TracedBuffer {
	if t.namedArgs == nil {
		t.namedArgs = map[string]any{}
	}
	t.namedArgs[name] = value
	return t
}
func (t *TracedBuffer) AppendArg(value any) *TracedBuffer {
	t.args = append(t.args, value)
	return t
}

// Append 将value追加到buffer中
func (t *TracedBuffer) Append(value any) *TracedBuffer {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			t.AppendString(t.Keyword("NULL"))
		} else {
			t.Append(v.Elem().Interface())
		}
	} else {
		if handler, exists := handleMap[v.Kind()]; exists {
			handler(value, t)
		} else {
			DefaultHandler(value, t)
		}
	}
	return t
}

func (t *TracedBuffer) NewLine() *TracedBuffer {
	t.Builder.WriteString("\n")
	return t
}
func (t *TracedBuffer) AppendString(s string) *TracedBuffer {
	t.Builder.WriteString(s)
	return t
}
func (t *TracedBuffer) AppendExprs(exprs ...Expr) *TracedBuffer {
	for _, expr := range exprs {
		expr.Format(t)
	}
	return t
}

func (t *TracedBuffer) AppendKeyword(keyword string) *TracedBuffer {
	t.Builder.WriteString(t.Keyword(keyword))
	return t
}
func (t *TracedBuffer) AppendKeywordWithSpace(keyword string) *TracedBuffer {
	t.Builder.WriteString(t.KeywordWithSpace(keyword))
	return t
}
func (t *TracedBuffer) Build(exp Expr) (string, []any, error) {
	t.NamedVar = false
	t.Builder.Reset()
	exp.Format(t)
	return t.Builder.String(), t.args, nil
}
func (t *TracedBuffer) BuildNamed(exp Expr) (string, map[string]any, error) {
	t.NamedVar = true
	t.Builder.Reset()
	exp.Format(t)
	return t.Builder.String(), t.namedArgs, nil
}

type Expr interface {
	Format(buffer *TracedBuffer)
}
