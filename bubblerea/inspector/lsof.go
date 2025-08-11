package inspector

import (
	"bufio"
	"context"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
)

// listWithLSOF parses lsof machine-readable output
func listWithLSOF(ctx context.Context) ([]PortInfo, error) {
	// -nP: numeric hosts and ports; -i: internet; -F: machine-readable fields
	// Fields: p(pid) c(command) u(uid) P(protocol) n(name like ip:port->ip:port) T(TCP info e.g., TST=LISTEN)
	cmd := exec.CommandContext(ctx, "bash", "-lc", "lsof -nP -i -F pcPuPnT")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseLSOF(string(out)), nil
}

func parseLSOF(s string) []PortInfo {
	scanner := bufio.NewScanner(strings.NewReader(s))
	var (
		currentPID    int
		currentCmd    string
		currentUID    string
		currentUser   string
		currentProto  string
		currentState  string
		items         []PortInfo
	)

	flushUser := func() {
		if currentUser == "" && currentUID != "" {
			if u, err := user.LookupId(currentUID); err == nil {
				currentUser = u.Username
			}
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" { continue }
		// each record is field-tagged; first char is the tag
		tag := line[:1]
		val := line[1:]
		switch tag {
		case "p":
			pid, _ := strconv.Atoi(val)
			currentPID = pid
			currentCmd = ""
			currentUID = ""
			currentUser = ""
			currentProto = ""
			currentState = ""
		case "c":
			currentCmd = val
		case "u":
			currentUID = val
			currentUser = ""
		case "P":
			currentProto = strings.ToUpper(val)
		case "T":
			// e.g., TST=LISTEN
			if strings.HasPrefix(val, "ST=") || strings.HasPrefix(val, "ST=") {
				// some lsof variants use TST=, guard both
			}
			if strings.HasPrefix(val, "TST=") {
				currentState = strings.TrimPrefix(val, "TST=")
			}
			if strings.HasPrefix(val, "ST=") {
				currentState = strings.TrimPrefix(val, "ST=")
			}
		case "n":
			flushUser()
			localAddr, localPort, remoteAddr, remotePort := splitLSOFName(val)
			items = append(items, PortInfo{
				PID:           currentPID,
				Command:       currentCmd,
				User:          currentUser,
				Protocol:      currentProto,
				LocalAddress:  localAddr,
				LocalPort:     localPort,
				RemoteAddress: remoteAddr,
				RemotePort:    remotePort,
				State:         currentState,
			})
		}
	}
	return items
}

func splitLSOFName(n string) (la, lp, ra, rp string) {
	// examples:
	// 127.0.0.1:8000 (LISTEN)
	// 127.0.0.1:8000->127.0.0.1:56789
	// [::1]:53
	if i := strings.Index(n, " (LISTEN)"); i != -1 {
		n = n[:i]
	}
	if strings.Contains(n, "->") {
		parts := strings.SplitN(n, "->", 2)
		la, lp = splitHostPort(parts[0])
		ra, rp = splitHostPort(parts[1])
		return
	}
	la, lp = splitHostPort(n)
	ra, rp = "", ""
	return
}