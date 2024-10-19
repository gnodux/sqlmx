/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlmx

import (
	"context"
	"database/sql"
	"github.com/cookieY/sqlx"
)

const (
	sqlSuffix = ".sql"
)

// SelectFunc Select 函数类型，使用位置参数
type SelectFunc[T any] func(args ...any) ([]T, error)

// NamedSelectFunc NamedSelect 函数类型, 用于查询多条记录,使用命名参数
type NamedSelectFunc[T any] func(arg any) ([]T, error)

// GetFunc Get 函数类型, 用于查询单条记录
type GetFunc[T any] func(args ...any) (T, error)

// NamedGetFunc NamedGet 函数类型, 用于查询单条记录,使用命名参数
type NamedGetFunc[T any] func(arg any) (T, error)

// ExecFunc Exec 函数类型, 用于执行无返回值的SQL
type ExecFunc func(args ...any) (sql.Result, error)

// NamedExecFunc NamedExec 函数类型, 用于执行无返回值的SQL,使用命名参数
type NamedExecFunc func(arg any) (sql.Result, error)

// TxFunc Tx 函数类型, 用于执行事务
type TxFunc func(func(*Tx) error) error

// NewSelectFuncWith 创建一个 SelectFunc
// m: DBManager 数据库管理器
// db: 数据库名称
// tpl: SQL模版或者inline SQL
func NewSelectFuncWith[T any](m *DBManager, db, tpl string) SelectFunc[T] {
	return func(args ...any) ([]T, error) {
		d, err := m.Get(db)
		if err != nil {
			return nil, err
		}
		var v []T
		err = d.SelectEx(&v, tpl, args...)
		return v, err
	}
}

// NewSelectFunc 创建一个 SelectFunc
// db: 数据库名称（使用默认的数据库管理器 Manager）
// tpl: SQL模版或者inline SQL
func NewSelectFunc[T any](db, tpl string) SelectFunc[T] {
	return NewSelectFuncWith[T](Manager, db, tpl)
}

// NewTxFuncWith 创建一个 TxFunc
// db: 数据库名称
// tpl: SQL模版或者inline SQL
func NewTxFuncWith(db *DB, tpl string, opts *sql.TxOptions) TxFunc {
	return func(fn func(tx *Tx) error) error {
		return db.BatchEx(context.Background(), opts, tpl, fn)
	}
}

// NewNamedSelectFuncWith 创建一个 NamedSelectFunc
// manager: DBManager 数据库管理器
// db: 数据库名称
// tpl: SQL模版或者inline SQL
func NewNamedSelectFuncWith[T any](manager *DBManager, db, tpl string) NamedSelectFunc[T] {
	return func(arg any) ([]T, error) {
		var v []T
		d, err := manager.Get(db)
		if err != nil {
			return nil, err
		}
		err = d.NamedSelectEx(&v, tpl, arg)
		return v, err
	}
}

// NewNamedSelectFunc 创建一个 NamedSelectFunc
// db: 数据库名称
// tpl: SQL模版或者inline SQL
func NewNamedSelectFunc[T any](db, tpl string) NamedSelectFunc[T] {
	return NewNamedSelectFuncWith[T](Manager, db, tpl)
}

//	func GetFn[T any](db, tpl string) GetFunc[T] {
//		return GetFnWith[T](Manager, db, tpl)
//	}

// NewNamedGetFuncWith 创建一个 NamedGetFunc
// m: DBManager 数据库管理器
// db: 数据库名称
// tpl: SQL模版或者inline SQL
func NewNamedGetFuncWith[T any](m *DBManager, db, tpl string) NamedGetFunc[T] {
	return func(arg any) (v T, err error) {
		var d *DB
		if d, err = m.Get(db); err != nil {
			return
		}
		var n *sqlx.NamedStmt
		if n, err = d.PrepareNamedEx(tpl, arg); err != nil {
			return v, err
		}
		defer func(n *sqlx.NamedStmt) {
			err = n.Close()
		}(n)
		err = n.Get(&v, arg)
		return v, err
	}
}

// NewNamedExecFuncWith 创建一个 NamedExecFunc
// db:*DB 数据库管理器
// tpl: SQL模版或者inline SQL
func NewNamedExecFuncWith(db *DB, tpl string) NamedExecFunc {
	return func(arg any) (sql.Result, error) {
		return db.NamedExecEx(tpl, arg)
	}
}

// NewExecFuncWith 创建一个 ExecFunc
// db:*DB 数据库
// tpl: SQL模版或者inline SQL
func NewExecFuncWith(db *DB, tpl string) ExecFunc {
	return func(args ...any) (sql.Result, error) {
		return db.ExecEx(tpl, args...)
	}
}
