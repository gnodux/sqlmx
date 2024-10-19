/*
 * Copyright (c) 2023.
 * all right reserved by gnodux<gnodux@gmail.com>
 */

package dialect

type Dialect struct {
	//驱动名称（mysql/mssql）等
	Name string
	//是否使用命名参数
	SupportNamed bool
	//命名参数前缀
	NamedPrefix string
	//参数占位符
	PlaceHolder string
	//SQLNameFunc SQL名称转换函数
	SQLNameFunc func(any) string
	//NameFunc 字段名称转换函数
	NameFunc func(string) string
	//DateFormat 日期格式化
	DateFormat string
	//Keywords 关键字映射
	Keywords map[string]string
}

func (d *Dialect) Keyword(name string) string {
	if d.Keywords == nil {
		return name
	}
	if k, ok := d.Keywords[name]; ok {
		return k
	}
	return name
}
func (d *Dialect) KeywordWith(prefix string, kw string, suffix string) string {

	return prefix + d.Keyword(kw) + suffix
}
func (d *Dialect) KeywordWithSpace(kw string) string {
	return d.KeywordWith(" ", kw, " ")
}
