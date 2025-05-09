package datastore

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

var kvs *bolt.DB

func init() {
	// initializeKvs()
}

func initializeKvs() error {
	db, err := bolt.Open("kv.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err == nil {
		kvs = db
		return nil
	} else {
		return fmt.Errorf("failed to initialize key value store: %w", err)
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
		var persistentCache Cache
		persistentCache = sync.OnceValue(InitializePersistentCache)()
		return persistentCache
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
	var (
		marshaledValue []byte
		marshalErr     error
	)
	if marshaledValue, marshalErr = json.Marshal(value); marshalErr != nil {
		return marshalErr
	}
	err := put(bucket, key, marshaledValue)
	if err != nil {
		log.Println(err)
	}
	return err
}

func get(bucketName string, key string) ([]byte, error) {

	var byteVal []byte
	err := kvs.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))

		if bucket != nil {
			byteVal = bucket.Get([]byte(key))
		}

		return nil
	})

	return byteVal, err
}

func put(bucketName string, key string, value []byte) error {

	err := kvs.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err == nil {
			err = bucket.Put([]byte(key), value)
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
