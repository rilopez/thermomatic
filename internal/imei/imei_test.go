package imei

import (
	"runtime"
	"testing"
)

func TestDecode(t *testing.T) {
	b := []byte{4, 9, 0, 1, 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 8}
	expectedIMEI := uint64(490154203237518)
	actualIMEI, _ := Decode(b)

	if actualIMEI != expectedIMEI {
		t.Errorf("expecting imei %d but got %d", expectedIMEI, actualIMEI)
	}
}

func TestDecodeErrCheckSum(t *testing.T) {
	b := []byte{4, 9, 0, 1, 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 1}
	_, err := Decode(b)
	if err != ErrChecksum {
		t.Errorf("expecting ErrChecksum but got  %v", err)
	}
}

func TestDecodeWithNonDigitsErrInvalid(t *testing.T) {
	b := []byte{4, 9, 'A', 'B', 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 8}
	_, err := Decode(b)
	if err != ErrInvalid {
		t.Errorf("expecting ErrInvalid but got %s", err)
	}
}

func TestDecodePanicAtLeast15byteslong(t *testing.T) {
	//it panics if b isn't at least 15 bytes long.
	onlyTwoBytes := []byte{1, 2}

	shouldPanic(t, func() {
		_, _ = Decode(onlyTwoBytes)
	})
}

func TestDecodeAllocations(t *testing.T) {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(1))
	var start, end runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&start)
	_, _ = Decode([]byte{4, 9, 0, 1, 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 8})
	runtime.ReadMemStats(&end)
	alloc := end.TotalAlloc - start.TotalAlloc
	if alloc > 0 {
		t.Errorf("Decode should NOT allocate under any condition, it allocated %d bytes", alloc)
	}
}

func BenchmarkDecode(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Decode([]byte{4, 9, 0, 1, 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 8})
	}
	b.StopTimer()
}

func shouldPanic(t *testing.T, f func()) {
	defer func() { recover() }()
	f()
	t.Errorf("should have panicked")
}
