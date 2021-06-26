package dialect

import (
	"fmt"
	"reflect"
	"time"
)

/*
	dialect对SQLite的支持
*/

type sqlite3 struct {
}

// 将空值nil转换为*sqlite3类型，再转换为Dialect接口
// 如果转换失败，说说明sqlite3没有实现Dialect接口的所有方法
var _ Dialect = (*sqlite3)(nil)


// init 函数为自动调用
// 包在第一次加载时，会将 sqlite3 的 dialect 自动注册到全局
func init() {
	RegisterDialect("sqlite3", &sqlite3{})
}

// DataTypeOf 将 Go 语言的类型映射为 SQLite 的数据类型
func (s *sqlite3) DataTypeOf(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uintptr:
		return "integer"
	case reflect.Int64, reflect.Uint64:
		return "bigint"
	case reflect.Float32, reflect.Float64:
		return "real"
	case reflect.String:
		return "text"
	case reflect.Array, reflect.Slice:
		return "blob"
	case reflect.Struct:
		if _, ok := typ.Interface().(time.Time); ok {
			return "datetime"
		}

	}
	panic(fmt.Sprintf("invalid sql type %s (%s)", typ.Type().Name(), typ.Kind()))
}

func (s *sqlite3) TableExistSQL(tableName string) (string, []interface{}) {
	args := []interface{}{tableName}
	return "SELECT name FROM sqlite_master WHERE type='table' and name = ?", args
}
