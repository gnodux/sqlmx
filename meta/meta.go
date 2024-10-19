/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package meta

import (
	"github.com/gnodux/sqlmx/expr"
	"github.com/gnodux/sqlmx/utils"
	"reflect"
	"strings"
)

const (
	TagField      = "dbx"
	MarkPK        = "primaryKey"
	MarkIgnore    = "_"
	MarkTenantKey = "tenantKey"
	MarkIsDeleted = "softDelete"
)

var ()

type TableName interface {
	TableName() string
}

type Entity struct {
	Columns        []*Column
	TableName      string
	Name           string
	Type           reflect.Type
	PrimaryKey     *Column
	TenantKey      *Column
	LogicDeleteKey *Column
}

func (m *Entity) String() string {
	return m.TableName
}

func (m *Entity) Format(buffer *expr.TracedBuffer) {
	buffer.AppendString(buffer.SQLNameFunc(m.TableName))
}
func (m *Entity) ColumnExprs() []expr.Expr {
	var exprs []expr.Expr
	for _, col := range m.Columns {
		exprs = append(exprs, col)
	}
	return exprs
}

// ColumnName return column name by field name
func (m *Entity) ColumnName(name string) string {
	for _, col := range m.Columns {
		if col.Name == name {
			return col.ColumnName
		}
	}
	return name
}

func (m *Entity) Column(name string) *Column {
	for _, col := range m.Columns {
		if col.Name == name || col.ColumnName == name {
			return col
		}
	}
	return nil
}

type Column struct {
	Name             string
	ColumnName       string
	Type             reflect.Type
	IsPrimaryKey     bool
	IsTenantKey      bool
	IsLogicDeleteKey bool
	Ignore           bool
}

func (c *Column) String() string {
	return c.ColumnName
}

func (c *Column) Format(buffer *expr.TracedBuffer) {
	buffer.AppendString(buffer.SQLNameFunc(c.ColumnName))
}

func NewEntity(v any) *Entity {
	meta := &Entity{
		TableName: GetTableName(v),
		Columns:   ListValueColumns(v),
		Type:      reflect.TypeOf(v),
		Name:      GetTypeName(v),
	}
	utils.Each(meta.Columns, func(idx int, col *Column) bool {
		if col.IsTenantKey {
			meta.TenantKey = col
		}
		if col.IsPrimaryKey {
			meta.PrimaryKey = col
		}
		if col.IsLogicDeleteKey {
			meta.LogicDeleteKey = col
		}
		return true
	})
	return meta
}

func ListValueColumns(v any) []*Column {
	argv := reflect.TypeOf(v)
	return ListColumns(argv)

}
func ListColumns(t reflect.Type) []*Column {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var fields []*Column
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			if field.Anonymous {
				fields = append(fields, ListColumns(field.Type)...)
			} else {
				fields = append(fields, NewColumnDefWith(t.Field(i)))
			}
		}
	}
	return fields
}
func parseTags(col *Column, tags string) {
	tagList := strings.Split(tags, ",")
	for _, tag := range tagList {
		switch tag {
		case MarkPK:
			col.IsPrimaryKey = true
		case MarkIgnore:
			col.Ignore = true
		case MarkTenantKey:
			col.IsTenantKey = true
		case MarkIsDeleted:
			col.IsLogicDeleteKey = true
		}
	}
}
func NewColumnDefWith(f reflect.StructField) *Column {
	col := &Column{}
	parseTags(col, f.Tag.Get(TagField))
	col.Name = f.Name
	col.ColumnName = utils.LowerCase(f.Name)
	if col.ColumnName == "tenant_id" {
		col.IsTenantKey = true
	}
	if col.ColumnName == "id" {
		col.IsPrimaryKey = true
	}
	if col.ColumnName == "is_deleted" {
		col.IsLogicDeleteKey = true
	}
	col.Type = f.Type
	return col
}

func GetTableName(v any) string {
	if ni, ok := v.(TableName); ok {
		return ni.TableName()
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return utils.LowerCase(t.Name())
}
func GetTypeName(v any) string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
