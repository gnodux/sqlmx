/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package dialect

import (
	"fmt"
	"github.com/gnodux/sqlmx/utils"
)

var (
	//MySQL MySQL驱动
	MySQL = &Dialect{
		Name:         "mysql",
		SupportNamed: true,
		NamedPrefix:  ":",
		DateFormat:   "'2006-01-02 15:04:05'",
		SQLNameFunc:  MakeNameFunc("`", "`"),
		NameFunc:     utils.LowerCase,
		PlaceHolder:  "?",
	}

	//SQLServer SQLServer驱动
	SQLServer = &Dialect{
		Name:         "mssql",
		SupportNamed: true,
		NamedPrefix:  "@",
		PlaceHolder:  "?",
		DateFormat:   "'2006-01-02 15:04:05'",
		SQLNameFunc:  MakeNameFunc("[", "]"),
		NameFunc:     utils.LowerCase,
	}
	// Postgres 驱动
	Postgres = &Dialect{
		Name:         "postgres",
		SupportNamed: true,
		NamedPrefix:  "$",
		PlaceHolder:  "$",
		DateFormat:   "'2006-01-02 15:04:05'",
		SQLNameFunc:  MakeNameFunc("\"", "\""),
		NameFunc:     utils.LowerCase,
	}
)

func MakeNameFunc(prefix, suffix string) func(any) string {
	return func(name any) string {
		return QuotedName(name, prefix, suffix)
	}
}

func QuotedName(name any, prefix, suffix string) string {
	col := ""
	switch n := name.(type) {
	case string:
		col = prefix + n + suffix
	case fmt.Stringer:
		col = prefix + n.String() + suffix
	default:
		col = fmt.Sprintf("%s%v%s", prefix, n, suffix)
	}
	return col
}
