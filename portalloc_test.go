package portalloc

import (
	"errors"
	"net"
	"reflect"
	"strconv"
	"testing"
)

func TestPortAlloc(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		port := 20000
		wantPort := 20000
		var wantErr error = nil

		got, err := Alloc(uint64(port))
		if !reflect.DeepEqual(err, wantErr) {
			t.Fatalf("Alloc() error := %v, want := %v", err, wantErr)
		}

		if got <= 0 || !reflect.DeepEqual(port, wantPort) {
			t.Fatalf("Alloc() got := %v expected := %v", got, wantPort)
		}
	})

	t.Run("ErrPortIsBusy", func(t *testing.T) {
		port := 30000
		wantErr := ErrPortIsBusy

		helperAlloc(t, uint64(port))

		_, err := Alloc(uint64(port))
		if !reflect.DeepEqual(err, wantErr) {
			t.Fatalf("Alloc() error := %v, want := %v", err, wantErr)
		}
	})

	t.Run("NetAddrError", func(t *testing.T) {
		_, err := Alloc(100000)

		var netAddErr *net.AddrError
		if !errors.As(err, &netAddErr) {
			t.Fatalf("Alloc() error := %v, want := %v", err, netAddErr)

		}
	})
}

func TestAllocInRange(t *testing.T) {
	t.Run("AllocFirst", func(t *testing.T) {
		from := 20000
		to := 20001
		want := []uint64{20000, 20001}
		var wantErr error = nil

		got, err := AllocInRange(uint64(from), uint64(to))
		if !reflect.DeepEqual(err, wantErr) {
			t.Fatalf("AllocInRange() error := %v, want := %v", err, wantErr)
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("AllocInRange() got := %v, want := %v", got, want)
		}
	})

	t.Run("AllocSecond", func(t *testing.T) {
		from := 20000
		to := 20001
		want := []uint64{20001}
		var wantErr error = nil

		helperAlloc(t, uint64(from))

		got, err := AllocInRange(uint64(from), uint64(to))
		if !reflect.DeepEqual(err, wantErr) {
			t.Fatalf("AllocInRange() error := %v, want := %v", err, wantErr)
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("AllocInRange() got := %v, want := %v", got, want)
		}
	})

	t.Run("AllBusy", func(t *testing.T) {
		from := 20000
		to := 20001
		want := make([]uint64, 0)
		var wantErr error = nil

		helperAlloc(t, uint64(from))
		helperAlloc(t, uint64(to))

		got, err := AllocInRange(uint64(from), uint64(to))
		if !reflect.DeepEqual(err, wantErr) {
			t.Fatalf("AllocInRange() error := %v, want := %v", err, wantErr)
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("AllocInRange() got := %#v, want := %#v", got, want)
		}
	})
}

func TestAllocInSlice(t *testing.T) {
	t.Run("AllocFirst", func(t *testing.T) {
		ports := []uint64{20000, 20001}
		want := []uint64{20000, 20001}
		var wantErr error = nil

		got, err := AllocInSlice(ports)
		if !reflect.DeepEqual(err, wantErr) {
			t.Fatalf("AllocInRange() error := %v, want := %v", err, wantErr)
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("AllocInRange() got := %v, want := %v", got, want)
		}
	})

	t.Run("AllocSecond", func(t *testing.T) {
		ports := []uint64{20000, 20001}
		want := []uint64{20001}
		var wantErr error = nil

		helperAlloc(t, ports[0])

		got, err := AllocInSlice(ports)
		if !reflect.DeepEqual(err, wantErr) {
			t.Fatalf("AllocInRange() error := %v, want := %v", err, wantErr)
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("AllocInRange() got := %v, want := %v", got, want)
		}
	})

	t.Run("AllBusy", func(t *testing.T) {
		from := 20000
		to := 20001
		want := make([]uint64, 0)
		var wantErr error = nil

		helperAlloc(t, uint64(from))
		helperAlloc(t, uint64(to))

		got, err := AllocInRange(uint64(from), uint64(to))
		if !reflect.DeepEqual(err, wantErr) {
			t.Fatalf("AllocInRange() error := %v, want := %v", err, wantErr)
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("AllocInRange() got := %#v, want := %#v", got, want)
		}
	})
}

func BenchmarkAllocInSlice(b *testing.B) {
	ports := helperMakePortsSlice(b)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := AllocInSlice(ports); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAllocInRange(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := AllocInRange(0, 65000); err != nil {
			b.Fatal(err)
		}
	}
}

func helperMakePortsSlice(t testing.TB) []uint64 {
	t.Helper()

	ports := make([]uint64, 65000)

	for i := 1; i < 65000; i++ {
		ports[i] = uint64(i)
	}

	return ports
}

func helperAlloc(t testing.TB, port uint64) {
	t.Helper()

	addr, resolveErr := net.ResolveTCPAddr("tcp", ":"+strconv.FormatInt(int64(port), 10))
	if resolveErr != nil {
		t.Errorf("failed to resolve TCP address: %v", resolveErr)
		return
	}

	l, listenErr := net.ListenTCP("tcp", addr)
	if listenErr != nil {
		t.Fatalf("failed to allocate port: %v", listenErr)
	}

	t.Cleanup(func() {
		if err := l.Close(); err != nil {
			t.Logf("failed to close %s", l.Addr())
		}
	})
}
