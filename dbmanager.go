/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlmx

import (
	"fmt"
	"github.com/cookieY/sqlx"
	"github.com/gnodux/sqlmx/builtin"
	"github.com/gnodux/sqlmx/dialect"
	"github.com/gnodux/sqlmx/utils"
	"io/fs"
	"sync"
	"text/template"
)

var (
	DefaultName = "Default"
	//Manager default connection manager
	Manager = NewDBManager(DefaultName)
	//Get a db from
	Get = Manager.Get
	//MustGet a db,if db not exists,raise a panic
	MustGet = Manager.MustGet
	//Set a db
	Set = Manager.Set
	//SetWithConnFunc set a db with constructors func
	SetWithConnFunc = Manager.SetWithConnFunc
	//OpenWithConnFunc initialize a db
	OpenWithConnFunc = Manager.OpenWithConnFunc

	//Open a db
	Open = Manager.Open
	//MustOpen a db,if db not exists,raise a panic
	MustOpen = Manager.MustOpen

	//OpenDefault open a db with default name
	OpenDefault = Manager.OpenDefault
	//OpenWith open a db with driver and datasource
	OpenWith = Manager.OpenWith
	//SetTemplateFS set sql template from filesystem
	SetTemplateFS = Manager.SetTemplateFS

	SetDialect = Manager.SetDefaultDialect
	//ClearTemplateFS clear sql template from filesystem
	ClearTemplateFS = Manager.ClearTemplateFS

	//Shutdown manager and close all db
	Shutdown = Manager.Shutdown

	////SetTemplate set sql template
	//SetTemplate = Manager.SetTemplate

	//ParseTemplateFS set sql template from filesystem
	//ParseTemplateFS = Manager.ParseTemplateFS

	//ParseTemplate create a new template
	//ParseTemplate = Manager.ParseTemplate
)

type ConnFunc func() (*DB, error)

type TplFS struct {
	FS       fs.FS
	Patterns []string
}

type DBManager struct {
	driver       *dialect.Dialect
	name         string
	dbs          map[string]*DB
	constructors map[string]ConnFunc
	lock         *sync.RWMutex
	templateFS   []*TplFS
}

func NewManagerWithDriver(name string, driver *dialect.Dialect) *DBManager {
	f := &DBManager{
		name:         name,
		driver:       driver,
		dbs:          map[string]*DB{},
		constructors: map[string]ConnFunc{},
		lock:         &sync.RWMutex{},
	}
	return f
}
func NewDBManager(name string) *DBManager {
	return NewManagerWithDriver(name, DefaultDialect)
}

func (m *DBManager) SetTemplateFS(f fs.FS, patterns ...string) {
	m.templateFS = append(m.templateFS, &TplFS{
		FS:       f,
		Patterns: patterns,
	})
}
func (m *DBManager) ClearTemplateFS() {
	m.templateFS = nil
}

//Get 获取一个数据库连接
//name: 数据库连接名称

func (m *DBManager) Get(name string) (*DB, error) {
	conn, ok := func() (*DB, bool) {
		m.lock.RLock()
		defer m.lock.RUnlock()
		c, o := m.dbs[name]
		return c, o
	}()
	if !ok {
		if loader, lok := m.constructors[name]; lok {
			err := func() error {
				m.lock.Lock()
				defer func() {
					//无论是否成功，都移除loader，避免反复初始化导致异常
					delete(m.constructors, name)
					m.lock.Unlock()
				}()
				var err error
				conn, err = loader()
				if err != nil {
					return err
				} else {
					m.dbs[name] = conn
				}
				conn.SetManager(m)
				conn.MapperFunc(NameFunc)
				return nil
			}()
			if err != nil {
				return nil, fmt.Errorf("initialize database %s error:%s", name, err)
			}
		} else {
			return nil, fmt.Errorf("database %s not found in %s", name, m.name)
		}
	}
	return conn, nil
}

// MustGet 获取一个数据库连接，如果不存在则panic
func (m *DBManager) MustGet(name string) *DB {
	return utils.Must(m.Get(name))
}

// OpenWithConnFunc  创建新的数据库连接并放入管理器中
func (m *DBManager) OpenWithConnFunc(name string, fn ConnFunc) (*DB, error) {
	d, err := fn()
	if err != nil {
		return nil, err
	}
	m.Set(name, d)
	return d, nil
}

// Open 打开一个数据库连接
func (m *DBManager) Open(name, driverName, dsn string) (*DB, error) {
	if m.Exists(name) {
		return m.Get(name)
	}
	db, err := m.OpenWith(Dialects[driverName], dsn)
	if err != nil {
		return nil, err
	}
	m.Set(name, db)
	return db, nil
}

// MustOpen 打开一个数据库连接，如果失败则panic
func (m *DBManager) MustOpen(name, driverName, dsn string) *DB {
	return utils.Must(m.Open(name, driverName, dsn))
}
func (m *DBManager) OpenDefault(driverName, dsn string) (*DB, error) {
	return m.Open(DefaultName, driverName, dsn)
}
func (m *DBManager) Exists(name string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, ok := m.dbs[name]
	return ok
}

// Set a database
func (m *DBManager) Set(name string, db *DB) {
	m.lock.Lock()
	defer m.lock.Unlock()
	db.m = m
	m.dbs[name] = db
}

// SetWithConnFunc set a database constructor(Lazy create DB)
func (m *DBManager) SetWithConnFunc(name string, connFunc ConnFunc) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.constructors[name] = connFunc
}

func (m *DBManager) BoostMapper(dest any, dataSource string) error {
	return BoostMapper(dest, m, dataSource)
}
func (m *DBManager) Shutdown() error {
	for _, v := range m.dbs {
		if err := v.Close(); err != nil {
			return err
		}
	}
	m.dbs = map[string]*DB{}
	return nil
}

// SetDefaultDialect set default dialect
func (m *DBManager) SetDefaultDialect(driver *dialect.Dialect) {
	m.driver = driver
}
func (m *DBManager) OpenWith(curDialect *dialect.Dialect, datasource string) (*DB, error) {
	if curDialect == nil {
		curDialect = m.driver
	}
	if curDialect == nil {
		return nil, ErrNilDriver
	}
	db, err := sqlx.Open(curDialect.Name, datasource)
	if err != nil {
		return nil, err
	}
	db.MapperFunc(NameFunc)

	newDb := &DB{DB: db, m: m, driver: curDialect, template: template.New("sql").Funcs(MakeFuncMap(curDialect))}
	err = newDb.ParseTemplateFS(builtin.Builtin, "builtin/*.sql")
	if err != nil {
		return nil, err
	}
	for _, tfs := range m.templateFS {
		err = newDb.ParseTemplateFS(tfs.FS, tfs.Patterns...)
		if err != nil {
			return nil, err
		}
	}
	return newDb, err
}

func (m *DBManager) String() string {
	return fmt.Sprintf("db[%s]", m.name)
}
