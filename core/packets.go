// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"encoding"
	"fmt"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
	"github.com/brocaar/lorawan"
)

const (
	type_rpacket byte = iota
	type_bpacket
	type_hpacket
	type_apacket
	type_jpacket
	type_cpacket
	type_spacket
)

// ---------------------------------
//
// ----- HELPERS -------------------
//
// ---------------------------------

// marshalBases is used to marshal in chain several bases which compose a bigger Packet struct
func marshalBases(t byte, bases ...baseMarshaler) ([]byte, error) {
	data := []byte{t}

	for _, base := range bases {
		dataBase, err := base.Marshal()
		if err != nil {
			return nil, err
		}
		data = append(data, dataBase...)
	}
	return data, nil
}

// unmarshalBases do the reverse operation of marshalBases
func unmarshalBases(t byte, data []byte, bases ...baseUnmarshaler) error {
	if len(data) < 1 || data[0] != t {
		return errors.New(errors.Structural, "Not an expected packet")
	}

	var rest []byte
	var err error

	rest = data[1:]
	for _, base := range bases {
		if rest, err = base.Unmarshal(rest); err != nil {
			return err
		}
	}

	return err
}

// UnmarshalPacket takes raw binary data and try to marshal it into a given packet interface:
//
// - RPacket
// - BPacket
// - HPacket
// - APacket
// - JPacket
// - CPacket
//
// It returns an interface so that its easy and handy to perform a type assertion out of it.
// If data are wrong or, if the packet is not unmarshalable, it returns an error.
func UnmarshalPacket(data []byte) (interface{}, error) {
	if len(data) < 1 {
		return nil, errors.New(errors.Structural, "Cannot unmarshal, not a packet")
	}

	var packet interface {
		encoding.BinaryUnmarshaler
	}

	switch data[0] {
	case type_rpacket:
		packet = new(rpacket)
	case type_bpacket:
		packet = new(bpacket)
	case type_hpacket:
		packet = new(hpacket)
	case type_apacket:
		packet = new(apacket)
	case type_jpacket:
		packet = new(jpacket)
	case type_cpacket:
		packet = new(cpacket)
	}

	err := packet.UnmarshalBinary(data)
	return packet, err
}

// ---------------------------------
//
// ----- RPACKET -------------------
//
// ---------------------------------

// rpacket implements the core.RPacket interface
type rpacket struct {
	baserpacket
	basempacket
	gatewayId baseapacket
}

// NewRPacket construct a new router packet given a payload and metadata
func NewRPacket(payload lorawan.PHYPayload, gatewayId []byte, metadata Metadata) (RPacket, error) {
	if payload.MACPayload == nil {
		return nil, errors.New(errors.Structural, "MACPAyload should not be empty")
	}

	_, ok := payload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return nil, errors.New(errors.Structural, "Packet does not carry a MACPayload")
	}

	return &rpacket{
		baserpacket: baserpacket{payload: payload},
		basempacket: basempacket{metadata: metadata},
		gatewayId:   baseapacket{payload: gatewayId},
	}, nil
}

// GatewayId implements the core.RPacket interface
func (p rpacket) GatewayId() []byte {
	return p.gatewayId.payload
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p rpacket) MarshalBinary() ([]byte, error) {
	return marshalBases(type_rpacket, p.baserpacket, p.basempacket, p.gatewayId)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (p *rpacket) UnmarshalBinary(data []byte) error {
	return unmarshalBases(type_rpacket, data, &p.baserpacket, &p.basempacket, &p.gatewayId)
}

// String implements the Stringer interface
func (p rpacket) String() string {
	str := "Packet {"
	str += fmt.Sprintf("\n\t%s}", p.metadata.String())
	str += fmt.Sprintf("\n\tPayload%+v\n}", p.payload)
	return str
}

// ---------------------------------
//
// ----- BPACKET -------------------
//
// ---------------------------------

// bpacket implements the core.BPacket interface
type bpacket struct {
	baserpacket
	basempacket
}

// NewBPacket constructs a new broker packets given a payload and metadata
func NewBPacket(payload lorawan.PHYPayload, metadata Metadata) (BPacket, error) {
	if payload.MACPayload == nil {
		return nil, errors.New(errors.Structural, "MACPAyload should not be empty")
	}

	macPayload, ok := payload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return nil, errors.New(errors.Structural, "Packet does not carry a MACPayload")
	}

	if len(macPayload.FRMPayload) != 1 {
		return nil, errors.New(errors.Structural, "Invalid frame payload. Expected exactly 1")
	}

	if _, ok := macPayload.FRMPayload[0].(*lorawan.DataPayload); !ok {
		return nil, errors.New(errors.Structural, "Invalid frame payload. Expected only data")
	}

	return &bpacket{
		baserpacket: baserpacket{payload: payload},
		basempacket: basempacket{metadata: metadata},
	}, nil
}

// ValidateMIC implements the core.BPacket interface
func (p bpacket) ValidateMIC(key lorawan.AES128Key) (bool, error) {
	return p.baserpacket.payload.ValidateMIC(key)
}

// Commands implements the core.BPacket interface
func (p bpacket) Commands() []lorawan.MACCommand {
	return p.baserpacket.payload.MACPayload.(*lorawan.MACPayload).FHDR.FOpts
}

// String implements the fmt.Stringer interface
func (p bpacket) String() string {
	return "TODO"
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p bpacket) MarshalBinary() ([]byte, error) {
	return marshalBases(type_bpacket, p.baserpacket, p.basempacket)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (p *bpacket) UnmarshalBinary(data []byte) error {
	return unmarshalBases(type_bpacket, data, &p.baserpacket, &p.basempacket)
}

// ---------------------------------
//
// ----- HPACKET -------------------
//
// ---------------------------------

// hpacket implements the HPacket interface
type hpacket struct {
	basehpacket
	payload baserpacket
	basempacket
}

// NewHPacket constructs a new Handler packet
func NewHPacket(appEUI lorawan.EUI64, devEUI lorawan.EUI64, payload lorawan.PHYPayload, metadata Metadata) (HPacket, error) {
	if payload.MACPayload == nil {
		return nil, errors.New(errors.Structural, "MACPAyload should not be empty")
	}

	macPayload, ok := payload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return nil, errors.New(errors.Structural, "Packet does not carry a MACPayload")
	}

	if len(macPayload.FRMPayload) != 1 {
		return nil, errors.New(errors.Structural, "Invalid frame payload. Expected exactly 1")
	}

	if _, ok := macPayload.FRMPayload[0].(*lorawan.DataPayload); !ok {
		return nil, errors.New(errors.Structural, "Invalid frame payload. Expected only data")
	}

	return &hpacket{
		basehpacket: basehpacket{
			appEUI: appEUI,
			devEUI: devEUI,
		},
		payload: baserpacket{
			payload: payload,
		},
		basempacket: basempacket{metadata: metadata},
	}, nil
}

// Payload implements the core.HPacket interface
func (p hpacket) Payload(key lorawan.AES128Key) ([]byte, error) {
	macPayload := p.payload.payload.MACPayload.(*lorawan.MACPayload)
	if err := macPayload.DecryptFRMPayload(key); err != nil {
		return nil, errors.New(errors.Structural, err)
	}
	if len(macPayload.FRMPayload) != 1 {
		return nil, errors.New(errors.Structural, "Unexpected Frame payload")
	}
	return macPayload.FRMPayload[0].(*lorawan.DataPayload).Bytes, nil
}

// FCnt implements the core.HPacket interface
func (p hpacket) FCnt() uint32 {
	return p.payload.FCnt()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p hpacket) MarshalBinary() ([]byte, error) {
	return marshalBases(type_hpacket, p.basehpacket, p.payload, p.basempacket)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (p *hpacket) UnmarshalBinary(data []byte) error {
	return unmarshalBases(type_hpacket, data, &p.basehpacket, &p.payload, &p.basempacket)
}

// String implements the fmt.Stringer interface
func (p hpacket) String() string {
	str := "Packet {"
	str += fmt.Sprintf("\n\t%s}", p.metadata.String())
	str += fmt.Sprintf("\n\tAppEUI:%+x\n,", p.appEUI)
	str += fmt.Sprintf("\n\tDevEUI:%+x\n,", p.devEUI)
	str += fmt.Sprintf("\n\tPayload:%v\n}", p.Payload)
	return str
}

// ---------------------------------
//
// ----- APACKET -------------------
//
// ---------------------------------

// apacket implements the core.APacket interface
type apacket struct {
	baseapacket
	devEUI baseapacket
	basegpacket
}

// NewAPacket constructs a new application packet
func NewAPacket(payload []byte, devEUI lorawan.EUI64, metadata []Metadata) (APacket, error) {
	if len(payload) == 0 {
		return nil, errors.New(errors.Structural, "Application packet must hold a payload")
	}

	return &apacket{
		baseapacket: baseapacket{payload: payload},
		devEUI:      baseapacket{payload: devEUI[:]},
		basegpacket: basegpacket{metadata: metadata},
	}, nil
}

// DevEUI implements the core.APacket interface
func (p apacket) DevEUI() lorawan.EUI64 {
	var devEUI lorawan.EUI64
	copy(devEUI[:], p.devEUI.payload)
	return devEUI
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p apacket) MarshalBinary() ([]byte, error) {
	return marshalBases(type_apacket, p.baseapacket, p.devEUI, p.basegpacket)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (p *apacket) UnmarshalBinary(data []byte) error {
	return unmarshalBases(type_apacket, data, &p.baseapacket, &p.devEUI, &p.basegpacket)
}

// String implements the fmt.Stringer interface
func (p apacket) String() string {
	return "TODO"
}

// ---------------------------------
//
// ----- JPACKET -------------------
//
// ---------------------------------

// joinPacket implements the core.JoinPacket interface
type jpacket struct {
	baseapacket baseapacket
	basehpacket
	basempacket
}

// NewJoinPacket constructs a new JoinPacket
func NewJPacket(appEUI lorawan.EUI64, devEUI lorawan.EUI64, devNonce [2]byte, metadata Metadata) JPacket {
	return &jpacket{
		basehpacket: basehpacket{
			appEUI: appEUI,
			devEUI: devEUI,
		},
		baseapacket: baseapacket{payload: devNonce[:]},
		basempacket: basempacket{metadata: metadata},
	}
}

// DevNonce implements the core.JoinPacket interface
func (p jpacket) DevNonce() [2]byte {
	return [2]byte{p.baseapacket.payload[0], p.baseapacket.payload[1]}
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p jpacket) MarshalBinary() ([]byte, error) {
	return marshalBases(type_jpacket, p.basehpacket, p.baseapacket, p.basempacket)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (p *jpacket) UnmarshalBinary(data []byte) error {
	return unmarshalBases(type_jpacket, data, &p.basehpacket, &p.baseapacket, &p.basempacket)
}

// String implements the fmt.Stringer interface
func (p jpacket) String() string {
	return "TODO"
}

// acceptpacket implements the core.AcceptPacket interface
type cpacket struct {
	basehpacket
	baseapacket
	nwkSKey baseapacket
}

// NewAcceptPacket constructs a new CPacket
func NewCPacket(appEUI lorawan.EUI64, devEUI lorawan.EUI64, payload []byte, nwkSKey lorawan.AES128Key) (CPacket, error) {
	if len(payload) == 0 {
		return nil, errors.New(errors.Structural, "Payload cannot be empty")
	}

	return &cpacket{
		basehpacket: basehpacket{
			appEUI: appEUI,
			devEUI: devEUI,
		},
		baseapacket: baseapacket{payload: payload},
		nwkSKey:     baseapacket{payload: nwkSKey[:]},
	}, nil
}

// NwkSKey implements the core.AcceptPacket interface
func (p cpacket) NwkSKey() lorawan.AES128Key {
	var key lorawan.AES128Key
	copy(key[:], p.nwkSKey.payload)
	return key
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (p cpacket) MarshalBinary() ([]byte, error) {
	return marshalBases(type_cpacket, p.basehpacket, p.baseapacket, p.nwkSKey)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (p *cpacket) UnmarshalBinary(data []byte) error {
	return unmarshalBases(type_cpacket, data, &p.basehpacket, &p.baseapacket, &p.nwkSKey)
}

// String implements the fmt.Stringer interface
func (p cpacket) String() string {
	return "TODO"
}

// --------------------------------------
// --------------------------------------
// --------------------------------------
// ----- BASE PACKETS -------------------
//
// All base packet are small components that are used by packets above to define accessors and
// marshaling / unmarshaling methods on a struct.
// All Unmarshal methods return the remaining unconsumed bytes from the input data such that one
// could actually chain calls for different basexxxpacket
//
// --------------------------------------
// --------------------------------------
// --------------------------------------
//
// basempacket -> metadata Metadata
// baserpacket -> payload lorawan.PHYPayload
// baseapacket -> payload []byte
// basehpacket -> appEUI lorawan.EUI64, devEUI lorawan.EUI64
// (ALWAYS LAST) basegpacket -> metadata []Metadata

type baseMarshaler interface {
	Marshal() ([]byte, error)
}

type baseUnmarshaler interface {
	Unmarshal(data []byte) ([]byte, error)
}

// basempacket is used to compose other packets
type basempacket struct {
	metadata Metadata
}

func (p basempacket) Metadata() Metadata {
	return p.metadata
}

func (p basempacket) Marshal() ([]byte, error) {
	dataMetadata, err := p.metadata.MarshalJSON()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	rw := readwriter.New(nil)
	rw.Write(dataMetadata)
	return rw.Bytes()
}

func (p *basempacket) Unmarshal(data []byte) ([]byte, error) {
	rw := readwriter.New(data)

	var dataMetadata []byte
	rw.Read(func(data []byte) { dataMetadata = data })

	p.metadata = Metadata{}
	if err := p.metadata.UnmarshalJSON(dataMetadata); err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	return rw.Bytes()
}

// baserpacket is used to compose other packets
type baserpacket struct {
	payload lorawan.PHYPayload
}

func (p baserpacket) Payload() lorawan.PHYPayload {
	return p.payload
}

// DevEUI implements the core.RPacket interface
func (p baserpacket) DevEUI() lorawan.EUI64 {
	var devEUI lorawan.EUI64
	copy(devEUI[4:], p.payload.MACPayload.(*lorawan.MACPayload).FHDR.DevAddr[:])
	return devEUI
}

// FCnt implements the core.BPacket interface
func (p baserpacket) FCnt() uint32 {
	return p.payload.MACPayload.(*lorawan.MACPayload).FHDR.FCnt
}

// Marshal transforms the given basepacket to binaries
func (p baserpacket) Marshal() ([]byte, error) {
	var mtype byte
	switch p.payload.MHDR.MType {
	case lorawan.JoinRequest:
		fallthrough
	case lorawan.UnconfirmedDataUp:
		fallthrough
	case lorawan.ConfirmedDataUp:
		mtype = 1 // Up
	case lorawan.JoinAccept:
		fallthrough
	case lorawan.UnconfirmedDataDown:
		fallthrough
	case lorawan.ConfirmedDataDown:
		mtype = 2 // Down
	default:
		msg := fmt.Sprintf("Unsupported mtype: %s", p.payload.MHDR.MType.String())
		return nil, errors.New(errors.Implementation, msg)
	}

	dataPayload, err := p.payload.MarshalBinary()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	rw := readwriter.New(nil)
	rw.Write(mtype)
	rw.Write(dataPayload)
	return rw.Bytes()
}

// Unmarshal hydrates the given basepacket from binaries data.
func (p *baserpacket) Unmarshal(data []byte) ([]byte, error) {
	if len(data) < 1 {
		return nil, errors.New(errors.Structural, "Not a valid packet")
	}

	var isUp bool
	rw := readwriter.New(data)
	rw.Read(func(data []byte) {
		if data[0] == 1 {
			isUp = true
		}
	})

	var dataPayload []byte
	rw.Read(func(data []byte) { dataPayload = data })

	data, err := rw.Bytes()
	if err != nil {
		return nil, errors.New(errors.Structural, rw.Err())
	}

	p.payload = lorawan.NewPHYPayload(isUp)
	if err := p.payload.UnmarshalBinary(dataPayload); err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	return data, nil
}

// basehpacket is used to compose other packets
type basehpacket struct {
	appEUI lorawan.EUI64
	devEUI lorawan.EUI64
}

func (p basehpacket) AppEUI() lorawan.EUI64 {
	return p.appEUI
}

func (p basehpacket) DevEUI() lorawan.EUI64 {
	return p.devEUI
}

func (p basehpacket) Marshal() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(p.appEUI)
	rw.Write(p.devEUI)
	return rw.Bytes()
}

func (p *basehpacket) Unmarshal(data []byte) ([]byte, error) {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) { copy(p.appEUI[:], data) })
	rw.Read(func(data []byte) { copy(p.devEUI[:], data) })
	return rw.Bytes()
}

// baseapacket is used to compose other packets
type baseapacket struct {
	payload []byte
}

func (p baseapacket) Payload() []byte {
	return p.payload
}

func (p baseapacket) Marshal() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(p.payload)
	return rw.Bytes()
}

func (p *baseapacket) Unmarshal(data []byte) ([]byte, error) {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) { p.payload = data })
	return rw.Bytes()
}

// basegpacket is used to compose other packets
type basegpacket struct {
	metadata []Metadata
}

func (p basegpacket) Metadata() []Metadata {
	return p.metadata
}

func (p basegpacket) Marshal() ([]byte, error) {
	rw := readwriter.New(nil)
	for _, m := range p.metadata {
		data, err := m.MarshalJSON()
		if err != nil {
			return nil, errors.New(errors.Structural, err)
		}
		rw.Write(data)
	}
	return rw.Bytes()
}

func (p *basegpacket) Unmarshal(data []byte) ([]byte, error) {
	p.metadata = make([]Metadata, 0)
	rw := readwriter.New(data)

	for {
		var dataMetadata []byte
		rw.Read(func(data []byte) { dataMetadata = data })
		if rw.Err() != nil {
			err, ok := rw.Err().(errors.Failure)
			if ok && err.Nature == errors.Behavioural {
				break
			}
			return nil, errors.New(errors.Structural, rw.Err())
		}
		metadata := new(Metadata)
		if err := metadata.UnmarshalJSON(dataMetadata); err != nil {
			return nil, errors.New(errors.Structural, err)
		}

		p.metadata = append(p.metadata, *metadata)
	}

	return nil, nil
}
