/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlmx

import (
	"fmt"
	"github.com/gnodux/sqlmx/dialect"
	. "github.com/gnodux/sqlmx/meta"
	"github.com/gnodux/sqlmx/utils"
	"reflect"
	"strings"
	"text/template"
	"time"
)

func MakeFuncMap(driver *dialect.Dialect) template.FuncMap {
	return template.FuncMap{
		"where":      func(v any) string { return where(driver, v) },
		"namedWhere": func(v any) string { return namedWhere(driver, v) },
		"nwhere":     func(v any) string { return namedWhere(driver, v) },
		//"asc":        func(cols []string) string { return orderByMap(driver, expr.SimpleAsc(cols...)) },
		//"desc":       func(cols []string) string { return orderByMap(driver, expr.SimpleDesc(cols...)) },
		"v":          func(v any) string { return sqlValue(driver, v) },
		"n":          driver.SQLNameFunc,
		"sqlName":    driver.SQLNameFunc,
		"list":       func(v []any) string { return sqlValues(driver, v) },
		"columns":    func(v []*Column) string { return columns(driver, v) },
		"allColumns": func(v []*Column) string { return allColumns(driver, v) },
		"args":       func(v []*Column) string { return args(driver, v) },
		"setArgs":    func(v []*Column) string { return sets(v, driver) },
		"orderBy":    func(v map[string]string) string { return orderByMap(driver, v) },
		"driver":     func() string { return driver.Name },
		"dialect": func() string {
			return driver.Name
		},
	}
}

func orderByMap(driver *dialect.Dialect, order map[string]string) string {
	if len(order) == 0 {
		return ""
	}
	sb := strings.Builder{}
	pre := driver.KeywordWithSpace("ORDER BY")
	for k, v := range order {
		sb.WriteString(pre)
		sb.WriteString(driver.SQLNameFunc(driver.NameFunc(k)))
		sb.WriteString(" ")
		sb.WriteString(v)
		pre = ","
	}
	if sb.Len() > 0 {
		sb.WriteString(" ")
	}
	return sb.String()
}
func namedWhere(driver *dialect.Dialect, v any) string {
	return whereWith(driver, v, driver.KeywordWithSpace("AND"), true)
}

func columns(driver *dialect.Dialect, cols []*Column) string {
	sb := strings.Builder{}
	pre := ""
	for _, c := range cols {
		if c.Ignore || c.IsPrimaryKey {
			continue
		}
		sb.WriteString(pre)
		sb.WriteString(driver.SQLNameFunc(c.ColumnName))
		pre = ","
	}
	return sb.String()
}
func allColumns(driver *dialect.Dialect, cols []*Column) string {
	sb := strings.Builder{}
	pre := ""
	for _, c := range cols {
		sb.WriteString(pre)
		sb.WriteString(driver.SQLNameFunc(c.ColumnName))
		pre = ","
	}
	return sb.String()
}

func args(driver *dialect.Dialect, cols []*Column) string {
	sb := strings.Builder{}
	pre := ""
	for _, c := range cols {
		if c.Ignore || c.IsPrimaryKey {
			continue
		}
		sb.WriteString(pre)
		sb.WriteString(driver.NamedPrefix)
		sb.WriteString(c.ColumnName)
		pre = ","
	}
	return sb.String()
}

func sets(cols []*Column, driver *dialect.Dialect) string {
	sb := &strings.Builder{}
	pre := ""
	for _, c := range cols {
		if c.Ignore || c.IsPrimaryKey {
			continue
		}
		sb.WriteString(pre)
		sb.WriteString(driver.SQLNameFunc(c.ColumnName))
		sb.WriteString("=" + driver.NamedPrefix)
		sb.WriteString(c.ColumnName)
		pre = ","
	}
	return sb.String()
}

func where(driver *dialect.Dialect, v any) string {
	return whereWith(driver, v, driver.KeywordWithSpace("AND"), false)
}
func whereOr(driver *dialect.Dialect, v any) string {
	return whereWith(driver, v, driver.KeywordWithSpace("OR"), false)
}

func whereWith(driver *dialect.Dialect, arg any, op string, named bool) string {
	argv := reflect.ValueOf(arg)
	if arg == nil {
		return ""
	}
	if op == "" {
		op = driver.KeywordWithSpace("AND")
	}
	if op[0] != ' ' {
		op = " " + op + " "
	}

	buf := strings.Builder{}
	switch reflect.TypeOf(argv.Interface()).Kind() {
	case reflect.Map:
		comma := driver.KeywordWithSpace("WHERE")
		for _, k := range argv.MapKeys() {
			buf.WriteString(comma)
			buf.WriteString(driver.SQLNameFunc(driver.NameFunc(k.String())))
			value := argv.MapIndex(k)
			switch reflect.TypeOf(value.Interface()).Kind() {
			case reflect.String:
				const stringMatchers = "%.?"
				if strings.ContainsAny(value.Interface().(string), stringMatchers) {
					buf.WriteString(driver.KeywordWithSpace("LIKE"))
				} else {
					buf.WriteString("=")
				}
			default:
				buf.WriteString("=")
			}
			if named {
				buf.WriteString(fmt.Sprintf(":%s", k.String()))
			} else {
				buf.WriteString(sqlValue(driver, value.Interface()))
			}
			comma = op
		}
	case reflect.Struct:
		comma := driver.KeywordWithSpace("WHERE")
		for i := 0; i < argv.NumField(); i++ {
			if argv.Field(i).IsZero() {
				continue
			}
			buf.WriteString(comma)
			buf.WriteString(driver.SQLNameFunc(driver.NameFunc(argv.Type().Field(i).Name)))
			buf.WriteString("=")
			buf.WriteString(sqlValue(driver, argv.Field(i).Interface()))
			comma = op
		}

	}
	buf.WriteString(" ")
	return buf.String()

}

// sqlValues list of sqlValues
func sqlValues(driver *dialect.Dialect, v any) string {
	value := reflect.ValueOf(v)
	comma := ""
	sb := &strings.Builder{}
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		for idx := 0; idx < value.Len(); idx++ {
			sb.WriteString(comma)
			sb.WriteString(sqlValue(driver, value.Index(idx).Interface()))
			comma = ","
		}
	default:
		sb.WriteString(sqlValue(driver, value.Interface()))
	}
	return sb.String()
}

// value sql value converter(sql inject process)
func sqlValue(driver *dialect.Dialect, arg any) string {
	var ret string
	switch a := arg.(type) {
	case nil:
		ret = driver.Keyword("NULL")
	case string:
		ret = "'" + utils.Escape(a) + "'"
	case *string:
		ret = "'" + utils.Escape(*a) + "'"
	case time.Time:
		ret = a.Format(driver.DateFormat)
	case *time.Time:
		ret = a.Format(driver.DateFormat)
	case bool:
		if a {
			ret = driver.Keyword("TRUE")
		} else {
			ret = driver.Keyword("FALSE")
		}
	case uint, uint16, uint32, uint64, int, int16, int32, int64, float32, float64:
		return fmt.Sprintf("%v", a)
	default:
		ret = fmt.Sprintf("'%v'", a)
	}

	return ret
}
