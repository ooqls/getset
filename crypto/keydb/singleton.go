package keydb

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/ooqls/getset/crypto/keys"
)

var cdb CryptoDatabase = nil

func Init(sqllitePath string, systemKey keys.X509) error {
	if _, err := os.Stat(sqllitePath); os.IsNotExist(err) {
		_, err := os.Create(sqllitePath)
		if err != nil {
			return err
		}
	}

	db, err := sql.Open("sqlite3", sqllitePath)
	if err != nil {
		panic("failed to open sqllite3 database: " + err.Error())
	}

	cryptoDb := New(db, systemKey)
	err = cryptoDb.createTable(context.Background())
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	err = cryptoDb.SetSystemKey(&systemKey)
	if err != nil {
		return fmt.Errorf("failed to set system key: %v", err)
	}

	cdb = cryptoDb

	return nil
}

func GetCryptoDB() CryptoDatabase {
	if cdb == nil {
		panic("database not initialized")
	}

	return cdb
}
