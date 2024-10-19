/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlmx

import (
	"context"
	"database/sql"
	"errors"
	"github.com/cookieY/sqlx"
	"github.com/gnodux/sqlmx/dialect"
	"github.com/gnodux/sqlmx/expr"
	"io/fs"
	"strings"
	"sync"
	"text/template"
)

var (
	ErrNilDriver = errors.New("driver is nil")
	ErrNilDB     = errors.New("DB is nil")
)

// DB 数据库连接
// 对sqlx.DB的封装，提供更多的功能：
// 1. 通过模板文件生成SQL语句并执行，每个连接实例都可以拥有自己的模版系统
// 2. 通过expr包提供的表达式语法生成SQL语句并执行
type DB struct {
	m        *DBManager
	template *template.Template
	lock     sync.Mutex
	driver   *dialect.Dialect
	*sqlx.DB
}

func (d *DB) SetManager(m *DBManager) {
	d.m = m
}

func (d *DB) PrepareEx(sqlOrTpl string, args any) (*sqlx.Stmt, error) {
	if d == nil {
		return nil, ErrNilDB
	}
	query, err := d.ParseSQL(sqlOrTpl, args)
	if err != nil {
		return nil, err
	}
	return d.Preparex(query)
}

func (d *DB) RunPrepared(sqlOrTpl string, arg any, fn func(*sqlx.Stmt) error) (err error) {
	var stmt *sqlx.Stmt
	if strings.HasSuffix(sqlOrTpl, ".sql") {
		if stmt, err = d.PrepareEx(sqlOrTpl, arg); err != nil {
			return
		}
	} else {
		if stmt, err = d.Preparex(sqlOrTpl); err != nil {
			return
		}
	}
	defer func() {
		stErr := stmt.Close()
		if stErr != nil {
			err = stErr
		}
	}()
	return fn(stmt)
}

func (d *DB) PrepareNamedEx(tplName string, args any) (*sqlx.NamedStmt, error) {
	if d == nil {
		return nil, ErrNilDB
	}
	var (
		query string
		err   error
	)
	if strings.HasSuffix(tplName, ".sql") {
		query, err = d.ParseSQL(tplName, args)
	} else {
		query = tplName
	}
	if err != nil {
		return nil, err
	}
	return d.PrepareNamed(query)
}

// RunPrepareNamed run prepared statement with named args
// arg 如果是模版，是模版渲染参数，如果是动态SQL，则不需要(根据传入名称是否以.sql结尾判断)
func (d *DB) RunPrepareNamed(sqlOrTpl string, arg any, fn func(*sqlx.NamedStmt) error) (err error) {
	var stmt *sqlx.NamedStmt
	if strings.HasSuffix(sqlOrTpl, ".sql") {
		if stmt, err = d.PrepareNamedEx(sqlOrTpl, arg); err != nil {
			return
		}
	} else {
		if stmt, err = d.PrepareNamed(sqlOrTpl); err != nil {
			return
		}
	}
	defer func() {
		stErr := stmt.Close()
		if stErr != nil {
			err = stErr
		}
	}()
	return fn(stmt)
}
func (d *DB) SelectEx(dest interface{}, sqlOrTpl string, args ...any) error {
	if d == nil {
		return ErrNilDB
	}
	query, err := d.ParseSQL(sqlOrTpl, args)
	if err != nil {
		return err
	}
	log.Debug("select:", query, args)
	return d.DB.Select(dest, query, args...)
}
func (d *DB) NamedSelectEx(dest interface{}, sqlOrTpl string, args interface{}) (err error) {
	if d == nil {
		return ErrNilDB
	}
	var named *sqlx.NamedStmt
	named, err = d.PrepareNamedEx(sqlOrTpl, args)
	if err != nil {
		return err
	}
	defer func(named *sqlx.NamedStmt) {
		err = named.Close()

	}(named)
	if args == nil {
		args = map[string]any{}
	}
	log.Debug("named select tpl:", named.QueryString, args)
	return named.Select(dest, args)
}
func (d *DB) NamedSelect(dest interface{}, sql string, arg any) (err error) {
	if d == nil {
		return ErrNilDB
	}
	var named *sqlx.NamedStmt
	named, err = d.PrepareNamed(sql)
	if err != nil {
		return err
	}
	defer func(named *sqlx.NamedStmt) {
		err = named.Close()

	}(named)
	log.Debug("named select:", named.QueryString, arg)
	return named.Select(dest, arg)
}
func (d *DB) NamedExecEx(sqlOrTpl string, arg interface{}) (sql.Result, error) {
	if d == nil {
		return nil, ErrNilDB
	}
	query, err := d.ParseSQL(sqlOrTpl, arg)
	if err != nil {
		return nil, err
	}
	log.Debug("named exec:", query, arg)
	return d.NamedExec(query, arg)
}

func (d *DB) ExecEx(sqlOrTpl string, args ...interface{}) (sql.Result, error) {
	if d == nil {
		return nil, ErrNilDB
	}
	query, err := d.ParseSQL(sqlOrTpl, args)
	if err != nil {
		return nil, err
	}
	log.Debug("exec:", query, args)
	return d.Exec(query, args...)
}
func (d *DB) NamedQueryEx(sqlOrTpl string, arg interface{}) (*sqlx.Rows, error) {
	if d == nil {
		return nil, ErrNilDB
	}
	query, err := d.ParseSQL(sqlOrTpl, arg)
	if err != nil {
		return nil, err
	}
	log.Debug("named query:", query, arg)
	return d.NamedQuery(query, arg)
}
func (d *DB) Batch(ctx context.Context, opts *sql.TxOptions, fn func(tx *Tx) error) (err error) {
	return d.BatchEx(ctx, opts, "", fn)
}

func (d *DB) BatchEx(ctx context.Context, opts *sql.TxOptions, tpl string, fn func(tx *Tx) error) (err error) {
	if d == nil {
		return ErrNilDB
	}
	var tx *sqlx.Tx
	tx, err = d.BeginTxx(ctx, opts)
	if err != nil {
		return err
	}
	defer func() {
		if tx != nil {
			if err != nil {
				_ = tx.Rollback()
			} else {
				err = tx.Commit()
			}
		}
	}()
	if err = fn(NewTxWith(tx, d, tpl)); err != nil {
		return
	}
	return
}

// SelectExpr 使用表达式进行查询
func (d *DB) SelectExpr(dest interface{}, exp expr.Expr) error {
	if d == nil {
		return ErrNilDB
	}
	buff := expr.NewTracedBuffer(d.driver)
	if d.driver.SupportNamed {
		query, namedArgs, err := buff.BuildNamed(exp)
		if err != nil {
			return err
		}
		return d.NamedSelect(dest, query, namedArgs)
	} else {
		query, args, err := buff.Build(exp)
		if err != nil {
			return err
		}
		return d.Select(dest, query, args...)
	}
}

// ExecExpr 使用表达式进行执行
func (d *DB) ExecExpr(exp expr.Expr) (sql.Result, error) {
	if d == nil {
		return nil, ErrNilDB
	}
	buff := expr.NewTracedBuffer(d.driver)
	if d.driver.SupportNamed {
		query, namedArgs, err := buff.BuildNamed(exp)
		if err != nil {
			return nil, err
		}
		return d.NamedExec(query, namedArgs)
	} else {
		query, args, err := buff.Build(exp)
		if err != nil {
			return nil, err
		}
		return d.Exec(query, args...)
	}
}

func (d *DB) GetExpr(dest interface{}, exp expr.Expr, filters ...expr.FilterFn) error {
	if d == nil {
		return ErrNilDB
	}
	for _, filter := range filters {
		filter(exp)
	}
	buff := expr.NewTracedBuffer(d.driver)
	if d.driver.SupportNamed {
		query, namedArgs, err := buff.BuildNamed(exp)
		if err != nil {
			return err
		}
		return d.NamedGet(dest, query, namedArgs)
	} else {
		query, args, err := buff.Build(exp)
		if err != nil {
			return err
		}
		return d.Get(dest, query, args...)
	}
}

func (d *DB) NamedGet(dest interface{}, query string, arg interface{}) error {
	stmt, err := d.PrepareNamed(query)
	if err != nil {
		return err
	}
	defer func() {
		_ = stmt.Close()
	}()
	return stmt.Get(dest, arg)
}

// SetTemplate set template
func (d *DB) SetTemplate(tpl *template.Template) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.template = tpl
}
func (d *DB) Template() *template.Template {
	return d.template
}

// ParseTemplateFS parse template from filesystem。
// 为了保留目录结构，没有直接使用template的ParseFS(template中的ParseFS方法不会保留路径名称)
func (d *DB) ParseTemplateFS(f fs.FS, patterns ...string) error {
	log.Info("parse template from filesystem: ", f, " with patterns:", patterns)
	for _, pattern := range patterns {
		matches, err := fs.Glob(f, pattern)
		if err != nil {
			return err
		}
		for _, mf := range matches {
			buf, err := fs.ReadFile(f, mf)
			if err != nil {
				return err
			}
			log.Info("parse sql:", mf)
			if _, err = d.template.New(strings.ReplaceAll(mf, "\\", "/")).Parse(string(buf)); err != nil {
				return err
			}
		}
	}
	return nil
}
func (d *DB) MustParseTemplateFS(f fs.FS, patterns ...string) *DB {
	err := d.ParseTemplateFS(f, patterns...)
	if err != nil {
		panic(err)
	}
	return d
}

// ParseTemplate parse template from string
func (d *DB) ParseTemplate(name string, tpl string) (*template.Template, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	t, err := d.template.New(name).Parse(tpl)
	return t, err
}

// ParseSQL parse sql from template
// 2023-7-12: 由于template的Parse方法会将{{}}中的内容当作变量，所以不再使用template.Parse方法,由BoostMapper中预先解析，减小运行时性能消耗和锁定
func (d *DB) ParseSQL(sqlOrTpl string, args any) (query string, err error) {

	if !strings.HasSuffix(sqlOrTpl, sqlSuffix) {
		return sqlOrTpl, nil
	}
	sb := &strings.Builder{}
	err = d.template.ExecuteTemplate(sb, sqlOrTpl, args)
	if err == nil {
		query = sb.String()
	}
	//}
	log.Trace("parse sql:", sqlOrTpl, "=>", query, " with args:", args)
	return
}
