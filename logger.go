/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package sqlmx

import "github.com/sirupsen/logrus"

type Logger interface {
	Trace(...any)
	Tracef(string, ...any)
	Debug(...any)
	Debugf(string, ...any)
	Info(...any)
	Infof(string, ...any)

	Warn(...any)
	Warnf(string, ...any)
	Error(...any)
	Errorf(string, ...any)
}

var (
	log Logger = logrus.StandardLogger()
)

func SetLogger(l Logger) {
	log = l
}
