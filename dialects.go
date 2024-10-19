/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlmx

import "github.com/gnodux/sqlmx/dialect"

var (
	DefaultDialect = dialect.MySQL
	MySQL          = dialect.MySQL
	SQLServer      = dialect.SQLServer
	Postgres       = dialect.Postgres
	Dialects       = map[string]*dialect.Dialect{
		"mysql":    MySQL,
		"mssql":    SQLServer,
		"postgres": Postgres,
	}
)
