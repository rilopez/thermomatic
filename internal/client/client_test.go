package client

import (
	"testing"
)

func TestString(t *testing.T) {

	c := &Client{
		IMEI:             490154203237518,
		LastReadingEpoch: 1257894000000000000,
		LastReading: &Reading{
			Temperature:  67.77,
			Altitude:     2.63555,
			Latitude:     33.41,
			Longitude:    44.4,
			BatteryLevel: 0.25666,
		},
	}

	expectedRecord := "1257894000000000000,490154203237518,67.770000,2.635550,33.410000,44.400000,0.256660"
	actualRecord := c.String()
	//TODO follow up email to Rob about field string format requirements
	if expectedRecord != actualRecord {
		t.Errorf("expected %s got %s", expectedRecord, actualRecord)
	}
}
