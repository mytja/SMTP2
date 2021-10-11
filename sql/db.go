package sql

import "github.com/jmoiron/sqlx"

var (
	DB    SQL
	DBERR error
)

func (db *sqlImpl) Init() {
	db.db.MustExec(schema)
}

func (db *sqlImpl) GetDB() *sqlx.DB {
	return db.db
}

func (db *sqlImpl) GenerateNewTransaction() {
	db.tx = db.db.MustBegin()
}

func (db *sqlImpl) Commit() error {
	err := db.tx.Commit()
	db.GenerateNewTransaction()
	return err
}
