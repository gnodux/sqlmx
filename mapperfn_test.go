/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlmx

import (
	"github.com/gnodux/sqlmx/utils"
	"testing"
)

func TestLowerCase(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"test1", args{"HelloWorld"}, "hello_world"},
		{"test2", args{"HelloWorldHelloWorld"}, "hello_world_hello_world"},
		{"test3", args{"UserId"}, "user_id"},
		{"test4", args{"UserID"}, "user_id"},
		{"test5", args{"Tenant.TenantID"}, "tenant.tenant_id"},
		{"test6", args{"CreateDATE"}, "create_date"},
		{"test7", args{"ID"}, "id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.LowerCase(tt.args.s); got != tt.want {
				t.Errorf("LowerCase() = %v, want %v", got, tt.want)
			}
		})
	}
}
