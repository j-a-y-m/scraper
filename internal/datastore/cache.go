package datastore

type Cache interface {
	Get(bucket, key string) (any, error)

	Put(bucket, key string, value any) error

	Delete(key string)

	CleanUp()
}
