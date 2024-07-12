// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package dtls

import (
	"errors"
	"reflect"
	"testing"
)

func TestFragmentBuffer(t *testing.T) {
	for _, test := range []struct {
		Name     string
		In       [][]byte
		Expected [][]byte
		Epoch    uint16
	}{
		{
			Name: "Single Fragment",
			In: [][]byte{
				{0x16, 0xfe, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0F, 0x03, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xfe, 0xff, 0x00},
			},
			Expected: [][]byte{
				{0x03, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xfe, 0xff, 0x00},
			},
			Epoch: 0,
		},
		{
			Name: "Single Fragment Epoch 3",
			In: [][]byte{
				{0x16, 0xfe, 0xff, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0F, 0x03, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xfe, 0xff, 0x00},
			},
			Expected: [][]byte{
				{0x03, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xfe, 0xff, 0x00},
			},
			Epoch: 3,
		},
		{
			Name: "Multiple Fragments",
			In: [][]byte{
				{0x16, 0xfe, 0xfd, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x81, 0x0b, 0x00, 0x00, 0x0F, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x01, 0x02, 0x03, 0x04},
				{0x16, 0xfe, 0xfd, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x81, 0x0b, 0x00, 0x00, 0x0F, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x05, 0x05, 0x06, 0x07, 0x08, 0x09},
				{0x16, 0xfe, 0xfd, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x81, 0x0b, 0x00, 0x00, 0x0F, 0x00, 0x00, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x05, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E},
			},
			Expected: [][]byte{
				{0x0b, 0x00, 0x00, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0f, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e},
			},
			Epoch: 0,
		},
		{
			Name: "Multiple Unordered Fragments",
			In: [][]byte{
				{0x16, 0xfe, 0xfd, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x81, 0x0b, 0x00, 0x00, 0x0F, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x01, 0x02, 0x03, 0x04},
				{0x16, 0xfe, 0xfd, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x81, 0x0b, 0x00, 0x00, 0x0F, 0x00, 0x00, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x05, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E},
				{0x16, 0xfe, 0xfd, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x81, 0x0b, 0x00, 0x00, 0x0F, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x05, 0x05, 0x06, 0x07, 0x08, 0x09},
			},
			Expected: [][]byte{
				{0x0b, 0x00, 0x00, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0f, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e},
			},
			Epoch: 0,
		},
		{
			Name: "Multiple Handshakes in Single Fragment",
			In: [][]byte{
				{
					0x16, 0xfe, 0xfd, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x30, /* record header */
					0x03, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0xfe, 0xff, 0x01, 0x01, /*handshake msg 1*/
					0x03, 0x00, 0x00, 0x04, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0xfe, 0xff, 0x01, 0x01, /*handshake msg 2*/
					0x03, 0x00, 0x00, 0x04, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0xfe, 0xff, 0x01, 0x01, /*handshake msg 3*/
				},
			},
			Expected: [][]byte{
				{0x03, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0xfe, 0xff, 0x01, 0x01},
				{0x03, 0x00, 0x00, 0x04, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0xfe, 0xff, 0x01, 0x01},
				{0x03, 0x00, 0x00, 0x04, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0xfe, 0xff, 0x01, 0x01},
			},
			Epoch: 0,
		},
		// Assert that a zero length fragment doesn't cause the fragmentBuffer to enter an infinite loop
		{
			Name: "Zero Length Fragment",
			In: [][]byte{
				{
					0x16, 0xfe, 0xfd, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0d, 0x00, 0x00,
					0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
			},
			Expected: [][]byte{
				{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00},
			},
			Epoch: 0,
		},
	} {
		fragmentBuffer := newFragmentBuffer()
		for _, frag := range test.In {
			status, _, err := fragmentBuffer.push(frag)
			if err != nil {
				t.Error(err)
			} else if !status {
				t.Errorf("fragmentBuffer didn't accept fragments for '%s'", test.Name)
			}
		}

		for _, expected := range test.Expected {
			out, epoch := fragmentBuffer.pop()
			if !reflect.DeepEqual(out, expected) {
				t.Errorf("fragmentBuffer '%s' push/pop: got % 02x, want % 02x", test.Name, out, expected)
			}
			if epoch != test.Epoch {
				t.Errorf("fragmentBuffer returned wrong epoch: got %d, want %d", epoch, test.Epoch)
			}
		}

		if frag, _ := fragmentBuffer.pop(); frag != nil {
			t.Errorf("fragmentBuffer popped single buffer multiple times for '%s'", test.Name)
		}
	}
}

func TestFragmentBuffer_Overflow(t *testing.T) {
	fragmentBuffer := newFragmentBuffer()

	// Push a buffer that doesn't exceed size limits
	if _, _, err := fragmentBuffer.push([]byte{0x16, 0xfe, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0F, 0x03, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xfe, 0xff, 0x00}); err != nil {
		t.Fatal(err)
	}

	// Allocate a buffer that exceeds cache size
	largeBuffer := make([]byte, fragmentBufferMaxSize)
	if _, _, err := fragmentBuffer.push(largeBuffer); !errors.Is(err, errFragmentBufferOverflow) {
		t.Fatalf("Pushing a large buffer returned (%s) expected(%s)", err, errFragmentBufferOverflow)
	}
}
