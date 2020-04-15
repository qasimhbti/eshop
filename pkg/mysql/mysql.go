package mysql

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/pkg/errors"
)

// GetDB get DB connection
func GetDB(userName, password, host, dbname string) (*sql.DB, error) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s", userName, password, host, dbname)
	log.Println("con string :", connectionString)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, errors.WithMessage(err, "db connection")
	}
	log.Println("MySQL DB connected")

	err = db.Ping()
	if err != nil {
		return nil, errors.WithMessage(err, "db ping")
	}
	log.Println("MySQL DB Pinged Successfully")

	return db, nil
}
