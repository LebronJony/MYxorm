package MXorm

import (
	"MXorm/dialect"
	"MXorm/log"
	"MXorm/session"
	"database/sql"
	"fmt"
	"strings"
)

/*
	Engine 是 GeeORM 与用户交互的入口
	交互前的准备工作（比如连接/测试数据库），交互后的收尾工作（关闭连接）等由 Engine 来负责
*/

type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

// NewEngine ,driver为数据库名
func NewEngine(driver, source string) (e *Engine, err error) {

	// 连接数据库
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return
	}

	// 发送一个ping包以确保数据库连接没有断开
	if err = db.Ping(); err != nil {
		log.Error(err)
		return
	}

	dial, ok := dialect.GetDialect(driver)
	if !ok {
		log.Errorf("dialect %s Not Found", driver)
		return
	}

	e = &Engine{db: db, dialect: dial}
	log.Info("Connect database success")
	return

}

// Close 关闭数据库连接
func (engine *Engine) Close() {
	if err := engine.db.Close(); err != nil {
		log.Error("Fail to close database")
	}
	log.Info("Close database success")
}

// NewSession 通过 Engine 实例创建会话，进而与数据库进行交互
func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db, engine.dialect)
}

type TxFunc func(*session.Session) (interface{}, error)

// Transaction 用户只需要将所有的操作放到一个回调函数中，
// 作为入参传递给 engine.Transaction()，发生任何错误，自动回滚，如果没有错误发生，则提交
func (engine *Engine) Transaction(f TxFunc) (result interface{}, err error) {
	s := engine.NewSession()
	if err := s.Begin(); err != nil {
		return nil, err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = s.Rollback()
			panic(p)
		} else if err != nil {
			_ = s.Rollback()
		} else {
			err = s.Commit()
		}
	}()

	return f(s)
}

// difference 用来计算前后两个字段切片的差集。
// 新表 - 旧表 = 新增字段，旧表 - 新表 = 删除字段
func difference(a []string, b []string) (diff []string) {
	mapB := make(map[string]bool)
	for _, v := range b {
		mapB[v] = true
	}
	for _, v := range a {
		if _, ok := mapB[v]; !ok {
			diff = append(diff, v)
		}
	}
	return
}

// Migrate 结构体(struct)变更时，数据库表的字段(field)自动迁移(migrate)
func (engine *Engine) Migrate(value interface{}) error {
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		if !s.Model(value).HasTable() {
			log.Infof("table %s doesn't exist", s.RefTable().Name)
			return nil, s.CreateTable()
		}
		table := s.RefTable()
		rows, _ := s.Raw(fmt.Sprintf("SELECT * FROM %s LIMIT 1", table.Name)).QueryRows()
		columns, _ := rows.Columns()
		addCols := difference(table.FieldNames, columns)
		delCols := difference(columns, table.FieldNames)
		log.Infof("added cols %v,deleted cols %v", addCols, delCols)

		for _, col := range addCols {
			f := table.GetField(col)
			sqlStr := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", table.Name, f.Name, f.Type)
			if _, err = s.Raw(sqlStr).Exec(); err != nil {
				return
			}
		}

		if len(delCols) == 0 {
			return
		}

		tmp := "tmp_" + table.Name
		fieldStr := strings.Join(table.FieldNames, ", ")
		s.Raw(fmt.Sprintf("CREATE TABLE %s AS SELECT %s from %s;", tmp, fieldStr, table.Name))
		s.Raw(fmt.Sprintf("DROP TABLE %s;", table.Name))
		s.Raw(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", tmp, table.Name))
		_, err = s.Exec()
		return
	})
	return err
}
