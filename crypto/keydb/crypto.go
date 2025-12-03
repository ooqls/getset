package keydb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ooqls/getset/crypto/keys"
)

type CryptoDatabase interface {
	GetKeyPair(hashpw string) ([]byte, []byte, error)
	InsertKeyPair(hashedpw string, privateKey, publicKey []byte) error
	SetSystemKey(systemKey keys.Key) error
	GetSystemKey() keys.Key
	IsSystemKey(key keys.Key) (bool, error)
}

type SQLCryptoDatabase struct {
	db        *sql.DB
	systemKey keys.X509
}

func New(db *sql.DB, systemKey keys.X509) *SQLCryptoDatabase {
	return &SQLCryptoDatabase{
		db:        db,
		systemKey: systemKey,
	}
}

func LoadExisting(path string, key keys.X509) (*SQLCryptoDatabase, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	cdb := &SQLCryptoDatabase{
		db:        db,
		systemKey: key,
	}

	isSystemK, err := cdb.IsSystemKey(&key)
	if err != nil {
		return nil, fmt.Errorf("the given key is not the correct system key: %v", err)
	}

	if !isSystemK {
		return nil, ErrIncorrectSystemKey
	}

	return cdb, nil
}

func (c *SQLCryptoDatabase) SetSystemKey(systemKey keys.Key) error {
	_, pubKeyB := systemKey.PublicKey()

	row := c.db.QueryRow("SELECT public_key FROM system_meta")
	if row.Err() != nil && row.Err() != sql.ErrNoRows {
		return fmt.Errorf("failed to get existing cert signature: %v", row.Err())
	}

	_, err := c.db.Exec("INSERT INTO system_meta (public_key) VALUES (?)", pubKeyB)
	return err
}

func (c *SQLCryptoDatabase) GetSystemKey() keys.Key {
	return &c.systemKey
}

func (c *SQLCryptoDatabase) createTable(ctx context.Context) error {
	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	_, err = tx.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS key_pairs (hashed_pw varchar(128), private_key BLOB, public_key BLOB)")
	if err != nil {
		return fmt.Errorf("failed to create key_pairs table: %v", err)
	}

	_, err = tx.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS system_meta (public_key BLOB)")
	if err != nil {
		return fmt.Errorf("failed to create system_meta table: %v", err)
	}

	return tx.Commit()
}

func (c *SQLCryptoDatabase) GetKeyPair(hashpw string) ([]byte, []byte, error) {
	var hashedPw, encPrivateKey, publicKey []byte
	err := c.db.QueryRow("SELECT hashed_pw, private_key, public_key FROM key_pairs WHERE hashed_pw == ?", hashpw).Scan(&hashedPw, &encPrivateKey, &publicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get key pair: %v", err)
	}

	decPrivateKey, err := c.systemKey.Decrypt(encPrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt private key: %v", err)
	}

	return decPrivateKey, publicKey, err
}

func (c *SQLCryptoDatabase) InsertKeyPair(hashedpw string, privateKey, publicKey []byte) error {
	encPrivKey, err := c.systemKey.Encrypt(privateKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt private key: %v", err)
	}

	_, err = c.db.Exec("INSERT INTO key_pairs (hashed_pw, private_key, public_key) VALUES (?, ?, ?)", hashedpw, encPrivKey, publicKey)
	return err
}

func (c *SQLCryptoDatabase) IsSystemKey(key keys.Key) (bool, error) {
	var existingPublicKey []byte
	err := c.db.QueryRow("SELECT public_key FROM system_meta").Scan(&existingPublicKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, ErrDBNotInitialized
		}

		return false, fmt.Errorf("failed to get cert signature: %v", err)
	}

	_, pubKeyB := key.PublicKey()
	if string(existingPublicKey) != string(pubKeyB) {
		return false, nil
	}

	return true, nil
}

func (c *SQLCryptoDatabase) SystemKey() keys.Key {
	return &c.systemKey
}
