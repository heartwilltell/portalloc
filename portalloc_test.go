package portalloc

import (
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

		helperAlloc(t, port)

		_, err := Alloc(uint64(port))
		if !reflect.DeepEqual(err, wantErr) {
			t.Fatalf("Alloc() error := %v, want := %v", err, wantErr)
		}
	})
}

func TestAllocInRange(t *testing.T) {
	t.Run("First", func(t *testing.T) {
		from := 20000
		to := 20001
		want := from
		var wantErr error = nil

		got, err := AllocInRange(uint64(from), uint64(to))
		if !reflect.DeepEqual(err, wantErr) {
			t.Fatalf("AllocInRange() error := %v, want := %v", err, wantErr)
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("AllocInRange() got := %v, want := %v", got, want)
		}
	})

	t.Run("Second", func(t *testing.T) {
		from := 20000
		to := 20001
		want := to
		var wantErr error = nil

		helperAlloc(t, from)

		got, err := AllocInRange(uint64(from), uint64(to))
		if !reflect.DeepEqual(err, wantErr) {
			t.Fatalf("AllocInRange() error := %v, want := %v", err, wantErr)
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("AllocInRange() got := %v, want := %v", got, want)
		}
	})

	t.Run("ErrPortIsBusy", func(t *testing.T) {
		from := 20000
		to := 20001
		wantErr := ErrPortIsBusy

		helperAlloc(t, from)
		helperAlloc(t, to)

		_, err := AllocInRange(uint64(from), uint64(to))
		if !reflect.DeepEqual(err, wantErr) {
			t.Fatalf("AllocInRange() error := %v, want := %v", err, wantErr)
		}
	})
}

func helperAlloc(t *testing.T, port int) {
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
