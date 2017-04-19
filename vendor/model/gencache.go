package model

import (
	"log"
	"strconv"

	"github.com/boltdb/bolt"
)

//CachePath defines a package global path
const CachePath = "av.db"

//DB wrap bolt.db
type DB struct {
	bdb *bolt.DB
}

//OpenBolt return a boltdb instanse
func OpenBolt() (*DB, error) {
	db, err := bolt.Open(CachePath, 0600, nil)
	if err != nil {
		return nil, err
	}
	wrapdb := new(DB)
	wrapdb.bdb = db
	return wrapdb, nil
}

//CreateBucket  create the bucket
func (db *DB) CreateBucket(bucketName string) error {
	err := db.bdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})
	defer db.bdb.Close()
	return err
}

//PutAv2db insert av data ..
func (db *DB) PutAv2db(bname string, src []map[int]string) error {
	err := db.bdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bname))
		for _, v := range src {
			for kk, vv := range v {
				err := b.Put([]byte(strconv.Itoa(kk)), []byte(vv))
				log.Fatal(err)
			}
		}
		return nil
	})
	defer db.bdb.Close()
	if err != nil {
		return err
	}
	return nil

}
