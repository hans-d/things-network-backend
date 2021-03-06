// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"fmt"

	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
	dbutil "github.com/TheThingsNetwork/ttn/utils/storage"
	"github.com/brocaar/lorawan"
)

type DevStorage interface {
	Lookup(appEUI lorawan.EUI64, devEUI lorawan.EUI64) (devEntry, error)
	Store(r HRegistration) error
}

type devEntry struct {
	Recipient []byte
	AppSKey   lorawan.AES128Key
	NwkSKey   lorawan.AES128Key
}

type appEntry struct {
	AppKey lorawan.AES128Key
}

type devStorage struct {
	db   dbutil.Interface
	Name string
}

// NewDevStorage creates a new Device Storage for handler
func NewDevStorage(name string) (DevStorage, error) {
	itf, err := dbutil.New(name)
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	return devStorage{db: itf, Name: "entry"}, nil
}

// Lookup implements the handler.DevStorage interface
func (s devStorage) Lookup(appEUI lorawan.EUI64, devEUI lorawan.EUI64) (devEntry, error) {
	itf, err := s.db.Lookup(fmt.Sprintf("%x.%x", appEUI[:], devEUI[:]), []byte(s.Name), &devEntry{})
	if err != nil {
		return devEntry{}, errors.New(errors.Operational, err)
	}
	entries, ok := itf.([]devEntry)
	if !ok || len(entries) != 1 {
		return devEntry{}, errors.New(errors.Structural, "Invalid stored entry")
	}
	return entries[0], nil
}

// Store implements the handler.DevStorage interface
func (s devStorage) Store(reg HRegistration) error {
	appEUI := reg.AppEUI()
	devEUI := reg.DevEUI()
	data, err := reg.Recipient().MarshalBinary()
	if err != nil {
		return errors.New(errors.Structural, "Cannot marshal recipient")
	}

	e := []dbutil.Entry{
		&devEntry{
			Recipient: data,
			AppSKey:   reg.AppSKey(),
			NwkSKey:   reg.NwkSKey(),
		},
	}
	return s.db.Store(fmt.Sprintf("%x.%x", appEUI[:], devEUI[:]), []byte(s.Name), e)
}

// Close implements the handler.DevStorage interface
func (s devStorage) Close() error {
	return s.db.Close()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e devEntry) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(e.Recipient)
	rw.Write(e.AppSKey)
	rw.Write(e.NwkSKey)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *devEntry) UnmarshalBinary(data []byte) error {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) { e.Recipient = data })
	rw.Read(func(data []byte) { copy(e.AppSKey[:], data) })
	rw.Read(func(data []byte) { copy(e.NwkSKey[:], data) })
	return rw.Err()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e appEntry) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(e.AppKey)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *appEntry) UnmarshalBinary(data []byte) error {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) { copy(e.AppKey[:], data) })
	return rw.Err()
}
