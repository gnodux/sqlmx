/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlmx

import (
	"database/sql"
	"fmt"
	"github.com/cookieY/sqlx"
	"github.com/gnodux/sqlmx/dialect"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"text/template"
	"time"
)

type MyMapper struct {
	*DB
	ListAll       SelectFunc[*User]     `sql:"examples/select_users.sql"`
	ListAdminUser SelectFunc[User]      `sql:"select * from user where role = 'admin' and name like ?"`
	ListUserBy    NamedSelectFunc[User] `sql:"examples/select_user_where.sql"`
	//没有tag,自动根据名称寻找模版
	ListUserByIds   NamedSelectFunc[User]
	GetById         GetFunc[User]         `sql:"examples/get_user_by_id.sql"`
	GetPtrById      GetFunc[*User]        `sql:"examples/get_user_by_id.sql"`
	GetByNamedId    NamedGetFunc[User]    `sql:"examples/get_user_by_id_name.sql"`
	GetPtrByNamedId NamedGetFunc[User]    `sql:"examples/get_user_by_id_name.sql"`
	ListUserByName  NamedSelectFunc[User] `sql:"examples/select_user_by_name.sql"`
	GetUserCount    GetFunc[int]          `sql:"examples/count_user.sql"`
	AddBy           ExecFunc              `sql:"examples/insert_users.sql"`
	Add             NamedExecFunc         `sql:"examples/insert_users.sql"`
	BatchAddUser    TxFunc
}

func TestMapper(t *testing.T) {
	d1, err := NewMapper[MyMapper](DefaultName)
	assert.NoError(t, err, "create a mapper")
	tests := []struct {
		Name    string
		fn      func() (any, error)
		wantErr bool
		err     error
	}{
		{
			Name: "select",
			fn: func() (any, error) {
				var testUser []*User
				err := d1.Select(&testUser, "select * from `user` limit 1")
				return testUser, err
			},
		}, {
			Name: "list all",
			fn: func() (any, error) {
				return d1.ListAll()
			},
		}, {
			Name: "list admin user",
			fn: func() (any, error) {
				return d1.ListAdminUser("user_1%")
			},
		}, {
			Name: "list user by nil",
			fn: func() (any, error) {
				return d1.ListUserBy(nil)
			},
		}, {
			Name: "get user by id",
			fn: func() (any, error) {
				return d1.GetById(1)
			},
		}, {
			Name: "get user by id(Pointer)",
			fn: func() (any, error) {
				return d1.GetPtrById(1)
			},
		}, {
			Name: "get user by id(not exists)",
			fn: func() (any, error) {
				return d1.GetById(-1)
			},
			wantErr: true,
			err:     sql.ErrNoRows,
		}, {
			Name: "get user by id (named args)",
			fn: func() (any, error) {
				return d1.GetByNamedId(map[string]any{
					"id": 1,
				})
			},
		}, {
			Name: "get user(pointer) by id (named args)",
			fn: func() (any, error) {
				return d1.GetPtrByNamedId(map[string]any{
					"id": 1,
				})
			},
		}, {
			Name: "get user count",
			fn: func() (any, error) {
				return d1.GetUserCount()
			},
		}, {
			Name: "get user by name filter",
			fn: func() (any, error) {
				return d1.ListUserByName(&User{Name: "%user_1%"})
			},
		},
		{
			Name: "list user by ids",
			fn: func() (any, error) {
				return d1.ListUserByIds(map[string]any{
					"ids": []interface{}{
						12, 18,
					},
				})
			},
		}, {
			Name: "batch add user",
			fn: func() (any, error) {
				var ids []int64
				err := d1.BatchAddUser(func(tx *Tx) error {
					return tx.RunPrepareNamedEx("initialize/insert_user.sql", map[string]any{}, func(stmt *sqlx.NamedStmt) error {
						for _, user := range []*User{
							{
								Name:     "batch_user_1",
								Role:     "admin",
								Birthday: time.Now(),
							}, {
								Name:     "batch_user_2",
								Role:     "user",
								Birthday: time.Now(),
							}, {
								Name:     "batch_user_3",
								Role:     "user",
								Birthday: time.Now(),
							},
						} {
							var result sql.Result
							if result, err = stmt.Exec(user); err != nil {
								return err
							} else {
								if id, err := result.LastInsertId(); err == nil {
									ids = append(ids, id)
								}

							}
						}
						return nil
					})
				})
				if len(ids) > 0 {
					query, err := d1.ParseSQL("examples/delete_user_by_ids.sql", nil)
					if err != nil {
						return nil, err
					}
					var param []any
					query, param, err = sqlx.In(query, ids)
					if err != nil {
						return nil, err
					}
					result, err := d1.Exec(query, param...)
					return result, err
				}

				return nil, err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			ret, err := tt.fn()
			if tt.wantErr {
				assert.Error(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}
			_ = encoder.Encode(ret)
		})
	}

}

func TestParseTemplate(t *testing.T) {
	tpl := template.New("sql").Funcs(MakeFuncMap(dialect.MySQL))
	_, err := tpl.ParseFS(os.DirFS("./testdata/"), "**/*.sql")
	assert.NoError(t, err)
	_, err = tpl.New("test/job").Parse("{{.Name}}:{{.Cat}}({{list .tags}})")
	assert.NoError(t, err)
	for _, tp := range tpl.Templates() {
		t.Log(tp.Name())
	}

	err = tpl.ExecuteTemplate(os.Stdout, "get_user_by_id_name.sql", map[string]any{
		"id": 1000,
	})
	fmt.Println()
	assert.NoError(t, err)
}
