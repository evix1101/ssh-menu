package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/evix1101/ssh-menu/internal"
)

func main() {
	// Add flags
	detailedPtr := flag.Bool("d", false, "Show detailed connection information in the UI")
	verbosePtr := flag.Bool("V", false, "Enable SSH verbose mode (-v flag)")
	groupPtr := flag.String("g", "", "Filter hosts by group")
	listGroupsPtr := flag.Bool("l", false, "List all available groups")

	// Add a flag for SSH options pass-through
	sshOptsPtr := flag.String("s", "", "Additional SSH options to pass through (e.g. \"-s '-A -J jumphost'\")")

	// Parse flags
	flag.Parse()

	detailed := *detailedPtr
	verbose := *verbosePtr
	group := *groupPtr
	listGroups := *listGroupsPtr
	sshOpts := *sshOptsPtr

	// Get non-flag arguments
	args := flag.Args()

	// Determine the home directory (check HOME then USERPROFILE)
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		fmt.Println("Unable to determine home directory.")
		os.Exit(1)
	}

	// Build the SSH config path in a portable way
	configPath := filepath.Join(home, ".ssh", "config")

	// Initialize the UI style system with color configuration
	internal.InitStyles(configPath)

	// Read all config files (main + config.d)
	hosts, err := internal.ReadConfigFiles(configPath)
	if err != nil {
		fmt.Printf("Error reading SSH config: %s\n", err)
		os.Exit(1)
	}

	if len(hosts) == 0 {
		fmt.Println("No menu hosts found in SSH config. Ensure hosts have a '# Menu ...' comment.")
		os.Exit(1)
	}

	// Assign and validate menu numbers
	hosts = internal.AssignMenuNumbers(hosts)

	// If -l flag is set, just list the groups and exit
	if listGroups {
		listAvailableGroups(hosts)
		return
	}

	// If -g flag is set, filter hosts by the specified group
	if group != "" {
		hosts = filterHostsByGroup(hosts, group)
		if len(hosts) == 0 {
			fmt.Printf("No hosts found in group '%s'\n", group)
			os.Exit(1)
		}
	}

	// Process command-line argument if provided
	if len(args) > 0 {
		handleDirectHostSelection(args[0], hosts, verbose, detailed, sshOpts)
		return
	}

	// Create and run the terminal UI
	startTerminalUI(hosts, verbose, detailed, sshOpts)
}

// connectSSH executes the SSH command
func connectSSH(h internal.Host, verbose bool, sshOpts string) error {
	// Prepare SSH command arguments
	args := []string{}

	// Add verbose flag to SSH if in verbose mode
	if verbose {
		args = append(args, "-v")
	}

	// Process additional SSH options if provided
	if sshOpts != "" {
		// Split the options string into individual arguments
		additionalArgs := strings.Fields(sshOpts)
		args = append(args, additionalArgs...)
	}

	// Add the host - SSH will read all connection details from its config
	args = append(args, h.ShortName)

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// handleDirectHostSelection handles direct host selection from command line arguments
func handleDirectHostSelection(input string, hosts []internal.Host, verbose bool, detailed bool, sshOpts string) {
	var selected *internal.Host

	// If numeric, search by menu number
	if num, err := strconv.Atoi(input); err == nil {
		for i, h := range hosts {
			if h.MenuNumber == num {
				selected = &hosts[i]
				break
			}
		}
		if selected == nil {
			fmt.Println("Invalid selection.")
			os.Exit(1)
		}
	} else {
		// Otherwise, search by shortname, longname, or IP
		for i, h := range hosts {
			if h.ShortName == input || h.LongName == input || h.IP == input {
				selected = &hosts[i]
				break
			}
		}
		if selected == nil {
			fmt.Println("Host not found.")
			os.Exit(1)
		}
	}

	// Connect directly to the selected host
	if err := connectSSH(*selected, verbose, sshOpts); err != nil {
		fmt.Printf("Error executing ssh: %v\n", err)
		os.Exit(1)
	}
}

// startTerminalUI creates and runs the terminal UI
func startTerminalUI(hosts []internal.Host, verbose bool, detailed bool, sshOpts string) {
	// Setup the UI components
	ui := internal.SetupUI(hosts, verbose, detailed, sshOpts)

	// Run the UI
	if err := internal.RunUI(ui); err != nil {
		fmt.Printf("Error running UI: %v\n", err)
		os.Exit(1)
	}

	// If a host was selected, show connection view and connect to it
	if ui.Selected != nil {
		// Connect to the selected host directly
		if err := connectSSH(*ui.Selected, verbose, sshOpts); err != nil {
			fmt.Printf("Error executing SSH: %v\n", err)
			os.Exit(1)
		}
	}
}

// listAvailableGroups prints all available groups and exits
func listAvailableGroups(hosts []internal.Host) {
	groups := internal.GetAllGroups(hosts)

	if len(groups) == 0 {
		fmt.Println("No groups found in SSH config.")
		return
	}

	fmt.Println("Available groups:")
	for _, g := range groups {
		// Count hosts in this group
		count := 0
		if g == "Ungrouped" {
			for _, h := range hosts {
				if len(h.Groups) == 0 {
					count++
				}
			}
		} else {
			for _, h := range hosts {
				for _, hg := range h.Groups {
					if hg == g {
						count++
						break
					}
				}
			}
		}

		fmt.Printf("  %s (%d hosts)\n", g, count)
	}
}

// filterHostsByGroup returns only hosts that belong to the specified group
func filterHostsByGroup(hosts []internal.Host, group string) []internal.Host {
	var filtered []internal.Host

	for _, h := range hosts {
		if group == "Ungrouped" {
			if len(h.Groups) == 0 {
				filtered = append(filtered, h)
			}
		} else {
			for _, g := range h.Groups {
				if g == group {
					filtered = append(filtered, h)
					break
				}
			}
		}
	}

	return filtered
}
