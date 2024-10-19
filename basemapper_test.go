/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlmx

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gnodux/sqlmx/expr"
	"github.com/gnodux/sqlmx/meta"
	"github.com/gnodux/sqlmx/utils"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestBatchInsert(t *testing.T) {
	var mapper *BaseMapper[*User]
	var err error

	mapper, err = NewMapper[BaseMapper[*User]](currentDB)
	assert.NoError(t, err)
	maxRecord := 1000
	users := make([]*User, maxRecord)
	for i := 0; i < maxRecord; i++ {
		users[i] = &User{
			TenantID: 10011002,
			Name:     fmt.Sprintf("test user%d", i),
			Password: fmt.Sprintf("%d", rand.Int63n(99999)),
			Birthday: time.Now(),
			Address:  "Room 404,Build 401, 302 Road,Beijing",
			Role:     "user",
		}
	}
	err = mapper.Insert(users...)
	err = mapper.Batch(context.Background(), &sql.TxOptions{Isolation: sql.LevelDefault}, func(tx *Tx) error {
		for _, u := range users {
			if _, err = tx.NamedExec("delete from user where id=:id", u); err != nil {
				return err
			}

		}
		return nil
	})
	assert.NoError(t, err)
}

func TestBaseMapper_Pointer(t *testing.T) {
	mapper, err := NewMapper[BaseMapper[*User]](currentDB)
	assert.NoError(t, err)

	user := User{
		TenantID: 10011002,
		Name:     "test user1",
		Password: fmt.Sprintf("%d", rand.Int63n(99999)),
		Birthday: time.Now(),
		Address:  "Room 404,Build 401, 302 Road,Beijing",
		Role:     "user",
	}
	tests := []struct {
		Name string
		fn   func() (any, error)
	}{
		{
			Name: "create user",
			fn: func() (any, error) {
				err = mapper.Create(&user)
				t.Log(user.ID, user.TenantID)
				return user, err
			},
		}, {
			Name: "list by id",
			fn: func() (any, error) {
				return mapper.ListById(0, 1, 2, 3, 4, 5)
			},
		}, {
			Name: "delete by id",
			fn: func() (any, error) {
				nu := user
				nu.Name = "test user2"
				err = mapper.Create(&nu)
				encoder.Encode(nu)
				err = mapper.DeleteById(nu.TenantID, nu.ID)
				if err != nil {
					return nil, err
				}
				err = mapper.EraseById(nu.TenantID, nu.ID)
				return nil, err
			},
		}, {
			Name: "update user",
			fn: func() (any, error) {
				if err = mapper.Create(&user); err != nil {
					return nil, err
				}
				user.Role = "test2"
				if err = mapper.Update(true, &user); err != nil {
					return user, err
				}
				if err = mapper.DeleteById(user.TenantID, user.ID); err != nil {
					return nil, err
				}
				return user, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result, err := tt.fn()
			_ = encoder.Encode(result)
			assert.NoError(t, err)
		})
	}
}

func TestBaseMapper_Struct(t *testing.T) {
	mapper, err := NewMapper[BaseMapper[User]](currentDB)
	assert.NoError(t, err)
	user := User{
		TenantID: 10011002,
		Name:     "test user1",
		Password: fmt.Sprintf("%d", rand.Int63n(99999)),
		Birthday: time.Now(),
		Address:  "Room 404,Build 401, 302 Road,Beijing",
		Role:     "user",
	}
	tests := []struct {
		Name string
		fn   func() (any, error)
	}{
		{
			Name: "create user",
			fn: func() (any, error) {
				err = mapper.Create(user)
				t.Log(user.ID, user.TenantID)
				mapper.DeleteById(user.TenantID, user.ID)
				return user, err
			},
		}, {
			Name: "list by id",
			fn: func() (any, error) {
				return mapper.ListById(0, 1, 2, 3, 4, 5)
			},
		}, {
			Name: "delete by id",
			fn: func() (any, error) {
				nu := user
				nu.Name = "test user2"
				err = mapper.Create(nu)
				encoder.Encode(nu)
				err = mapper.DeleteById(nu.TenantID, nu.ID)
				return nil, err
			},
		}, {
			Name: "update user",
			fn: func() (any, error) {
				if err = mapper.Create(user); err != nil {
					return nil, err
				}
				user.Role = "test2"
				if err = mapper.Update(true, user); err != nil {
					return user, err
				}
				if err = mapper.DeleteById(user.TenantID, user.ID); err != nil {
					return nil, err
				}
				return user, nil
			},
		},
		//{
		//	Name: "select by name",
		//	fn: func() (any, error) {
		//		users, err := mapper.SelectBy(map[string]any{
		//			"Name": "user_2%",
		//		}, expr.SimpleDesc("Name", "Role"), 100, 10)
		//		return users, err
		//	},
		//},
		//{
		//	Name: "count by name",
		//	fn: func() (any, error) {
		//		total, err := mapper.CountBy(map[string]any{
		//			"Name": "user_2%",
		//		})
		//		return total, err
		//	},
		//},
		{
			Name: "select by example",
			fn: func() (any, error) {
				users, _, err := mapper.SelectByExample(User{
					Address: "%57",
					Name:    "user_%",
				}, expr.UseLimit(10))
				return users, err
			},
		},
		{
			Name: "update by example",
			fn: func() (any, error) {
				count, err := mapper.UpdateByExample(User{
					Address: "Test address ...	",
					Name:    "MyUserName",
				}, User{
					ID:       1,
					TenantID: 100102,
				})
				return count, err
			},
		}, {
			Name: "delete by example",
			fn: func() (any, error) {
				count, err := mapper.DeleteByExample(User{
					ID:       1000,
					TenantID: 888123,
				})
				return count, err
			},
		},
		{
			Name: "Partial update user",
			fn: func() (any, error) {
				u := User{
					TenantID: 10011002,
					Name:     "test user1",
					Password: fmt.Sprintf("%d", rand.Int63n(99999)),
					Birthday: time.Now(),
					Address:  "Room 404,Build 401, 302 Road,Beijing",
					Role:     "user",
				}
				if err = mapper.Create(u); err != nil {
					return nil, err
				}
				if err = mapper.PartialUpdate(false, nil, User{
					ID:      u.ID,
					Address: "Room 412,DONGFENG KASO,JIUXIANQIAO Road,ChaoYang, BeiJing",
				}); err != nil {
					return nil, err
				}
				if err = mapper.AutoPartialUpdate(true, u); err != nil {
					return nil, err
				}
				err = mapper.DeleteById(u.TenantID, u.ID)
				return nil, err
			},
		}, {
			Name: "query test 1",
			fn: func() (any, error) {
				result, total, err := mapper.Select(expr.SelectFilter(func(e *expr.SelectExpr) {
					e.Where(
						expr.And(
							expr.Eq(expr.Name("name"),
								expr.Var("name", "test user1")))).Limit(10).Offset(0)
				}), expr.WithCount)
				t.Log("total", total)
				return result, err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result, err := tt.fn()
			_ = encoder.Encode(result)
			assert.NoError(t, err)
		})
	}
}

type UserMapper struct {
	User BaseMapper[User]
	Role BaseMapper[Role]
}

type RoleExtMapper struct {
	BaseMapper[Role]
	ListByRoleName SelectFunc[Role]       `sql:"select * from role where name = ?"`
	ListByRole     NamedSelectFunc[*Role] `sql:"select * from role where name = :name"`
	ListUsers      NamedSelectFunc[*User] `sql:"select * from user where role = :name"`
}

func TestExtMapper(t *testing.T) {
	m, err := NewMapper[RoleExtMapper](currentDB)
	assert.NoError(t, err, "create mapper error")
	users, err := m.ListByRoleName("admin")
	assert.NoError(t, err, "list by role error")
	encoder.Encode(users)
	roles, err := m.ListByRole(&Role{Name: "admin"})
	assert.NoError(t, err, "list by role error")
	encoder.Encode(roles)

	roles, err = m.ListByRole(&Role{Name: "admin"})
	assert.NoError(t, err, "list by role error")
	encoder.Encode(roles)
	adminUsers, err := m.ListUsers(&Role{Name: "admin"})
	assert.NoError(t, err, "list by role error")
	encoder.Encode(adminUsers)
}

func TestBaseMapper_Duck(t *testing.T) {
	m, err := NewMapper[UserMapper](currentDB)
	assert.NoError(t, err)
	users, _, err := m.User.SelectByExample(User{
		Name: "user_1%",
		Role: "admin",
	}, expr.UseLimit(1), expr.AutoFuzzy)
	assert.NoError(t, err)
	encoder.Encode(users)
}

type Std struct {
	CreateTime time.Time
	CreateBy   int64
	ModifyTime time.Time
	ModifyBy   int64
	IsDeleted  bool
}

type MyUser struct {
	UserName string
	Password string
	Std
	Transaction
}

func Test_listColumns(t *testing.T) {
	my := MyUser{}
	cols := meta.ListColumns(reflect.TypeOf(my))
	err := encoder.Encode(cols)
	if err != nil {
		t.Fatal(err)
	}
}

type UserExt struct {
	User
	Nation string `json:"nation"`
	Phone  string `json:"phone"`
}

func Test_ToMap(t *testing.T) {
	m := map[string]interface{}{
		"Name":    "test",
		"ID":      int64(9000912),
		"Address": "Room 404,Build 401, 302 Road,Beijing",
		"Nation":  "China",
		"Phone":   "123456789",
	}
	ue := &UserExt{
		User: User{
			Name:    "test",
			ID:      9000912,
			Address: "Room 404,Build 401, 302 Road,Beijing",
		},
		Nation: "China",
		Phone:  "123456789",
	}
	assert.Equal(t, m, utils.ToMap(ue))
	encoder.Encode(utils.ToMap(ue))
}

func TestBaseMapper_Insert(t *testing.T) {
	mapper, err := NewMapper[BaseMapper[*User]](currentDB)
	if err != nil {
		t.Fatal(err)
	}
	newUser := User{
		Name:     "test user1",
		TenantID: 100102,
		Password: fmt.Sprintf("%d", rand.Int63n(99999)),
		Birthday: time.Now(),
		Address:  "Room 404,Build 401, 302 Road,Beijing",
		Role:     "user",
	}

	err = mapper.Insert(&newUser)
	assert.NoError(t, err)
	assert.Greater(t, newUser.ID, int64(0))
	encoder.Encode(newUser)
	effect, err := mapper.DeleteBy(func(e *expr.DeleteExpr) {
		e.Where(expr.N("id").Eq(newUser.ID))
	})
	assert.NoError(t, err)
	if effect == 0 {
		t.Fatal("delete error")
	}

}
