/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package builtin

import (
	"embed"
	_ "embed"
)

var (
	//go:embed builtin/*.sql
	Builtin embed.FS
)
