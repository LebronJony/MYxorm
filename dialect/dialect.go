package dialect

import "reflect"

/*
	实现 ORM 映射的第一步，需要思考如何将 Go 语言的类型映射为数据库中的类型。
	同时，不同数据库支持的数据类型也是有差异的，即使功能相同，在 SQL 语句的表达上也可能有差异
	ORM 框架往往需要兼容多种数据库，因此我们需要将差异的这一部分提取出来，
	每一种数据库分别实现，实现最大程度的复用和解耦。这部分代码称之为 dialect。
*/

// key 数据库名 ,value Dialect接口实例
var dialectsMap = map[string]Dialect{}

type Dialect interface {
	// DataTypeOf 将 Go 语言的类型转换为该数据库的数据类型
	DataTypeOf(typ reflect.Value) string

	// TableExistSQL 返回某个表是否存在的 SQL 语句，参数是表名(table)
	TableExistSQL(tableName string) (string, []interface{})
}

// RegisterDialect 注册dialect实例
// 如果新增加对某个数据库的支持，那么调用 RegisterDialect 即可注册到全局
func RegisterDialect(name string, dialect Dialect) {
	dialectsMap[name] = dialect
}

func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectsMap[name]
	return
}
