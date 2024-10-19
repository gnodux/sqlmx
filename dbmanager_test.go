/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlmx

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gnodux/sqlmx/utils"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"os"
	"testing"
	"time"
)

var (
	encoder   = json.NewEncoder(os.Stdout)
	currentDB = DefaultName
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.TraceLevel)
	encoder.SetIndent("", "  ")
	//start mysql test
	currentDB = DefaultName
	MustOpen(currentDB, "mysql", "xxtest:xxtest@tcp(localhost)/sqlmx?charset=utf8&parseTime=true&multiStatements=true").
		ParseTemplateFS(os.DirFS("./testdata"), "examples/*.sql", "initialize/*.sql", "my_mapper/*.sql")
	SetTemplateFS(os.DirFS("./testdata"), "examples/*.sql", "initialize/*.sql", "my_mapper/*.sql")
	initData()
	m.Run()
	//Manager.Shutdown()
	////start postgres test
	//SetConstructor(DefaultName, func() (*DB, error) {
	//	d, err := Open("postgres", "postgres://xxtest:xxtest@localhost:15432/sqlmx?sslmode=disable")
	//	if err == nil {
	//		d.ParseTemplateFS(os.DirFS("./testdata"), "examples/*.sql", "initialize/*.sql", "my_mapper/*.sql")
	//	}
	//	return d, err
	//})
	////SetTemplateFS(os.DirFS("./testdata"), "examples/*.sql", "initialize/*.sql", "my_mapper/*.sql")
	//
	//initData()
	//m.Run()
}
func TestSimple(t *testing.T) {
	rows, err := MustGet(DefaultName).Query("SELECT * FROM tenant")
	println(rows)
	if err != nil {
		t.Fatal(err)
	}
}
func initData() {
	err := MustGet(currentDB).Batch(context.Background(), &sql.TxOptions{ReadOnly: false, Isolation: sql.LevelReadCommitted}, func(tx *Tx) error {
		if _, err := tx.ExecEx("initialize/create_tables.sql"); err != nil {
			return err
		}
		count := 0

		if err := tx.Get(&count, "SELECT COUNT(1) FROM tenant"); err != nil {
			return err
		}
		if count == 0 {
			if _, err := tx.NamedExecEx("initialize/insert_tenant.sql", Tenant{
				Name: "test tenant",
			}); err != nil {
				return err
			}
		}

		if err := tx.GetEx(&count, "examples/count_user.sql"); err != nil {
			return err
		}
		if count == 0 {
			//add users
			for i := 0; i < 100; i++ {
				if _, err := tx.NamedExecEx("initialize/insert_user.sql", User{
					Name:     fmt.Sprintf("user_%d", i),
					TenantID: 1,
					Password: "password",
					Birthday: time.Now(),
					Address:  fmt.Sprintf("address %d", i),
					Role:     "admin",
				}); err != nil {
					return err
				}
			}
		}
		if err := tx.GetEx(&count, "examples/count_role.sql"); err != nil {
			return err
		}
		if count == 0 {
			//add roles
			utils.Must(tx.NamedExecEx("initialize/insert_role.sql", Role{
				Name: "admin",
				Desc: "system administrator",
			}))

			utils.Must(tx.NamedExecEx("initialize/insert_role.sql", Role{
				Name: "user",
				Desc: "normal user",
			}))
			utils.Must(tx.NamedExecEx("initialize/insert_role.sql", Role{
				Name: "customer",
				Desc: "customer",
			}))
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}
func TestSelectUsers(t *testing.T) {
	fn := NewSelectFunc[User](currentDB, "examples/select_users.sql")
	users, err := fn()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(users)
}
func TestSelectUserLikeName(t *testing.T) {
	fn := NewNamedSelectFunc[User](currentDB, "examples/select_user_where.sql")
	users, err := fn(User{
		Name: "user_6",
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(users)
}

func TestNilDB(t *testing.T) {
	var d *DB
	var u []User
	err := d.SelectEx(&u, "select_users.sql")
	fmt.Println(err)
}
func TestMustGet(t *testing.T) {
	encoder.Encode(utils.Must(Get(currentDB)))
}
