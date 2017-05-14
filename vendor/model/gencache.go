package model

import (
	"encoding/json"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

type avinfo struct {
	AID   int      `json:"vid"`
	Tags  []string `json:"tags"`
	Title string   `json:"vname"`
}

var (
	//Bucketname set package-global bucket name
	Bucketname = "avs"
	//DefaultCachePath defines a package global path
	DefaultCachePath = ".avcache"
)

//CreateBucket  create the bucket
func CreateBucket(bucketName string) error {
	db, err := bolt.Open(DefaultCachePath, 0600, nil)
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
	db, err := bolt.Open(DefaultCachePath, 0600, nil)
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

//InsertDB insert data
func InsertDB(bname string, key string, value string) error {
	db, err := bolt.Open(DefaultCachePath, 0666, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bname))

		b.Put([]byte(key), []byte(value))

		return nil
	})

	if err != nil {
		return err
	}
	return nil

}

//ChangeType change key of aid to tag
func ChangeType(dbpath string, bname string) (string, error) {
	db, err := bolt.Open(dbpath, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return "", err
	}
	defer db.Close()
	tagsmap := make(map[string][]string)
	//do a db select
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("avs"))
		if b == nil {
			b, err = tx.CreateBucketIfNotExists([]byte(Bucketname))
			if err != nil {
				panic(err)
			}
		}
		c := b.Cursor()
		if c != nil {
			//iterator
			for k, v := c.First(); k != nil; k, v = c.Next() {
				info := new(avinfo)

				err = json.Unmarshal(v, info)
				if err != nil {
					log.Fatal(err)
					return err
				}

				for _, tag := range info.Tags {
					tagsmap[tag] = append(tagsmap[tag], string(v))
				}
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}
	tagsjson, err := json.Marshal(tagsmap)
	if err != nil {
		return "", err
	}
	return string(tagsjson), nil
}

//GetFirstVID always get first vid
func GetFirstVID(dbpath string) (map[string]string, error) {
	db, err := bolt.Open(dbpath, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, err
	}
	defer db.Close()
	firstInfo := make(map[string]string)
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(Bucketname))
		if b == nil {
			panic("bucket not find")
		}
		c := b.Cursor()
		k, v := c.Last()
		firstInfo[string(k)] = string(v)
		return nil
	})
	return firstInfo, nil
}
