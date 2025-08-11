package inspector

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
)

type PortInfo struct {
	PID           int
	Command       string
	User          string
	Protocol      string
	LocalAddress  string
	LocalPort     string
	RemoteAddress string
	RemotePort    string
	State         string
}

// ListPorts returns a slice of PortInfo for busy/listening ports.
func ListPorts(ctx context.Context) ([]PortInfo, error) {
	goos := runtime.GOOS
	if goos == "linux" {
		var items []PortInfo
		var err error
		if hasBinary("ss") {
			items, err = listWithSS(ctx)
			if err == nil && len(items) > 0 {
				return items, nil
			}
		}
		if hasBinary("lsof") {
			return listWithLSOF(ctx)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list ports with ss, and lsof not found: %w", err)
		}
		return nil, errors.New("neither 'ss' nor 'lsof' is installed; install iproute2 (ss) or lsof")
	}
	// macOS and others: prefer lsof
	if hasBinary("lsof") {
		return listWithLSOF(ctx)
	}
	return nil, errors.New("'lsof' not found; please install it to list busy ports")
}

func hasBinary(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}