// Package common implements utilities & functionality commonly consumed by the
// rest of the packages.
package common

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"testing"
)

// ErrNotImplemented is raised throughout the codebase of the challenge to
// denote implementations to be done by the candidate.
var ErrNotImplemented = errors.New("not implemented")

func CreatePayload(expectedTemperature float64, expectedAltitude float64, expectedLatitude float64, expectedLongitude float64, expectedBatteryLevel float64) [40]byte {
	var payload [40]byte
	binary.BigEndian.PutUint64(payload[0:], math.Float64bits(expectedTemperature))
	binary.BigEndian.PutUint64(payload[8:], math.Float64bits(expectedAltitude))
	binary.BigEndian.PutUint64(payload[16:], math.Float64bits(expectedLatitude))
	binary.BigEndian.PutUint64(payload[24:], math.Float64bits(expectedLongitude))
	binary.BigEndian.PutUint64(payload[32:], math.Float64bits(expectedBatteryLevel))
	return payload
}

func ImeiStringToBytes(imei *string) ([15]byte, error) {
	var buf [15]byte

	if len(*imei) < 15 {
		return buf, errors.New("IMEI string should have at least 15 characters")
	}
	for i := 0; i < len(*imei); i++ {
		s := (*imei)[i]
		digit, err := strconv.Atoi(string(s))
		if err == nil {
			buf[i] = byte(digit)
		} else {
			return buf, errors.New(fmt.Sprintf("IMEI has an invalid digit character at %d", i))
		}
	}
	return buf, nil
}

//TODO move to testing utility module
func ShouldPanic(t *testing.T, f func()) {
	defer func() { recover() }()
	f()
	t.Errorf("should have panicked")
}
