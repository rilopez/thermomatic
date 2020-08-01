// Package common implements utilities & functionality commonly consumed by the
// rest of the packages.
package common

import (
	"encoding/binary"
	"errors"
	"math"
	"testing"
)

// ErrNotImplemented is raised throughout the codebase of the challenge to
// denote implementations to be done by the candidate.
var ErrNotImplemented = errors.New("not implemented")

func CreatePayload(expectedTemperature float64, expectedAltitude float64, expectedLatitude float64, expectedLongitude float64, expectedBatteryLevel float64) []byte {
	payload := make([]byte, 40)
	binary.BigEndian.PutUint64(payload[0:], math.Float64bits(expectedTemperature))
	binary.BigEndian.PutUint64(payload[8:], math.Float64bits(expectedAltitude))
	binary.BigEndian.PutUint64(payload[16:], math.Float64bits(expectedLatitude))
	binary.BigEndian.PutUint64(payload[24:], math.Float64bits(expectedLongitude))
	binary.BigEndian.PutUint64(payload[32:], math.Float64bits(expectedBatteryLevel))
	return payload
}

//TODO move to testing utility module
func ShouldPanic(t *testing.T, f func()) {
	defer func() { recover() }()
	f()
	t.Errorf("should have panicked")
}
