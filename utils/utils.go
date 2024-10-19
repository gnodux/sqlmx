/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package utils

import (
	"reflect"
	"unicode"
)

func Each[T any](lst []T, fn func(int, T) bool) {
	for idx, itm := range lst {
		if !fn(idx, itm) {
			break
		}
	}
}
func ToMap(v any, excludes ...string) map[string]any {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	vv := reflect.ValueOf(v)
	if vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
		if vv.Kind() == reflect.Ptr {
			vv = vv.Elem()
		}
	}
	result := map[string]any{}
	typ := reflect.TypeOf(vv.Interface())
	for idx := 0; idx < vv.NumField(); idx++ {
		f := vv.Field(idx)
		if Contains(excludes, func(exclude string) bool {
			return exclude == typ.Field(idx).Name || exclude == LowerCase(typ.Field(idx).Name)
		}) {
			continue
		}
		ft := typ.Field(idx)
		if ft.IsExported() && !f.IsZero() {
			if ft.Anonymous {
				for k, v := range ToMap(f.Interface()) {
					result[k] = v
				}
			} else {
				result[ft.Name] = f.Interface()
			}
		}
	}

	return result
}

func Search[T any](lst []T, fn func(T) bool) (result []T) {
	for _, itm := range lst {
		if fn(itm) {
			result = append(result, itm)
		}
	}
	return
}
func Contains[T any](lst []T, fn func(T) bool) bool {
	for _, itm := range lst {
		if fn(itm) {
			return true
		}
	}
	return false
}

func ValueOrZero(v any, typ reflect.Type) reflect.Value {
	if v == nil {
		return reflect.Zero(typ)
	} else {
		return reflect.ValueOf(v)
	}
}

func LowerCase(s string) string {
	if s == "" {
		return s
	}
	rawStr := []rune(s)
	var newStr []rune
	preupper := true
	for idx, r := range rawStr {
		if unicode.IsUpper(r) {
			if idx > 0 && rawStr[idx-1] != '.' && !preupper {
				newStr = append(newStr, '_', unicode.ToLower(r))
				preupper = true
			} else {
				newStr = append(newStr, unicode.ToLower(r))
			}
		} else {
			preupper = false
			newStr = append(newStr, r)
		}
	}
	return string(newStr)
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func BigCamelCase(s string) string {
	if s == "" {
		return s
	}
	rawStr := []rune(s)
	var newStr []rune
	toUpper := false
	for idx, r := range rawStr {
		if idx == 0 {
			newStr = append(newStr, unicode.ToUpper(r))
			continue
		}
		if r == '_' {
			toUpper = true
			continue
		}
		if r == '.' {
			newStr = append(newStr, r)
			toUpper = true
			continue
		}

		if toUpper {
			newStr = append(newStr, unicode.ToUpper(r))
			toUpper = false
		} else {
			newStr = append(newStr, r)
		}
	}
	return string(newStr)

}
func SmallCamelCase(s string) string {
	if s == "" {
		return s
	}
	rawStr := []rune(s)
	var newStr []rune
	toLower := false
	for idx, r := range rawStr {
		if idx == 0 {
			newStr = append(newStr, unicode.ToLower(r))
			continue
		}
		if r == '_' {
			toLower = true
			continue
		}
		if r == '.' {
			newStr = append(newStr, r)
			toLower = true
			continue
		}
		if toLower {
			newStr = append(newStr, unicode.ToLower(r))
			toLower = false
		} else {
			newStr = append(newStr, r)
		}
	}
	return string(newStr)
}

var (
	encodeRef = map[byte]byte{
		'\x00': '0',
		'\'':   '\'',
		'"':    '"',
		'\b':   'b',
		'\n':   'n',
		'\r':   'r',
		'\t':   't',
		26:     'Z', // ctl-Z
		'\\':   '\\',
	}
	EncodeMap  [256]byte
	DONTESCAPE byte = 255
)

// Escape only support utf-8
func Escape(sql string) string {
	dest := make([]byte, 0, 2*len(sql))

	for _, w := range []byte(sql) {
		if c := EncodeMap[w]; c == DONTESCAPE {
			dest = append(dest, w)
		} else {
			dest = append(dest, '\\', c)
		}
	}

	return string(dest)
}

func init() {
	for i := range EncodeMap {
		EncodeMap[i] = DONTESCAPE
	}
	for i := range EncodeMap {
		if to, ok := encodeRef[byte(i)]; ok {
			EncodeMap[byte(i)] = to
		}
	}
}
