/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlmx

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"text/template"

	"github.com/gnodux/sqlmx/dialect"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestTemplateFunc(t *testing.T) {

	type args struct {
		tpl string
		arg any
	}
	var tests []struct {
		name string
		args args
		want string
	}
	tpl := template.New("tests").Funcs(MakeFuncMap(dialect.MySQL))
	_, err := tpl.ParseFS(os.DirFS("testdata/examples"), "*.sql")
	assert.NoError(t, err)

	// 初始化测试用例
	tests = []struct {
		name string
		args args
		want string
	}{
		// TODO: 在此添加测试用例
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &strings.Builder{}
			err := tpl.ExecuteTemplate(buf, tt.args.tpl, tt.args.arg)
			assert.NoError(t, err)
			if tt.want != "" {
				assert.Equal(t, tt.want, buf.String())
			}
			fmt.Println(buf)
		})
	}
}
func TestMapLength(t *testing.T) {
	m := map[string]any{}
	mv := reflect.ValueOf(m)
	if mv.IsZero() {
		fmt.Println("zero")
	}
	fmt.Println(mv.Len())
}

//
//func TestPg(t *testing.T) {
//	c, err := sql.Open("postgres", "")
//}
