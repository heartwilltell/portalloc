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

	// ErrInvalidPortRange represents an error which indicates that port range is invalid.
	ErrInvalidPortRange Error = "invalid port range"
)

// Compilation time check for interface implementation.
var _ error = (Error)("")

// Error represents package level errors.
type Error string

func (e Error) Error() string { return string(e) }

// Alloc tries to allocate given port.
// Returns ErrPortIsBusy in case the port has already been allocated.
func Alloc(port uint64) (p uint64, aErr error) {
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

	return uint64(tcpAddr.Port), nil
}

// AllocInSlice tries to allocate each port in the given slice of ports.
// Returns a list of free ports.
func AllocInSlice(ports []uint64) (freePorts []uint64, allocErr error) {
	freePorts = make([]uint64, 0, len(ports))

	for _, p := range ports {
		port, err := Alloc(p)
		if err != nil && !errors.Is(err, ErrPortIsBusy) {
			return nil, err
		}

		// Lets try next port.
		if errors.Is(err, ErrPortIsBusy) {
			continue
		}

		freePorts = append(freePorts, port)
	}

	return freePorts, nil
}

// AllocInRange tries to allocate each port in the given range of ports.
// Returns a list of free ports.
func AllocInRange(from, to uint64) (freePorts []uint64, allocErr error) {
	if from > to {
		return nil, fmt.Errorf("%w: to can't be lower than from", ErrInvalidPortRange)
	}

	freePorts = make([]uint64, 0, to-from)

	for p := from; p <= to; p++ {
		port, err := Alloc(p)
		if err != nil && !errors.Is(err, ErrPortIsBusy) {
			return nil, err
		}

		// Lets try next port.
		if errors.Is(err, ErrPortIsBusy) {
			continue
		}

		freePorts = append(freePorts, port)
	}

	return freePorts, nil
}
