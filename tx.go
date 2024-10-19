/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlmx

import (
	"database/sql"
	"github.com/cookieY/sqlx"
	"github.com/gnodux/sqlmx/expr"
)

// Tx transaction wrapper
type Tx struct {
	*sqlx.Tx
	db  *DB
	tpl string
}

func (t *Tx) Tpl() string {
	return t.tpl
}
func (t *Tx) Parse(tplName string, args any) (string, error) {
	if t.db == nil {
		return "", ErrNilDB
	}
	return t.db.ParseSQL(tplName, args)
}

// SelectExpr 使用表达式进行查询
func (t *Tx) SelectExpr(dest interface{}, exp expr.Expr) error {
	if t == nil {
		return ErrNilDB
	}
	buff := expr.NewTracedBuffer(t.db.driver)
	if t.db.driver.SupportNamed {
		query, namedArgs, err := buff.BuildNamed(exp)
		if err != nil {
			return err
		}
		return t.NamedSelect(dest, query, namedArgs)
	} else {
		query, args, err := buff.Build(exp)
		if err != nil {
			return err
		}
		return t.Select(dest, query, args...)
	}
}

func (t *Tx) NamedSelect(dest interface{}, sql string, arg any) (err error) {
	if t == nil {
		return ErrNilDB
	}
	var named *sqlx.NamedStmt
	named, err = t.PrepareNamed(sql)
	if err != nil {
		return err
	}
	defer func(named *sqlx.NamedStmt) {
		err = named.Close()

	}(named)
	log.Debug("named select:", named.QueryString, arg)
	return named.Select(dest, arg)
}

// ExecExpr 使用表达式进行执行
func (t *Tx) ExecExpr(exp expr.Expr) (sql.Result, error) {
	if t == nil {
		return nil, ErrNilDB
	}
	buff := expr.NewTracedBuffer(t.db.driver)
	if t.db.driver.SupportNamed {
		query, namedArgs, err := buff.BuildNamed(exp)
		if err != nil {
			return nil, err
		}
		return t.NamedExec(query, namedArgs)
	} else {
		query, args, err := buff.Build(exp)
		if err != nil {
			return nil, err
		}
		return t.Exec(query, args...)
	}
}

func (t *Tx) GetExpr(dest interface{}, exp expr.Expr) error {
	if t == nil {
		return ErrNilDB
	}
	buff := expr.NewTracedBuffer(t.db.driver)
	if t.db.driver.SupportNamed {
		query, namedArgs, err := buff.BuildNamed(exp)
		if err != nil {
			return err
		}
		return t.NamedGet(dest, query, namedArgs)
	} else {
		query, args, err := buff.Build(exp)
		if err != nil {
			return err
		}
		return t.Get(dest, query, args...)
	}
}
func (t *Tx) NamedGet(dest interface{}, query string, arg interface{}) error {
	stmt, err := t.PrepareNamed(query)
	if err != nil {
		return err
	}
	defer func() {
		_ = stmt.Close()
	}()
	return stmt.Get(dest, arg)
}

// ParseAndPrepareNamed use tplName to parse and prepare named statement
func (t *Tx) ParseAndPrepareNamed(tplName string, arg any) (*sqlx.NamedStmt, error) {
	query, err := t.Parse(tplName, arg)
	if err != nil {
		return nil, err
	}
	return t.PrepareNamed(query)
}

// RunPrepareNamedEx use tplName to prepare named statement
func (t *Tx) RunPrepareNamedEx(sqlOrTpl string, arg any, fn func(*sqlx.NamedStmt) error) (err error) {
	var stmt *sqlx.NamedStmt
	if stmt, err = t.ParseAndPrepareNamed(sqlOrTpl, arg); err != nil {
		return
	}
	defer func() {
		stErr := stmt.Close()
		if stErr != nil {
			err = stErr
		}
	}()
	return fn(stmt)
}

// RunCurrentPrepareNamed use current tpl to prepare named statement
func (t *Tx) RunCurrentPrepareNamed(arg any, fn func(*sqlx.NamedStmt) error) (err error) {
	return t.RunPrepareNamedEx(t.tpl, arg, fn)
}

func (t *Tx) ParseAndPrepare(sqlOrTpl string, arg any) (*sqlx.Stmt, error) {
	query, err := t.Parse(sqlOrTpl, arg)
	if err != nil {
		return nil, err
	}
	return t.PrepareEx(query)
}
func (t *Tx) RunPreparedEx(sqlOrTpl string, arg any, fn func(*sqlx.Stmt) error) (err error) {
	var stmt *sqlx.Stmt
	if stmt, err = t.ParseAndPrepare(sqlOrTpl, arg); err != nil {
		return
	}
	defer func() {
		stErr := stmt.Close()
		if stErr != nil {
			err = stErr
		}
	}()
	return fn(stmt)
}
func (t *Tx) RunCurrentPrepared(arg any, fn func(*sqlx.Stmt) error) (err error) {
	return t.RunPreparedEx(t.tpl, arg, fn)
}

func (t *Tx) PrepareNamed(query string) (*sqlx.NamedStmt, error) {
	log.Debug("prepare named query:", query)
	return t.Tx.PrepareNamed(query)
}
func (t *Tx) PrepareEx(query string) (*sqlx.Stmt, error) {
	log.Debug("prepare query:", query)
	return t.Tx.Preparex(query)
}

// NamedExecEx  use tpl to query named statement
func (t *Tx) NamedExecEx(sqlOrTpl string, arg interface{}) (sql.Result, error) {
	query, err := t.Parse(sqlOrTpl, arg)
	if err != nil {
		return nil, err
	}
	log.Debug("named exec tpl:", query, arg)
	return t.NamedExec(query, arg)
}

func (t *Tx) ExecEx(sqlOrTpl string, args ...interface{}) (sql.Result, error) {
	query, err := t.Parse(sqlOrTpl, args)
	if err != nil {
		return nil, err
	}
	log.Debug("exec query:", query, args)
	return t.Exec(query, args...)
}

// ExecCurrent use current tpl to exec
func (t *Tx) ExecCurrent(args ...interface{}) (sql.Result, error) {
	return t.ExecEx(t.tpl, args...)
}

// NamedExecCurrent use current tpl to exec named statement
func (t *Tx) NamedExecCurrent(arg interface{}) (sql.Result, error) {
	return t.NamedExecEx(t.tpl, arg)
}

func (t *Tx) GetEx(dest any, tpl string, args ...any) error {
	query, err := t.Parse(tpl, args)
	if err != nil {
		return err
	}
	log.Debug("get query:", query, args)
	return t.Get(dest, query, args...)
}

func NewTxWith(tx *sqlx.Tx, d *DB, tpl string) *Tx {
	return &Tx{
		Tx:  tx,
		db:  d,
		tpl: tpl,
	}
}
