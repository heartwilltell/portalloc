package portalloc

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"syscall"
)

const (
	// ErrPortIsBusy represents an error which indicates that port is busy.
	ErrPortIsBusy Error = "port is busy"
)

// Error represents package level errors.
type Error string

func (e Error) Error() string { return string(e) }

// PortAlloc tries to allocate given port.
// Returns ErrPortIsBusy in case the port has already been allocated.
func PortAlloc(port uint64) (p int, aErr error) {
	addr, resolveErr := net.ResolveTCPAddr("tcp", ":"+strconv.FormatUint(port, 10))
	if resolveErr != nil {
		return 0, fmt.Errorf("failed to resolve TCP address: %w", resolveErr)
	}

	l, listenErr := net.ListenTCP("tcp", addr)
	if listenErr != nil {
		if errors.Is(listenErr, syscall.EADDRINUSE) {
			return 0, ErrPortIsBusy
		}

		return 0, fmt.Errorf("failed to allocate TCP port: %w", listenErr)
	}

	defer func(l *net.TCPListener) {
		if err := l.Close(); err != nil {
			aErr = errors.Join(aErr, err)
		}
	}(l)

	tcpAddr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("failed to convert address to net.TCPAddr")
	}

	return tcpAddr.Port, nil
}
