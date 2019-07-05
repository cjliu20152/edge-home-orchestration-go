/*******************************************************************************
 * Copyright 2019 Samsung Electronics All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 *******************************************************************************/
package wrapper

import (
	"common/errors"

	"github.com/boltdb/bolt"
)

const (
	PATH = "/var/data/db/data.db"
	PORT = 0600
)

type (
	Database interface {
		Get(key []byte) ([]byte, error)
		Put(key []byte, value []byte) error
		List() (map[string]interface{}, error)
		Delete(key []byte) error
	}

	BoltDB struct {
		bucketname string
		boltdb     *bolt.DB
	}
)

func NewBoltDB(bucketname string) Database {
	return &BoltDB{bucketname: bucketname}
}

func (db *BoltDB) dbOpen() error {
	conn, err := bolt.Open(PATH, PORT, nil)
	if err != nil {
		return errors.DBConnectionError{Message: err.Error()}
	}
	db.boltdb = conn
	return nil
}

func (db *BoltDB) dbClose() {
	db.boltdb.Close()
}

func (db *BoltDB) Get(key []byte) ([]byte, error) {
	err := db.dbOpen()
	if err != nil {
		return nil, err
	}
	defer db.dbClose()

	var data []byte
	err = db.boltdb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(db.bucketname))
		if bucket == nil {
			return errors.NotFound{Message: string(key[:]) + " does not exist"}
		}

		v := bucket.Get(key)
		if len(v) == 0 {
			return errors.NotFound{Message: string(key[:]) + " does not exist"}
		}
		data = make([]byte, len(v))
		copy(data, v)
		return nil
	})
	return data, err
}

func (db *BoltDB) Put(key []byte, value []byte) error {
	err := db.dbOpen()
	if err != nil {
		return err
	}
	defer db.dbClose()

	return db.boltdb.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(db.bucketname))
		if err != nil {
			return errors.DBOperationError{Message: err.Error()}
		}
		return bucket.Put(key, value)
	})
}

func (db BoltDB) List() (map[string]interface{}, error) {
	err := db.dbOpen()
	if err != nil {
		return nil, err
	}
	defer db.dbClose()

	data := make(map[string]interface{})
	err = db.boltdb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(db.bucketname))
		if bucket != nil {
			c := bucket.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				data[string(k)] = string(v)
			}
		}
		return nil
	})

	return data, err
}

func (db *BoltDB) Delete(key []byte) error {
	err := db.dbOpen()
	if err != nil {
		return err
	}
	defer db.dbClose()

	return db.boltdb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(db.bucketname))
		if bucket == nil {
			return errors.NotFound{Message: string(key[:]) + " does not exist"}
		}

		v := bucket.Get(key)
		if len(v) == 0 {
			return errors.NotFound{Message: string(key[:]) + " does not exist"}
		}

		return bucket.Delete(key)
	})
}
