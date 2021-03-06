// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
	"github.com/boltdb/bolt"
)

// Entry offers a friendly interface on which the storage will operate.
// Basically, a Entry is nothing more than a binary marshaller/unmarshaller.
type Entry interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// The storage Interface provides an abstraction on top of the bolt database to store and access
// data in a local in-memory database.
// All "table" or "bucket" in a database can be accessed by their name as a string where the dot
// "." is use as a level-separator to select or create nested buckets.
type Interface interface {
	Store(name string, key []byte, entries []Entry) error
	Replace(name string, key []byte, entries []Entry) error
	Lookup(name string, key []byte, shape Entry) (interface{}, error)
	Flush(name string, key []byte) error
	Reset(name string) error
	Close() error
}

type store struct {
	db *bolt.DB
}

// New creates a new storage instance ready-to-use
func New(name string) (Interface, error) {
	db, err := bolt.Open(name, 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}
	return store{db}, nil
}

func getBucket(tx *bolt.Tx, name string) (*bolt.Bucket, error) {
	path := strings.Split(name, ".")
	if len(path) < 1 {
		return nil, errors.New(errors.Structural, "Invalid bucket name")
	}
	var next interface {
		CreateBucketIfNotExists(b []byte) (*bolt.Bucket, error)
		Bucket(b []byte) *bolt.Bucket
	}

	var err error
	next = tx
	for _, bname := range path {
		next = next.Bucket([]byte(bname))
		if next == nil {
			if next, err = next.CreateBucketIfNotExists([]byte(bname)); err != nil {
				return nil, errors.New(errors.Operational, err)
			}
		}
	}
	return next.(*bolt.Bucket), nil
}

// Store put a new set of entries in the given bolt database. It adds the entries to an existing set
// or create a new set.
func (itf store) Store(bucketName string, key []byte, entries []Entry) error {
	var marshalled [][]byte

	for _, entry := range entries {
		m, err := entry.MarshalBinary()
		if err != nil {
			return errors.New(errors.Structural, err)
		}
		marshalled = append(marshalled, m)
	}

	err := itf.db.Update(func(tx *bolt.Tx) error {
		bucket, err := getBucket(tx, bucketName)
		if err != nil {
			return err
		}
		w := readwriter.New(bucket.Get(key))
		for _, m := range marshalled {
			w.Write(m)
		}
		data, err := w.Bytes()
		if err != nil {
			return errors.New(errors.Structural, err)
		}
		if err := bucket.Put(key, data); err != nil {
			return errors.New(errors.Operational, err)
		}
		return nil
	})

	return err
}

// Lookup retrieves a set of entry from a given bolt database.
//
// The shape is used as a template for retrieving and creating the data. All entries extracted from
// the database will be interpreted as instance of shape and the return result will be a slice of
// the same type of shape.
func (itf store) Lookup(bucketName string, key []byte, shape Entry) (interface{}, error) {
	// First, lookup the raw entries
	var rawEntry []byte
	err := itf.db.View(func(tx *bolt.Tx) error {
		bucket, err := getBucket(tx, bucketName)
		if err != nil {
			return err
		}
		rawEntry = bucket.Get(key)
		if rawEntry == nil {
			return errors.New(errors.Behavioural, fmt.Sprintf("Not found %+v", key))
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Then, interpret them as instance of 'shape'
	r := readwriter.New(rawEntry)
	entryType := reflect.TypeOf(shape).Elem()
	entries := reflect.MakeSlice(reflect.SliceOf(entryType), 0, 0)
	for {
		r.Read(func(data []byte) {
			entry := reflect.New(entryType).Interface()
			entry.(Entry).UnmarshalBinary(data)
			entries = reflect.Append(entries, reflect.ValueOf(entry).Elem())
		})
		if err = r.Err(); err != nil {
			if err == io.EOF {
				break
			}
			return nil, errors.New(errors.Operational, err)
		}
	}
	return entries.Interface(), nil
}

// Flush empties each entry of a bucket associated to a given device
func (itf store) Flush(bucketName string, key []byte) error {
	return itf.db.Update(func(tx *bolt.Tx) error {
		bucket, err := getBucket(tx, bucketName)
		if err != nil {
			return err
		}
		if err := bucket.Delete(key); err != nil {
			return errors.New(errors.Operational, err)
		}
		return nil
	})
}

// Replace stores entries in the database by replacing them by a new set
func (itf store) Replace(bucketName string, key []byte, entries []Entry) error {
	var marshalled [][]byte

	for _, entry := range entries {
		m, err := entry.MarshalBinary()
		if err != nil {
			return errors.New(errors.Structural, err)
		}
		marshalled = append(marshalled, m)
	}

	return itf.db.Update(func(tx *bolt.Tx) error {
		bucket, err := getBucket(tx, bucketName)
		if err != nil {
			return err
		}
		if err := bucket.Delete(key); err != nil {
			return errors.New(errors.Operational, err)
		}
		w := readwriter.New(bucket.Get(key))
		for _, m := range marshalled {
			w.Write(m)
		}
		data, err := w.Bytes()
		if err != nil {
			return errors.New(errors.Structural, err)
		}
		if err := bucket.Put(key, data); err != nil {
			return errors.New(errors.Operational, err)
		}
		return nil
	})
}

// Reset resets a given bucket from a given bolt database
func (itf store) Reset(bucketName string) error {
	return itf.db.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket([]byte(bucketName)); err != nil {
			return errors.New(errors.Operational, err)
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(bucketName)); err != nil {
			return errors.New(errors.Operational, err)
		}
		return nil
	})
}

// Close terminates the db connection
func (itf store) Close() error {
	if err := itf.db.Close(); err != nil {
		return errors.New(errors.Operational, err)
	}
	return nil
}
