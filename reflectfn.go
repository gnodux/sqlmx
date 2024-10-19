/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlmx

import (
	"github.com/cookieY/sqlx"
	"reflect"
)

func getTpl(d *DB, templateList []string) string {
	if len(templateList) == 1 {
		return templateList[0]
	} else {
		for _, tpl := range templateList {
			if d.template.Lookup(tpl) != nil {
				return tpl
			}
		}
	}
	return ""
}
func SelectWith(p reflect.Type, db *DB, templateList []string, args []any) (any, error) {
	list := reflect.New(reflect.SliceOf(p))
	tpl := getTpl(db, templateList)
	err := db.SelectEx(list.Interface(), tpl, args...)
	if err != nil {
		return nil, err
	}

	return list.Elem().Interface(), err
}

func NamedSelectWith(p reflect.Type, db *DB, templateList []string, arg any) (any, error) {
	list := reflect.New(reflect.SliceOf(p))
	tpl := getTpl(db, templateList)
	err := db.NamedSelectEx(list.Interface(), tpl, arg)
	return list.Elem().Interface(), err
}

func NamedGetWith(p reflect.Type, db *DB, templateList []string, arg any) (any, error) {
	var o reflect.Value
	if p.Kind() == reflect.Pointer {
		o = reflect.New(p.Elem())
	} else {
		o = reflect.New(p)
	}
	tpl := getTpl(db, templateList)
	n, err := db.PrepareNamedEx(tpl, arg)
	if err != nil {
		return nil, err
	}
	defer func(n *sqlx.NamedStmt) {
		err = n.Close()
	}(n)
	err = n.Get(o.Interface(), arg)
	if p.Kind() == reflect.Pointer {
		return o.Interface(), err
	} else {
		return o.Elem().Interface(), err
	}
}

func GetWith(p reflect.Type, db *DB, templateList []string, args []any) (any, error) {
	var o reflect.Value
	if p.Kind() == reflect.Pointer {
		o = reflect.New(p.Elem())
	} else {
		o = reflect.New(p)
	}
	tpl := getTpl(db, templateList)
	n, err := db.PrepareEx(tpl, args)
	if err != nil {
		return nil, err
	}
	defer func(n *sqlx.Stmt) {
		err = n.Close()
	}(n)
	err = n.Get(o.Interface(), args...)
	if p.Kind() == reflect.Pointer {
		return o.Interface(), err
	} else {
		return o.Elem().Interface(), err
	}
}
