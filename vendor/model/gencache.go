package model

import (
	"log"

	"github.com/boltdb/bolt"
)

//CachePath defines a package global path
const CachePath = ".avcache"

//CreateBucket  create the bucket
func CreateBucket(bucketName string) error {
	db, err := bolt.Open(CachePath, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})

	return err
}

//PutAv2db insert av data ..
func PutAv2db(bname string, src map[string]string) error {
	db, err := bolt.Open(".avcache", 0666, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bname))

		for kk, vv := range src {
			err = b.Put([]byte(kk), []byte(vv))
			if err != nil {
				panic(err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}
	return nil

}

//SelectDB Select the k-v
func SelectDB(bname string, key string) ([]byte, error) {
	var res []byte
	db, err := bolt.Open(CachePath, 0600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	res = []byte("")
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bname))
		res = b.Get([]byte(key))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
