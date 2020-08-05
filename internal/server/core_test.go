package server

import (
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/spin-org/thermomatic/internal/client"
)

func frozenInTime() time.Time {
	loc, err := time.LoadLocation("EST")
	if err != nil {
		panic("Could not load EST location")
	}
	//Dragon splashdown historic event time
	return time.Date(2020, 8, 2, 14, 48, 0, 0, loc)
}

func TestNewCore(t *testing.T) {
	core := newCore(frozenInTime)
	expectedClientsLen := 0
	if len(core.clients) != expectedClientsLen {
		t.Errorf("expected len(core.client) to equal %d but got %d", expectedClientsLen, len(core.clients))
	}
}

func TestCore_HandleReading(t *testing.T) {
	//Setup
	expectedLastReadingEpoch := frozenInTime().UnixNano()
	expectedPayload := client.CreateRandReading()

	core := newCore(frozenInTime)
	expectedClientIMEI := uint64(448324242329542)
	device := &client.Client{IMEI: expectedClientIMEI}
	core.clients[expectedClientIMEI] = device

	//Exercise

	core.handleReading(expectedClientIMEI, expectedPayload[:])

	if device.LastReadingEpoch != expectedLastReadingEpoch {
		t.Errorf("expected LastReadingEpoch to equal %d but got %d",
			expectedLastReadingEpoch,
			device.LastReadingEpoch)
	}
	expectedReading := &client.Reading{}
	expectedReading.Decode(expectedPayload[:])
	if !reflect.DeepEqual(expectedReading, device.LastReading) {
		t.Errorf("expected LastReading to equal %v but got %v",
			expectedReading,
			device.LastReading)
	}
}

func TestCore_HandleReading_UnknownClient(t *testing.T) {
	//Setup
	core := newCore(frozenInTime)

	//Exercise

	unknownIMEI := uint64(123)
	err := core.handleReading(unknownIMEI, []byte{1, 2})
	if err == nil {
		t.Errorf("expected get an error for unknown client %d", unknownIMEI)
	}
}

func TestCore_HandleReading_InvalidPayload(t *testing.T) {
	//Setup
	core := newCore(frozenInTime)
	expectedClientIMEI := uint64(448324242329542)
	device := &client.Client{IMEI: expectedClientIMEI}
	core.clients[expectedClientIMEI] = device

	//Exercise bound check panic
	errBoundCheckPanic := core.handleReading(expectedClientIMEI, []byte{1, 2})
	if errBoundCheckPanic == nil {
		t.Errorf("expected get an error for unknown client %d", expectedClientIMEI)
	}

	invalidPayload := client.NewPayload(9999999, 9999999, 9999999, 9999999, 9999999)

	errInvalidPayload := core.handleReading(expectedClientIMEI, invalidPayload[:])
	if errInvalidPayload == nil {
		t.Errorf("expected get an error for unknown client %d", expectedClientIMEI)
	}
}

func TestCore_Register(t *testing.T) {
	core := newCore(frozenInTime)
	expectedClientIMEI := uint64(448324242329542)
	device := &client.Client{IMEI: expectedClientIMEI}

	err := core.register(device)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	_, exists := core.clients[device.IMEI]
	if !exists {
		t.Errorf("clients map should contain an entry for IMEI: %d", expectedClientIMEI)
	}
}

func TestCore_Register_ExistingClient(t *testing.T) {
	// Setup
	core := newCore(frozenInTime)
	expectedClientIMEI := uint64(448324242329542)
	device := &client.Client{
		IMEI: expectedClientIMEI,
		Conn: &net.UnixConn{},
	}

	//Exercise
	core.register(device)

	err := core.register(device)
	if err == nil {
		t.Errorf("An error is expected when trying to register an existing client ")
	}
}

func TestCore_Deregister_ExistingClient(t *testing.T) {
	// Setup
	core := newCore(frozenInTime)
	expectedClientIMEI := uint64(448324242329542)
	device := &client.Client{
		IMEI: expectedClientIMEI,
	}

	//Exercise
	core.register(device)

	err := core.deregister(device)
	if err != nil {
		t.Errorf("Unexpected error trying to deregister an existing client %v ", err)
	}
}

func TestCore_Deregister_UnknownClient(t *testing.T) {
	// Setup
	core := newCore(frozenInTime)
	expectedClientIMEI := uint64(448324242329542)
	device := &client.Client{
		IMEI: expectedClientIMEI,
	}

	//Exercise

	err := core.deregister(device)
	if err == nil {
		t.Errorf("An error is expected when trying to deregister an unknown client")
	}
}
