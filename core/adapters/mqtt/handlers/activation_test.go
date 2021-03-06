// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/adapters/mqtt"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

func TestActionTopic(t *testing.T) {
	wantTopic := "+/devices/+/activations"

	// Describe
	Desc(t, "Topic should equal: %s", wantTopic)

	// Build
	handler := Activation{}

	// Operate
	topic := handler.Topic()

	// Check
	checkTopics(t, wantTopic, topic)
}

func TestActivationHandle(t *testing.T) {
	tests := []struct {
		Desc    string      // The test's description
		Client  *testClient // An mqtt client to mock (or not) the behavior
		Topic   string      // The topic to which the message is addressed
		Payload []byte      // The message's payload

		WantError        *string            // The expected error from the handler
		WantSubscription *string            // The topic to which a subscription is expected
		WantRegistration core.HRegistration // The expected registration towards the adapter
		WantPacket       []byte             // The expected packet towards the adapter
	}{
		{
			Desc:   "Ok client | Valid Topic | Valid Payload",
			Client: newTestClient(),
			Topic:  "0101010101010101/devices/personalized/activations",
			Payload: []byte{ // DevEUI | NwkSKey | AppSKey
				02, 02, 02, 02,
				03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03,
				04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04,
			},

			WantError:        nil,
			WantSubscription: pointer.String("0101010101010101/devices/0000000002020202/up"),
			WantRegistration: activationRegistration{
				recipient: NewRecipient("0101010101010101/devices/0000000002020202/up", "WHATEVER"),
				devEUI:    lorawan.EUI64([8]byte{0, 0, 0, 0, 2, 2, 2, 2}),
				appEUI:    lorawan.EUI64([8]byte{1, 1, 1, 1, 1, 1, 1, 1}),
				nwkSKey:   lorawan.AES128Key([16]byte{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3}),
				appSKey:   lorawan.AES128Key([16]byte{4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4}),
			},
			WantPacket: nil,
		},
		{
			Desc:   "Ok client | Invalid Topic #1 | Valid Payload",
			Client: newTestClient(),
			Topic:  "PleaseRegisterMyDevice",
			Payload: []byte{ // DevEUI | NwkSKey | AppSKey
				02, 02, 02, 02,
				03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03,
				04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04,
			},

			WantError:        pointer.String(string(errors.Structural)),
			WantSubscription: nil,
			WantRegistration: nil,
			WantPacket:       nil,
		},
		{
			Desc:   "Ok client | Invalid Topic #2 | Valid Payload",
			Client: newTestClient(),
			Topic:  "0101010101010101/devices/0001020304050607/activations",
			Payload: []byte{ // DevEUI | NwkSKey | AppSKey
				02, 02, 02, 02,
				03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03,
				04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04,
			},

			WantError:        pointer.String(string(errors.Implementation)),
			WantSubscription: nil,
			WantRegistration: nil,
			WantPacket:       nil,
		},
		{
			Desc:   "Ok client | Invalid Topic #3 | Valid Payload",
			Client: newTestClient(),
			Topic:  "01010101/devices/personalized/activations",
			Payload: []byte{ // DevEUI | NwkSKey | AppSKey
				02, 02, 02, 02,
				03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03,
				04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04,
			},

			WantError:        pointer.String(string(errors.Structural)),
			WantSubscription: nil,
			WantRegistration: nil,
			WantPacket:       nil,
		},
		{
			Desc:   "Ok client | Valid Topic | Invalid Payload #1",
			Client: newTestClient(),
			Topic:  "0101010101010101/devices/personalized/activations",
			Payload: []byte{ // DevEUI | NwkSKey | AppSKey
				02, 02, 02, 02,
				03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03,
				04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04,
				01, 02, 03, 04,
			},

			WantError:        pointer.String(string(errors.Structural)),
			WantSubscription: nil,
			WantRegistration: nil,
			WantPacket:       nil,
		},
		{
			Desc:   "Ok client | Valid Topic | Invalid Payload #2",
			Client: newTestClient(),
			Topic:  "0101010101010101/devices/personalized/activations",
			Payload: []byte{ // DevEUI | NwkSKey | AppSKey
				02, 02,
				03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03,
				04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04,
			},

			WantError:        pointer.String(string(errors.Structural)),
			WantSubscription: nil,
			WantRegistration: nil,
			WantPacket:       nil,
		},
		{
			Desc:   "Valid inputs | Client -> Fail Subscribe",
			Client: newTestClient("Subscribe"),
			Topic:  "0101010101010101/devices/personalized/activations",
			Payload: []byte{ // DevEUI | NwkSKey | AppSKey
				02, 02, 02, 02,
				03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03, 03,
				04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04, 04,
			},

			WantError:        pointer.String(string(errors.Operational)),
			WantSubscription: pointer.String("0101010101010101/devices/0000000002020202/up"),
			WantRegistration: nil,
			WantPacket:       nil,
		},
	}

	for i, test := range tests {
		// Describe
		Desc(t, "#%d: %s", i, test.Desc)

		// Build
		consumer, chpkt, chreg := newTestConsumer()
		handler := Activation{}

		// Operate
		err := handler.Handle(test.Client, chpkt, chreg, testMessage{
			payload: test.Payload,
			topic:   test.Topic,
		})

		// Check
		CheckErrors(t, test.WantError, err)
		checkSubscriptions(t, test.WantSubscription, test.Client.Subscription)
		checkRegistrations(t, test.WantRegistration, consumer.Registration)
		checkPackets(t, test.WantPacket, consumer.Packet)
	}
}
