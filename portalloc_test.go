package portalloc

import (
	"net"
	"reflect"
	"strconv"
	"testing"
)

func TestAllocateTCP(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		port := 20000
		wantPort := 20000
		var wantErr error = nil

		got, err := PortAlloc(uint64(port))
		if !reflect.DeepEqual(err, wantErr) {
			t.Errorf("PortAlloc() error := %v, want := %v", err, wantErr)
			return
		}

		if got <= 0 || !reflect.DeepEqual(port, wantPort) {
			t.Errorf("PortAlloc() got := %v expected := %v", got, wantPort)
			return
		}
	})

	t.Run("ErrPortIsBusy", func(t *testing.T) {
		port := 30000
		wantErr := ErrPortIsBusy

		helperAlloc(t, port)

		_, err := PortAlloc(uint64(port))
		if !reflect.DeepEqual(err, wantErr) {
			t.Errorf("PortAlloc() error := %v, want := %v", err, wantErr)
			return
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
