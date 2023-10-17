package xnotify

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Process info
// Used by eventMetadata to transform the corresponding process information structure
type ProcInfo struct {
	Name  string
	Path  string
	Cmds  []string
	Pid   int
	EUid  int
	EUser string
	PPid  int
	PName string
	PPath string
}

// GetProcessName return pid name and ppid,if no error
func GetProcessName(pid int) (string, int, error) {
	var (
		name string
		ppid int
	)

	filename := fmt.Sprintf("%s%d%s", "/proc/", pid, "/status")
	dat, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("[FaNotify convert2ProcInfo] read file %s error:%s\n", filename, err)
		return "", 0, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(dat)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Name:\t") {
			name = line[6:]
			if i := strings.IndexAny(name, "/: ;,"); i > 0 {
				name = name[:i]
			}

			//the max len of cmd name in /stat is 16(include \r). if it's 15, it's maybe a short cut name
			//if it is exe, it's a symlink, not a real one.
			if name == "exec" || len(name) == maxStatCmdLen {
				if cmds, err := ReadCmdLine(pid); err == nil && len(cmds) > 0 && cmds[0] != "" {
					name = filepath.Base(cmds[0])
				}
			}
		} else if strings.HasPrefix(line, "PPid:\t") {
			ppid, _ = strconv.Atoi(line[6:])
			return name, ppid, nil
		}
	}
	return "", 0, fmt.Errorf("Process name not found in status\n")
}

// ReadCmdLine return cmdline info
func ReadCmdLine(pid int) ([]string, error) {
	var cmds []string

	file, err := os.Open(fmt.Sprintf("%s/%v/cmdline", "/proc", pid))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		cmds = strings.Split(string(scanner.Text()), "\x00")
		for i, t := range cmds {
			cmds[i] = string([]byte(t))
		}
		break
	}

	return cmds, nil
}

func GetFilePath(pid int) (string, error) {
	filename := fmt.Sprintf("%s%d%s", "/proc/", pid, "/exe")
	path, err := os.Readlink(filename)
	if err != nil {
		return "", err
	} else {
		// Sometime we see a path like "/bin/busybox (deleted)"
		// need to return the path part in order to detect the fast process like echo
		if strings.Contains(path, " (deleted)") {
			return strings.TrimRight(path, " (deleted)"), nil
		}

		return path, nil
	}
}
