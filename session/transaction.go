package session

/*
	封装事务的 Begin、Commit 和 Rollback 三个接口
	封装的另一个目的是统一打印日志，方便定位问题
 */

import "MXorm/log"

// Begin 封装事务的Begin
func (s *Session) Begin() (err error) {
	log.Info("transaction begin")
	// 调用 s.db.Begin() 得到 *sql.Tx 对象，赋值给 s.tx
	if s.tx, err = s.db.Begin(); err != nil {
		log.Error(err)
		return
	}
	return
}

// Commit 封装事务的Commit
func (s *Session) Commit() (err error) {
	log.Info("transcation commit")
	if err = s.tx.Commit(); err != nil {
		log.Error(err)
	}
	return
}

// Rollback 封装事务的Rollback
func (s *Session) Rollback() (err error) {
	log.Info("transaction rollback")
	if err = s.tx.Rollback(); err != nil {
		log.Error(err)
	}
	return
}
