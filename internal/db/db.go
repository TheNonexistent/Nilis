package db

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

const defaultBucketName = "nilis"

type Database struct {
	database *bolt.DB
}

func NewDatabase(path string) (*Database, error) {
	localdb, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	database := &Database{
		database: localdb,
	}

	if err := database.createDefaultBucket(); err != nil {
		database.Close()
		return nil, fmt.Errorf("failed creating default bucket: %w", err)
	}

	return database, nil
}

func (db *Database) createDefaultBucket() error {
	return db.database.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(defaultBucketName))
		return err
	})
}

func (db *Database) SetKey(key string, value []byte) error {
	return db.database.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(defaultBucketName))
		return b.Put([]byte(key), value)
	})
}

func (db *Database) GetKey(key string) ([]byte, error) {
	var value []byte

	err := db.database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(defaultBucketName))
		value = b.Get([]byte(key))
		return nil
	})

	if err != nil {
		return nil, err
	}

	return value, nil
}

func (db *Database) DeleteKey(key string) error {
	return db.database.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(defaultBucketName))
		return b.Delete([]byte(key))
	})
}

func (db *Database) Close() error {
	return db.database.Close()
}
