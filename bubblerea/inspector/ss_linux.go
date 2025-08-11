package inspector

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	reProc = regexp.MustCompile(`users:\(\("([^"]+)",pid=(\d+)`)
)

func listWithSS(ctx context.Context) ([]PortInfo, error) {
	// ss for TCP and UDP; include all states and process info
	cmdTCP := exec.CommandContext(ctx, "bash", "-lc", "ss -H -tanup || ss -H -tanp")
	cmdUDP := exec.CommandContext(ctx, "bash", "-lc", "ss -H -uanup || ss -H -uanp")

	outTCP, errTCP := cmdTCP.Output()
	outUDP, errUDP := cmdUDP.Output()

	var items []PortInfo
	items = append(items, parseSSLines(string(outTCP))...)
	items = append(items, parseSSLines(string(outUDP))...)

	if len(items) == 0 {
		if errTCP != nil {
			return nil, errTCP
		}
		if errUDP != nil {
			return nil, errUDP
		}
	}
	return items, nil
}

func parseSSLines(s string) []PortInfo {
	scanner := bufio.NewScanner(strings.NewReader(s))
	var out []PortInfo
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" { continue }
		// Expected columns: Netid State Recv-Q Send-Q Local:Port Peer:Port Process
		fields := splitSS(line)
		if len(fields) < 6 { continue }
		proto := fields[0]
		state := fields[1]
		local := fields[len(fields)-3]
		peer := fields[len(fields)-2]
		proc := fields[len(fields)-1]

		localAddr, localPort := splitHostPort(local)
		peerAddr, peerPort := splitHostPort(peer)

		cmdName, pid := parseProcess(proc)
		username := lookupUserLinux(pid)

		out = append(out, PortInfo{
			PID:           pid,
			Command:       cmdName,
			User:          username,
			Protocol:      strings.ToUpper(proto),
			LocalAddress:  localAddr,
			LocalPort:     localPort,
			RemoteAddress: peerAddr,
			RemotePort:    peerPort,
			State:         state,
		})
	}
	return out
}

func splitSS(line string) []string {
	// collapse multiple spaces and split; keep process part intact even if contains spaces
	parts := strings.Fields(line)
	return parts
}

func splitHostPort(s string) (host, port string) {
	if s == "*" || s == "*:*" || s == "*:*\n" { return "*", "*" }
	// IPv6 addresses are like [::]:80 or [::1]:443 or [fe80::1%lo0]:53
	if strings.HasPrefix(s, "[") {
		idx := strings.LastIndex(s, "]:")
		if idx != -1 {
			return s[1:idx], s[idx+2:]
		}
	}
	// Otherwise split by last ':'
	idx := strings.LastIndex(s, ":")
	if idx == -1 { return s, "" }
	return s[:idx], s[idx+1:]
}

func parseProcess(proc string) (cmd string, pid int) {
	m := reProc.FindStringSubmatch(proc)
	if len(m) == 3 {
		cmd = m[1]
		pid, _ = strconv.Atoi(m[2])
	}
	return
}

func lookupUserLinux(pid int) string {
	if pid <= 0 { return "" }
	statusPath := filepath.Join("/proc", strconv.Itoa(pid), "status")
	f, err := os.Open(statusPath)
	if err != nil { return "" }
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Uid:") {
			// Uid:	Real	Effective	Saved	Filesystem
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				uid := fields[1]
				if u, err := user.LookupId(uid); err == nil {
					return u.Username
				}
			}
			break
		}
	}
	return ""
}