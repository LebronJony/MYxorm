package session

import (
	"MXorm/clause"
	"errors"
	"reflect"
)

// Insert 需要将已经存在的对象的每一个字段的值平铺开来
func (s *Session) Insert(values ...interface{}) (int64, error) {
	// 字段value值的集合
	recordValues := make([]interface{}, 0)
	for _, value := range values {
		s.CallMethod(BeforeInsert, value)
		// 数据库表
		table := s.Model(value).RefTable()
		// 生成INSERT语句 设置该表的字段
		s.clause.Set(clause.INSERT, table.Name, table.FieldNames)
		// 添加字段值到集合
		recordValues = append(recordValues, table.RecordValues(value))
	}

	// 生成VALUES语句
	s.clause.Set(clause.VALUES, recordValues...)
	// 拼接语句
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	// 执行语句
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	s.CallMethod(AfterInsert, nil)
	return result.RowsAffected()
}

// Find 则是需要根据平铺开的字段的值构造出对象
func (s *Session) Find(values interface{}) error {
	s.CallMethod(BeforeQuery, nil)
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	// 获取切片的单个元素的类型 destType
	destType := destSlice.Type().Elem()
	// 使用 reflect.New() 方法创建一个 destType 的实例，作为 Model() 的入参，映射出表结构 RefTable()
	table := s.Model(reflect.New(destType).Elem().Interface()).RefTable()

	// 根据表结构，使用 clause 构造出 SELECT 语句，查询到所有符合条件的记录 rows
	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)
	sql, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	rows, err := s.Raw(sql, vars...).QueryRows()
	if err != nil {
		return err
	}

	// 遍历每一行记录
	for rows.Next() {
		// 利用反射创建 destType 的实例 dest
		dest := reflect.New(destType).Elem()
		var values []interface{}
		// 将 dest 的所有字段平铺开，构造切片 values
		for _, name := range table.FieldNames {
			values = append(values, dest.FieldByName(name).Addr().Interface())
		}
		// 调用 rows.Scan() 将该行记录每一列的值依次赋值给 values 中的每一个字段
		if err := rows.Scan(values...); err != nil {
			return err
		}

		// 钩子可以操作每一行记录
		s.CallMethod(AfterQuery, dest.Addr().Interface())

		// 将 dest 添加到切片 destSlice 中。循环直到所有的记录都添加到切片 destSlice 中
		destSlice.Set(reflect.Append(destSlice, dest))

	}

	return rows.Close()
}

// Update 接受 2 种入参，平铺开来的键值对和 map 类型的键值对
// 因为 generator 接受的参数是 map 类型的键值对，
// 因此 Update 方法会动态地判断传入参数的类型，如果是不是 map 类型，则会自动转换
func (s *Session) Update(kv ...interface{}) (int64, error) {
	s.CallMethod(BeforeUpdate, nil)
	m, ok := kv[0].(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
		for i := 0; i < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1]
		}
	}

	s.clause.Set(clause.UPDATE, s.RefTable().Name, m)
	sql, vars := s.clause.Build(clause.UPDATE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}

	s.CallMethod(AfterUpdate, nil)
	return result.RowsAffected()

}

func (s *Session) Delete() (int64, error) {
	s.CallMethod(BeforeDelete, nil)
	s.clause.Set(clause.DELETE, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.DELETE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	s.CallMethod(AfterDelete, nil)
	return result.RowsAffected()
}

func (s *Session) Count() (int64, error) {
	s.clause.Set(clause.COUNT, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.COUNT, clause.WHERE)
	row := s.Raw(sql, vars...).QueryRow()
	var tmp int64
	if err := row.Scan(&tmp); err != nil {
		return 0, err
	}
	return tmp, nil
}

func (s *Session) Limit(num int) *Session {
	s.clause.Set(clause.LIMIT, num)
	return s
}

func (s *Session) Where(desc string, args ...interface{}) *Session {
	var vars []interface{}
	s.clause.Set(clause.WHERE, append(append(vars, desc), args...)...)
	return s
}

func (s *Session) OrderBy(desc string) *Session {
	s.clause.Set(clause.ORDERBY, desc)
	return s
}

// First 我们期望 SQL 语句只返回一条记录，
//比如根据某个童鞋的学号查询他的信息，返回结果有且只有一条
func (s *Session) First(value interface{}) error {
	// 根据传入的类型，利用反射构造切片
	dest := reflect.Indirect(reflect.ValueOf(value))
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	// 调用 Limit(1) 限制返回的行数，调用 Find 方法获取到查询结果
	if err := s.Limit(1).Find(destSlice.Addr().Interface()); err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return errors.New("NOT FOUND")
	}
	dest.Set(destSlice.Index(0))
	return nil
}
