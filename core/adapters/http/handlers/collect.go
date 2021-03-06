// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"io"
	"net/http"

	. "github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// Collect defines a handler for retrieving raw packets sent by a POST request.
//
// It listens to requests of the form: [POST] /packets/
//
// It expects an http header Content-Type = application/octet-stream
//
// The body is expected to a binary marshaling of the given packet
//
// This handler does not generate any registration.
type Collect struct{}

// Url implements the http.Handler interface
func (p Collect) Url() string {
	return "/packets/"
}

// Handle implements the http.Handler interface
func (p Collect) Handle(w http.ResponseWriter, chpkt chan<- PktReq, chreg chan<- RegReq, req *http.Request) {
	// Check the http method
	if req.Method != "POST" {
		err := errors.New(errors.Structural, "Unreckognized HTTP method. Please use [POST] to transfer a packet")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(err.Error()))
		return
	}

	// Parse body and query params
	data, err := p.parse(req)
	if err != nil {
		BadRequest(w, err.Error())
		return
	}

	// Send the packet and wait for ack / nack
	chresp := make(chan MsgRes)
	chpkt <- PktReq{Packet: data, Chresp: chresp}
	r, ok := <-chresp
	if !ok {
		err := errors.New(errors.Operational, "Core server not responding")
		BadRequest(w, err.Error())
		return
	}
	w.WriteHeader(r.StatusCode)
	w.Write(r.Content)
}

// parse extracts params from the request and fails if the request is invalid.
func (p Collect) parse(req *http.Request) ([]byte, error) {
	// Check Content-type
	if req.Header.Get("Content-Type") != "application/octet-stream" {
		return nil, errors.New(errors.Structural, "Received invalid content-type in request")
	}

	// Check configuration in body
	body := make([]byte, req.ContentLength)
	n, err := req.Body.Read(body)
	if err != nil && err != io.EOF {
		return nil, errors.New(errors.Structural, err)
	}
	return body[:n], nil
}
