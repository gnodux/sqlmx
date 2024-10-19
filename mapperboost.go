/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlmx

import (
	"database/sql"
	"errors"
	"github.com/gnodux/sqlmx/utils"
	"path/filepath"
	"reflect"
	"strings"
)

const (

	// TagDS 数据源
	TagDS = "ds"
	// TagSQL sql模版（或inline sql）
	TagSQL = "sql"

	// TagTx 事务级别
	TagTx = "tx"

	// TagReadonly 事务是否只读
	TagReadonly = "readonly"

	// TxDefault 默认事务级别
	TxDefault = "Default"

	// TxReadUncommitted 事务级别：ReadUncommitted
	TxReadUncommitted = "ReadUncommitted"

	// TxReadCommitted 事务级别：ReadCommitted
	TxReadCommitted = "ReadCommitted"

	// TxWriteCommitted 事务级别：WriteCommitted
	TxWriteCommitted = "WriteCommitted"

	// TxRepeatableRead 事务级别：RepeatableRead
	TxRepeatableRead = "RepeatableRead"

	// TxSnapshot 事务级别：Snapshot
	TxSnapshot = "Snapshot"

	// TxSerializable 事务级别：Serializable
	TxSerializable = "Serializable"

	// TxLinearizable 事务级别：Linearizable
	TxLinearizable = "Linearizable"
)

var (
	ManagerType       = reflect.TypeOf((*DBManager)(nil))
	DBType            = reflect.TypeOf((*DB)(nil))
	ExecFuncType      = reflect.TypeOf(ExecFunc(nil))
	NamedExecFuncType = reflect.TypeOf(NamedExecFunc(nil))
	TxFuncType        = reflect.TypeOf(TxFunc(nil))
)

// parseExtTags 解析字段的自定义tag，包含：数据源、sql模版（或inline sql）、事务级别、事务是否只读等
func parseExtTags(field reflect.StructField) (ds, tpl string, level sql.IsolationLevel, readOnly bool) {
	ds = field.Tag.Get(TagDS)
	tpl = field.Tag.Get(TagSQL)
	txtLevel := field.Tag.Get(TagTx)
	switch txtLevel {
	case TxReadCommitted:
		level = sql.LevelReadCommitted
	case TxReadUncommitted:
		level = sql.LevelReadUncommitted
	case TxWriteCommitted:
		level = sql.LevelWriteCommitted
	case TxRepeatableRead:
		level = sql.LevelRepeatableRead
	case TxSnapshot:
		level = sql.LevelSnapshot
	case TxSerializable:
		level = sql.LevelSerializable
	case TxLinearizable:
		level = sql.LevelLinearizable
	default:
		level = sql.LevelDefault
	}
	r := field.Tag.Get(TagReadonly)
	if r != "" && strings.ToLower(r) != "false" {
		readOnly = true
	}
	return
}

// BoostMapper 对mapper的Field进行wrap处理、绑定数据源、绑定sql模版、绑定事务级别、绑定是否只读等
//
// change: 2023-7-12 修改绑定策略，从延迟绑定修改到boost时绑定，动态打开数据库的需求不高，且模版延迟绑定和获取数据库需要使用到锁，对性能有一定影响
func BoostMapper(dest interface{}, factory *DBManager, ds string) error {
	currentDb, err := factory.Get(ds)
	if err != nil {
		return err
	}
	v := reflect.ValueOf(dest)
	if v.Type().Kind() != reflect.Ptr {
		return errors.New("mapper must a pointer ")
	}
	if v.Elem().Type().Kind() != reflect.Struct {
		return errors.New("mapper can not apply to a struct pointer")
	}
	if ii, ok := dest.(interface{ init() }); ok {
		ii.init()
	}
	v = v.Elem()
	for idx := 0; idx < v.Type().NumField(); idx++ {
		field := v.Type().Field(idx)
		fieldDs, sqlTpl, isoLevel, readonly := parseExtTags(field)
		if fieldDs == "" {
			fieldDs = ds
		}
		if field.IsExported() && field.Type.Kind() == reflect.Struct {
			if err := BoostMapper(v.Field(idx).Addr().Interface(), factory, fieldDs); err != nil {
				return err
			}
			continue
		}
		switch field.Type.Kind() {
		case reflect.Ptr:
			fv := v.Field(idx)
			if fv.CanSet() && field.Type.AssignableTo(ManagerType) {
				//value is manager,Set a DBManager
				fv.Set(reflect.ValueOf(factory))
			}
			if fv.CanSet() && field.Type.AssignableTo(DBType) {
				//value is *DB,set datasource
				if db, err := factory.Get(ds); err != nil {
					return err
				} else {
					fv.Set(reflect.ValueOf(db))
				}
			}
		case reflect.Func:

			var tplList []string
			if sqlTpl == "" {
				sqlTpl = NameFunc(field.Name) + sqlSuffix
				pkgPath := NameFunc(v.Type().PkgPath())
				name := NameFunc(v.Type().Name())
				tplList = []string{
					filepath.Join(pkgPath, name, sqlTpl),
					filepath.Join(filepath.Base(pkgPath), name, sqlTpl),
					filepath.Join(name, sqlTpl),
					sqlTpl,
				}
			} else {
				if !strings.HasSuffix(sqlTpl, sqlSuffix) {
					//对inline(在注解中使用的SQL进行解析,模版名称替换成为完整的包名称
					tplName := filepath.Join(NameFunc(v.Type().PkgPath()), NameFunc(v.Type().Name()),
						NameFunc(field.Name)+sqlSuffix)
					if _, err = currentDb.ParseTemplate(tplName, sqlTpl); err != nil {
						return err
					}
					sqlTpl = tplName
				}
				tplList = append(tplList, sqlTpl)
			}
			switch field.Type {
			case ExecFuncType:
				v.Field(idx).Set(reflect.ValueOf(NewExecFuncWith(currentDb, sqlTpl)))
			case NamedExecFuncType:
				v.Field(idx).Set(reflect.ValueOf(NewNamedExecFuncWith(currentDb, sqlTpl)))
			case TxFuncType:
				v.Field(idx).Set(reflect.ValueOf(NewTxFuncWith(currentDb, sqlTpl, &sql.TxOptions{
					Isolation: isoLevel,
					ReadOnly:  readonly,
				})))
			default:
				name := field.Type.Name()
				//begin: 判断是否泛型，并去除泛型参数
				quotaIdx := strings.Index(name, "[")
				if quotaIdx > 0 {
					name = name[:quotaIdx]
				}
				//end
				var fnVal func([]reflect.Value) []reflect.Value
				switch name {
				case "SelectFunc":
					fnVal = func(values []reflect.Value) []reflect.Value {
						ret, err := SelectWith(field.Type.Out(0).Elem(), currentDb, tplList, values[0].Interface().([]any))
						return []reflect.Value{
							utils.ValueOrZero(ret, field.Type.Out(0)),
							utils.ValueOrZero(err, field.Type.Out(1)),
						}
					}
				case "NamedSelectFunc":
					fnVal = func(values []reflect.Value) []reflect.Value {
						ret, err := NamedSelectWith(field.Type.Out(0).Elem(), currentDb, tplList, values[0].Interface())
						return []reflect.Value{
							utils.ValueOrZero(ret, field.Type.Out(0)),
							utils.ValueOrZero(err, field.Type.Out(1)),
						}
					}
				case "GetFunc":
					fnVal = func(values []reflect.Value) []reflect.Value {
						ret, err := GetWith(field.Type.Out(0), currentDb, tplList, values[0].Interface().([]any))
						return []reflect.Value{
							utils.ValueOrZero(ret, field.Type.Out(0)),
							utils.ValueOrZero(err, field.Type.Out(1)),
						}
					}
				case "NamedGetFunc":
					fnVal = func(values []reflect.Value) []reflect.Value {
						ret, err := NamedGetWith(field.Type.Out(0), currentDb, tplList, values[0].Interface())
						return []reflect.Value{
							utils.ValueOrZero(ret, field.Type.Out(0)),
							utils.ValueOrZero(err, field.Type.Out(1)),
						}
					}
				}
				if fnVal != nil {
					v.Field(idx).Set(reflect.MakeFunc(field.Type, fnVal))
				}
			}
		}
	}
	return nil
}

// Boost 对mapper的Field进行wrap处理、绑定数据源、绑定sql模版、绑定事务级别、绑定是否只读等
// dest: mapper对象
// ds: 数据源
func Boost(dest interface{}, ds string) error {
	return BoostMapper(dest, Manager, ds)
}

// NewMapperWith 创建一个mapper
// factory: 数据库管理器
// dataSource: 数据源
func NewMapperWith[T any](factory *DBManager, dataSource string) (*T, error) {
	var d T
	err := BoostMapper(&d, factory, dataSource)
	return &d, err
}

// NewMapper 创建一个mapper
// dataSource: 数据源
func NewMapper[T any](dataSource string) (*T, error) {
	return NewMapperWith[T](Manager, dataSource)
}
