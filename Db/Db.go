package Db

import (
	"MessBot/Message"
	"encoding/json"
	"fmt"
	bolt "go.etcd.io/bbolt"
)

type Db struct {
	db *bolt.DB
}

func NewDB() (Db, error) {
	b, err := bolt.Open("Db.db", 0600, nil)
	if err != nil {
		return Db{}, err
	}
	b.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Bucket"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	db := Db{b}
	return db, err
}

func (db *Db) Add(idMess int, mess Message.TempMess) error {
	err := db.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Bucket"))
		idMessByte, err := json.Marshal(idMess)
		if err != nil {
			return err
		}
		messByte, err := json.Marshal(mess)
		if err != nil {
			return err
		}
		err = bucket.Put(idMessByte, messByte)
		return err
	})
	return err
}

func (db *Db) Get(idMess int) (Message.TempMess, error) {
	var mess Message.TempMess
	err := db.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Bucket"))
		idMessByte, err := json.Marshal(idMess)
		if err != nil {
			return err
		}
		MessByte := bucket.Get(idMessByte)
		err = json.Unmarshal(MessByte, &mess)
		return err
	})
	return mess, err
}

func (db *Db) Delete(idMess int) error {
	err := db.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Bucket"))
		idMessByte, err := json.Marshal(idMess)
		if err != nil {
			return err
		}
		err = bucket.Delete(idMessByte)
		return err
	})
	return err
}
