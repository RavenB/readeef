package base

import (
	"github.com/jmoiron/sqlx"
	"github.com/urandom/readeef/content/sql/db"
)

type Helper struct {
	sql map[string]string
}

func NewHelper() Helper {
	return Helper{sql: make(map[string]string)}
}

func (h Helper) SQL(name string) string {
	if v, ok := h.sql[name]; ok {
		return v
	}
	return sql[name]
}

func (h Helper) Set(name, stmt string) {
	h.sql[name] = stmt
}

func (h Helper) Upgrade(db *db.DB, old, new int) error {
	return nil
}

func (h Helper) CreateWithId(tx *sqlx.Tx, name string, args ...interface{}) (int64, error) {
	var id int64

	sql := h.SQL(name)
	if sql == "" {
		panic("No statement registered under " + name)
	}

	stmt, err := tx.Preparex(sql)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(args...)
	if err != nil {
		return 0, err
	}

	id, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

var (
	sql = make(map[string]string)
)
