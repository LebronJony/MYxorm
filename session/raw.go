package session

import (
	"MXorm/clause"
	"MXorm/dialect"
	"MXorm/log"
	"MXorm/schema"
	"database/sql"
	"strings"
)

/*
	用于实现与数据库的交互
*/

type Session struct {
	// 使用 sql.Open() 方法连接数据库成功之后返回的指针
	db      *sql.DB
	dialect dialect.Dialect

	// 如果要支持事务，需要更改为 sql.Tx 执行
	tx *sql.Tx

	refTable *schema.Schema
	clause   clause.Clause
	// 用来拼接 SQL 语句
	sql strings.Builder
	// SQL 语句中占位符的对应值
	sqlVars []interface{}
}

func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{
		db:      db,
		dialect: dialect,
	}
}

func (s *Session) Clear() {
	// 字符串置空
	s.sql.Reset()
	s.sqlVars = nil
	s.clause = clause.Clause{}
}

type CommonDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)


}

var _ CommonDB = (*sql.DB)(nil)
var _ CommonDB = (*sql.Tx)(nil)

// DB 当 tx 不为空时，则使用 tx 执行 SQL 语句，
// 否则使用 db 执行 SQL 语句。这样既兼容了原有的执行方式，又提供了对事务的支持。
func (s *Session) DB() CommonDB {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

// Raw 用户调用 Raw() 方法即可改变这两个变量的值
func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}

/*
	封装有 2 个目的，一是统一打印日志（包括 执行的SQL 语句和错误日志）。
	二是执行完成后，清空 (s *Session).sql 和 (s *Session).sqlVars 两个变量。
	这样 Session 可以复用，开启一次会话，可以执行多次 SQL。
*/

// Exec 封装Exec()执行sql语句
func (s *Session) Exec() (result sql.Result, err error) {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	// 执行sql语句
	if result, err = s.DB().Exec(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}

// QueryRow 封装QueryRow()执行一条单行查询
func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	return s.DB().QueryRow(s.sql.String(), s.sqlVars...)
}

// QueryRows 封装Query()执行一条多行查询
func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	if rows, err = s.DB().Query(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}
