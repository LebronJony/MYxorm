package schema

import (
	"MXorm/dialect"
	"go/ast"
	"reflect"
)

/*
	对象(object)和表(table)的转换。
	给定一个任意的对象，转换为关系型数据库中的表结构。
*/

// Field 字段代表数据库中的一列
type Field struct {
	// 字段名
	Name string
	// 类型
	Type string
	// 约束条件
	Tag string
}

// Schema 数据库表
type Schema struct {
	// 被映射的对象
	Model interface{}
	// 表名
	Name string
	// 字段指针数组
	Fields []*Field
	// 字段名数组 包含所有字段名（列名）
	FieldNames []string
	// 记录字段名和Field的映射关系,方便之后直接使用，无需遍历Fields
	fieldMap map[string]*Field
}

// Parse 将任意的对象解析为 Schema 实例
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	// TypeOf() 和 ValueOf() 是 reflect 包最为基本也是最重要的 2 个方法，
	// 分别用来返回入参的类型和值。因为设计的入参是一个对象的指针，
	// 因此需要 reflect.Indirect() 获取指针指向的实例。
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()

	schema := &Schema{
		Model: dest,
		// modelType.Name() 获取到结构体的名称作为表名
		Name:     modelType.Name(),
		fieldMap: make(map[string]*Field),
	}

	// NumField() 获取实例的字段的个数，即结构体的字段数
	// 然后通过下标获取到特定字段 p := modelType.Field(i)
	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i)
		// if该字段不是匿名字段且字段名是大写字母开头
		if !p.Anonymous && ast.IsExported(p.Name) {
			field := &Field{
				Name: p.Name,
				// 将字段类型转换为数据库的字段类型
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),
			}
			if v, ok := p.Tag.Lookup("geeorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.fieldMap[p.Name] = field
		}
	}
	return schema
}


func (schema *Schema) GetField(name string) *Field {
	return schema.fieldMap[name]
}

// RecordValues 根据数据库中列的顺序，从对象中找到对应的值，按顺序平铺
// u1 := &User{Name: "Tom", Age: 18} 即将u1转换为("Tom", 18)格式
func (schema *Schema) RecordValues(dest interface{}) []interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []interface{}
	for _, field := range schema.Fields {
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues
}

