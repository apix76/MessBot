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
		_, err := tx.CreateBucketIfNotExists([]byte("Posts"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte("RefuseModer"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	db := Db{b}
	return db, err
}

func (db *Db) AddPost(idMess int, mess Message.MessageInfo) error {
	err := db.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Posts"))
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

func (db *Db) GetPost(idMess int) (Message.MessageInfo, error) {
	mess := Message.MessageInfo{}
	err := db.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Posts"))
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

func (db *Db) DeletePost(idMess int) error {
	err := db.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Posts"))
		idMessByte, err := json.Marshal(idMess)
		if err != nil {
			return err
		}
		err = bucket.Delete(idMessByte)
		return err
	})
	return err
}

func (db *Db) AddRefuseModer(ModerID int64, PostID int) error {
	err := db.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("RefuseModer"))
		ModerIDByte, err := json.Marshal(ModerID)
		if err != nil {
			return err
		}
		PostIDByte, err := json.Marshal(PostID)
		if err != nil {
			return err
		}
		err = bucket.Put(ModerIDByte, PostIDByte)
		return err
	})
	return err
}

// return nil, if value not exest
func (db *Db) GetRefuseModer(ModerID int64) (int, bool, error) {
	var PostID int
	exist := true
	err := db.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("RefuseModer"))
		ModerIDByte, err := json.Marshal(ModerID)
		if err != nil {
			return err
		}
		PostIDByte := bucket.Get(ModerIDByte)
		if PostIDByte == nil {
			exist = false
		}
		err = json.Unmarshal(PostIDByte, &PostID)
		return err
	})
	return PostID, exist, err
}
func (db *Db) DeleteRefuseModer(ModerID int64) error {
	err := db.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("RefuseModer"))
		ModerIDByte, err := json.Marshal(ModerID)
		if err != nil {
			return err
		}
		err = bucket.Delete(ModerIDByte)
		return err
	})
	return err
}
