package client

import (
	"encoding/binary"
	"math"
)

// Reading is the set of device readings.
type Reading struct {
	// Temperature denotes the temperature reading of the message.
	Temperature float64

	// Altitude denotes the altitude reading of the message.
	Altitude float64

	// Latitude denotes the latitude reading of the message.
	Latitude float64

	// Longitude denotes the longitude reading of the message.
	Longitude float64

	// BatteryLevel denotes the battery level reading of the message.
	BatteryLevel float64
}

// Decode decodes the reading message payload in the given b into r.
//
// If any of the fields are outside their valid min/max ranges ok will be unset.
//
// Decode does NOT allocate under any condition. Additionally, it panics if b
// isn't at least 40 bytes long.
func (r *Reading) Decode(b []byte) (ok bool) {
	_ = b[39] // compiler bound check hint
	temperature := math.Float64frombits(binary.BigEndian.Uint64(b[0:]))
	if isInRange(temperature, -300, 300) {
		r.Temperature = temperature
	} else {
		ok = false
	}

	altitude := math.Float64frombits(binary.BigEndian.Uint64(b[8:]))
	if isInRange(altitude, -20000, 20000) {
		r.Altitude = altitude
	} else {
		ok = false
	}

	latitude := math.Float64frombits(binary.BigEndian.Uint64(b[16:]))
	if isInRange(latitude, -90, 90) {
		r.Latitude = latitude
	} else {
		ok = false
	}
	longitude := math.Float64frombits(binary.BigEndian.Uint64(b[24:]))
	if isInRange(longitude, -180, 180) {
		r.Longitude = longitude
	} else {
		ok = false
	}

	batteryLevel := math.Float64frombits(binary.BigEndian.Uint64(b[32:]))
	if isInRange(batteryLevel, 0, 100) {
		r.BatteryLevel = batteryLevel
	} else {
		ok = false
	}

	return ok
}

func isInRange(value, min, max float64) bool {
	return value >= min && value <= max
}
