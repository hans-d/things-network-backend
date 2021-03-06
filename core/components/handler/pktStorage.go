// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"fmt"

	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	dbutil "github.com/TheThingsNetwork/ttn/utils/storage"
	"github.com/brocaar/lorawan"
)

type PktStorage interface {
	Push(p HPacket) error
	Pull(appEUI lorawan.EUI64, devEUI lorawan.EUI64) (HPacket, error)
}

type pktStorage struct {
	db   dbutil.Interface
	Name string
}

type pktEntry struct {
	HPacket
}

// NewPktStorage creates a new PktStorage
func NewPktStorage(name string) (PktStorage, error) {
	itf, err := dbutil.New(name)
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}
	return pktStorage{db: itf, Name: "pktStorage"}, nil
}

func keyFromEUIs(appEUI lorawan.EUI64, devEUI lorawan.EUI64) []byte {
	return append(appEUI[:], devEUI[:]...)
}

// Push implements the PktStorage interface
func (s pktStorage) Push(p HPacket) error {
	err := s.db.Store(s.Name, keyFromEUIs(p.AppEUI(), p.DevEUI()), []dbutil.Entry{&pktEntry{p}})
	if err != nil {
		return errors.New(errors.Operational, err)
	}
	return nil
}

// Pull implements the PktStorage interface
func (s pktStorage) Pull(appEUI lorawan.EUI64, devEUI lorawan.EUI64) (HPacket, error) {
	key := keyFromEUIs(appEUI, devEUI)

	entries, err := s.db.Lookup(s.Name, key, &pktEntry{})
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	packets, ok := entries.([]*pktEntry)
	if !ok {
		return nil, errors.New(errors.Operational, "Unable to retrieve data from db")
	}

	// NOTE: one day, those entries will be more complicated, with a ttl.
	// Here's the place where we should check for that. Cheers.
	if len(packets) == 0 {
		return nil, errors.New(errors.Behavioural, fmt.Sprintf("Entry not found for %v", key))
	}

	pkt := packets[0]

	var newEntries []dbutil.Entry
	for _, p := range packets[1:] {
		newEntries = append(newEntries, p)
	}

	if err := s.db.Replace(s.Name, key, newEntries); err != nil {
		// TODO This is critical... we've just lost a packet
		return nil, errors.New(errors.Operational, "Unable to restore data in db")
	}

	return pkt.HPacket, nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e pktEntry) MarshalBinary() ([]byte, error) {
	return e.HPacket.MarshalBinary()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *pktEntry) UnmarshalBinary(data []byte) error {
	itf, err := UnmarshalPacket(data)
	if err != nil {
		return errors.New(errors.Structural, err)
	}
	packet, ok := itf.(HPacket)
	if !ok {
		return errors.New(errors.Structural, "Not a Handler packet")
	}
	e.HPacket = packet
	return nil
}
