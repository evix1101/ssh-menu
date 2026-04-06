package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/evix1101/ssh-menu/internal/config"
	"github.com/evix1101/ssh-menu/internal/host"
	"github.com/evix1101/ssh-menu/internal/theme"
	"github.com/evix1101/ssh-menu/internal/ui"
)

func main() {
	verbosePtr := flag.Bool("V", false, "Enable SSH verbose mode (-v flag)")
	groupPtr := flag.String("g", "", "Filter hosts by group")
	listGroupsPtr := flag.Bool("l", false, "List all available groups")
	sshOptsPtr := flag.String("s", "", "Additional SSH options to pass through")
	flag.Parse()

	configPath := sshConfigPath()

	theme.Init(configPath)

	hosts, err := config.ReadConfigFiles(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading SSH config: %s\n", err)
		os.Exit(1)
	}

	if len(hosts) == 0 {
		fmt.Fprintln(os.Stderr, "No menu hosts found in SSH config. Ensure hosts have a '# Menu ...' comment.")
		os.Exit(1)
	}

	hosts, err = host.AssignMenuNumbers(hosts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	hosts = host.ValidateHosts(hosts)

	if *listGroupsPtr {
		listGroups(hosts)
		return
	}

	if *groupPtr != "" {
		hosts = host.HostsForGroup(hosts, *groupPtr)
		if len(hosts) == 0 {
			fmt.Fprintf(os.Stderr, "No hosts found in group '%s'\n", *groupPtr)
			os.Exit(1)
		}
	}

	if args := flag.Args(); len(args) > 0 {
		h := findHost(args[0], hosts)
		if h == nil {
			fmt.Fprintln(os.Stderr, "Host not found.")
			os.Exit(1)
		}
		if err := connectSSH(*h, *verbosePtr, *sshOptsPtr); err != nil {
			fmt.Fprintf(os.Stderr, "Error executing SSH: %v\n", err)
			os.Exit(1)
		}
		return
	}

	m := ui.New(hosts, *verbosePtr, *sshOptsPtr)
	if err := ui.Run(m); err != nil {
		fmt.Fprintf(os.Stderr, "Error running UI: %v\n", err)
		os.Exit(1)
	}

	if m.Selected != nil {
		if err := connectSSH(*m.Selected, *verbosePtr, *sshOptsPtr); err != nil {
			fmt.Fprintf(os.Stderr, "Error executing SSH: %v\n", err)
			os.Exit(1)
		}
	}
}

func sshConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to determine home directory.")
		os.Exit(1)
	}
	return filepath.Join(home, ".ssh", "config")
}

func connectSSH(h host.Host, verbose bool, sshOpts string) error {
	var args []string
	if verbose {
		args = append(args, "-v")
	}
	if sshOpts != "" {
		args = append(args, strings.Fields(sshOpts)...)
	}
	args = append(args, h.ShortName)

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func findHost(input string, hosts []host.Host) *host.Host {
	if num, err := strconv.Atoi(input); err == nil {
		for i, h := range hosts {
			if h.MenuNumber == num {
				return &hosts[i]
			}
		}
		return nil
	}
	for i, h := range hosts {
		if h.ShortName == input || h.LongName == input || h.IP == input {
			return &hosts[i]
		}
	}
	return nil
}

func listGroups(hosts []host.Host) {
	groups := host.GetAllGroups(hosts)
	if len(groups) == 0 {
		fmt.Println("No groups found in SSH config.")
		return
	}
	fmt.Println("Available groups:")
	for _, g := range groups {
		count := len(host.HostsForGroup(hosts, g))
		fmt.Printf("  %s (%d hosts)\n", g, count)
	}
}
