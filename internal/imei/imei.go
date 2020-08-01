// Package imei implements an IMEI decoder.
package imei

// NOTE: for more information about IMEI codes and their structure you may
// consult with:
//
// https://en.wikipedia.org/wiki/International_Mobile_Equipment_Identity.

import (
	"errors"
	"math"
)

var (
	ErrInvalid  = errors.New("imei: invalid ")
	ErrChecksum = errors.New("imei: invalid checksum")
)

// Decode returns the IMEI code contained in the first 15 bytes of b.
//
// In case b isn't strictly composed of digits, the returned error will be
// ErrInvalid.
//
// In case b's checksum is wrong, the returned error will be ErrChecksum.
//
// Decode does NOT allocate under any condition. Additionally, it panics if b
// isn't at least 15 bytes long.
func Decode(b []byte) (code uint64, err error) {
	_ = b[14] // nice trick to hint bound checks to the compiler | https://medium.com/@brianblakewong/optimizing-go-bounds-check-elimination-f4be681ba030

	place := 14
	var checksum byte
	var imeiAsFloat float64

	for i := 0; i < 15; i++ {
		currDigit := b[i]
		digitForChecksum := currDigit
		if currDigit > 9 {
			// In case b isn't strictly composed of digits, the returned error will be
			// ErrInvalid.
			return 0, ErrInvalid
		}
		if (i+1)%2 == 0 {
			digitForChecksum *= 2
			if digitForChecksum > 9 {
				digitForChecksum = (digitForChecksum % 10) + 1
			}
		}
		checksum += digitForChecksum
		imeiAsFloat += float64(currDigit) * math.Pow10(place)
		place--
	}

	if checksum%10 != 0 {
		return 0, ErrChecksum
	}

	return uint64(imeiAsFloat), nil
}
