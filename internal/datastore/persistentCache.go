package datastore

import (
	"errors"
	"fmt"
	"log"

	bolt "go.etcd.io/bbolt"
)

var kvs *bolt.DB

func init() {
	// initializeKvs()
}

func initializeKvs() error {
	db, err := bolt.Open("kv.db", 0600, nil)
	if err == nil {
		kvs = db
		return nil
	} else {
		return errors.New("failed to initialize key value store")
	}
}

type persistentCache struct {
}

var persistentCacheInstnace Cache

func InitializePersistentCache() Cache {
	//TODO add mutex check while init
	if persistentCacheInstnace == nil {
		persistentCacheInstnace = &persistentCache{}
		if kvs == nil {
			if error := initializeKvs(); error != nil {
				log.Fatal(error)
			}
		}
	}
	return persistentCacheInstnace
}

func GetPersistentCache() Cache {
	if persistentCacheInstnace == nil {
		return InitializePersistentCache()
	}
	return persistentCacheInstnace
}

func (*persistentCache) Get(bucket string, key string) (any, error) {
	val, err := get(bucket, key)
	if err != nil {
		log.Println(err)
	}
	return val, err
}

func (*persistentCache) Put(bucket string, key string, value any) error {
	var stringerVal fmt.Stringer
	var stringerConvOk bool
	if stringerVal, stringerConvOk = value.(fmt.Stringer); !stringerConvOk {
		return errors.New("persistentCache: stringer not implemented, cant serialize value to string")
	}
	err := put(bucket, key, stringerVal.String())
	if err != nil {
		log.Println(err)
	}
	return err
}

func get(bucketName string, key string) (string, error) {

	var value string
	err := kvs.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))

		if bucket != nil {
			if byteVal := bucket.Get([]byte(key)); byteVal != nil {
				value = string(byteVal)
			}
		}

		return nil
	})

	return value, err
}

func put(bucketName string, key string, value string) error {

	err := kvs.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err == nil {
			err = bucket.Put([]byte(key), []byte(value))
		}
		return err
	})

	return err
}

func (*persistentCache) Delete(key string) {
	//TODO
}

func (*persistentCache) CleanUp() {
	if kvs != nil {
		kvs.Close()
		kvs = nil
		persistentCacheInstnace = nil
	}
}
