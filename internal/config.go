package internal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Constants for environment variable names
const (
	// Color configuration environment variables
	EnvColorBackground = "SSH_MENU_COLOR_BACKGROUND"
	EnvColorForeground = "SSH_MENU_COLOR_FOREGROUND"
	EnvColorBorder     = "SSH_MENU_COLOR_BORDER"
	EnvColorSelected   = "SSH_MENU_COLOR_SELECTED"
	EnvColorAccent     = "SSH_MENU_COLOR_ACCENT"
	EnvColorDimmed     = "SSH_MENU_COLOR_DIMMED"
)

// ReadConfigFiles reads all SSH config files (main + config.d)
// It returns a slice of Host objects representing all host entries
func ReadConfigFiles(configPath string) ([]Host, error) {
	// First read the main config file
	mainHosts, err := readConfigFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading main config file: %w", err)
	}

	// Check for config.d directory
	configDirPath := filepath.Join(filepath.Dir(configPath), "config.d")
	dirInfo, err := os.Stat(configDirPath)

	// If config.d doesn't exist or isn't a directory, just return the main hosts
	if os.IsNotExist(err) || (err == nil && !dirInfo.IsDir()) {
		return mainHosts, nil
	} else if err != nil {
		return nil, fmt.Errorf("error checking config.d directory: %w", err)
	}

	// Read all files in the config.d directory
	files, err := os.ReadDir(configDirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config.d directory: %w", err)
	}

	// Combine all hosts from all config files
	allHosts := mainHosts
	for _, file := range files {
		// Skip directories and hidden files
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}

		filePath := filepath.Join(configDirPath, file.Name())
		additionalHosts, err := readConfigFile(filePath)
		if err != nil {
			fmt.Printf("Warning: Error reading config file %s: %v\n", filePath, err)
			continue
		}

		allHosts = append(allHosts, additionalHosts...)
	}

	return allHosts, nil
}

// configParser holds the regex patterns and methods for parsing SSH config
type configParser struct {
	reHost                *regexp.Regexp
	reHostname            *regexp.Regexp
	reUser                *regexp.Regexp
	rePort                *regexp.Regexp
	reIdentity            *regexp.Regexp
	reMenu                *regexp.Regexp
	reIP                  *regexp.Regexp
	reGroup               *regexp.Regexp
	reConnTimeout         *regexp.Regexp
	reServerAliveInterval *regexp.Regexp
	reServerAliveCountMax *regexp.Regexp
}

// newConfigParser creates a new config parser with compiled regex patterns
func newConfigParser() *configParser {
	return &configParser{
		reHost:                regexp.MustCompile(`^Host\s+(.+)$`),
		reHostname:            regexp.MustCompile(`^Hostname\s+(.+)$`),
		reUser:                regexp.MustCompile(`^User\s+(.+)$`),
		rePort:                regexp.MustCompile(`^Port\s+(\d+)$`),
		reIdentity:            regexp.MustCompile(`^IdentityFile\s+(.+)$`),
		reMenu:                regexp.MustCompile(`^#\s*Menu(?:\s+(\d+))?:\s*(.+)$`),
		reIP:                  regexp.MustCompile(`^#\s*IP:\s*(.+)$`),
		reGroup:               regexp.MustCompile(`^#\s*Group:\s*(.+)$`),
		reConnTimeout:         regexp.MustCompile(`^ConnectTimeout\s+(\d+)$`),
		reServerAliveInterval: regexp.MustCompile(`^ServerAliveInterval\s+(\d+)$`),
		reServerAliveCountMax: regexp.MustCompile(`^ServerAliveCountMax\s+(\d+)$`),
	}
}

// parseLine processes a single line from the config file
func (p *configParser) parseLine(line string, current *Host, hosts *[]Host) error {
	if m := p.reHost.FindStringSubmatch(line); m != nil {
		return p.handleHostLine(m[1], current, hosts)
	}
	if m := p.reHostname.FindStringSubmatch(line); m != nil {
		current.LongName = m[1]
		return nil
	}
	if m := p.reUser.FindStringSubmatch(line); m != nil {
		current.User = m[1]
		return nil
	}
	if m := p.rePort.FindStringSubmatch(line); m != nil {
		current.Port = m[1]
		return nil
	}
	if m := p.reIdentity.FindStringSubmatch(line); m != nil {
		current.IdentityFile = m[1]
		return nil
	}
	if m := p.reMenu.FindStringSubmatch(line); m != nil {
		return p.handleMenuLine(m, current)
	}
	if m := p.reIP.FindStringSubmatch(line); m != nil {
		current.IP = m[1]
		return nil
	}
	if m := p.reGroup.FindStringSubmatch(line); m != nil {
		return p.handleGroupLine(m[1], current)
	}
	if m := p.reConnTimeout.FindStringSubmatch(line); m != nil {
		if timeout, err := strconv.Atoi(m[1]); err == nil {
			current.ConnectTimeout = timeout
		}
		return nil
	}
	if m := p.reServerAliveInterval.FindStringSubmatch(line); m != nil {
		if interval, err := strconv.Atoi(m[1]); err == nil {
			current.ServerAliveInterval = interval
		}
		return nil
	}
	if m := p.reServerAliveCountMax.FindStringSubmatch(line); m != nil {
		if count, err := strconv.Atoi(m[1]); err == nil {
			current.ServerAliveCountMax = count
		}
		return nil
	}
	return nil
}

// handleHostLine processes a Host line
func (p *configParser) handleHostLine(hostName string, current *Host, hosts *[]Host) error {
	if current.ShortName != "" && current.DescText != "" {
		*hosts = append(*hosts, *current)
	}
	*current = Host{
		ShortName:           hostName,
		LongName:            hostName,
		User:                "root",
		Port:                "22",
		Groups:              []string{},
		ConnectTimeout:      0,
		ServerAliveInterval: 0,
		ServerAliveCountMax: 0,
	}
	return nil
}

// handleMenuLine processes a Menu comment line
func (p *configParser) handleMenuLine(matches []string, current *Host) error {
	if matches[1] != "" {
		num, err := strconv.Atoi(matches[1])
		if err != nil {
			return fmt.Errorf("invalid menu number: %s", matches[1])
		}
		current.MenuNumber = num
	} else {
		current.MenuNumber = 0
	}
	current.DescText = matches[2]
	return nil
}

// handleGroupLine processes a Group comment line
func (p *configParser) handleGroupLine(groupValue string, current *Host) error {
	groupValue = strings.TrimSpace(groupValue)
	if !contains(current.Groups, groupValue) {
		current.Groups = append(current.Groups, groupValue)
	}
	return nil
}

// readConfigFile reads a single SSH config file and extracts host entries
// It returns hosts found in the file
func readConfigFile(configPath string) ([]Host, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	parser := newConfigParser()
	var hosts []Host
	current := Host{
		User:                "root",
		Port:                "22",
		Groups:              []string{},
		ConnectTimeout:      0,
		ServerAliveInterval: 0,
		ServerAliveCountMax: 0,
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if err := parser.parseLine(line, &current, &hosts); err != nil {
			return nil, err
		}
	}

	// Append the last host if valid
	if current.ShortName != "" && current.DescText != "" {
		hosts = append(hosts, current)
	}

	return hosts, nil
}

// contains checks if a string is present in a slice
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// ColorConfig holds the color configuration for the UI
type ColorConfig struct {
	Background string // Background color
	Foreground string // Foreground/text color
	Border     string // Border color
	Selected   string // Selected item color
	Accent     string // Accent color (used for titles, headers)
	Dimmed     string // Dimmed color (used for comments, less important text)
}

// getColorSettings reads color settings from environment variables and config file
// Returns a ColorConfig with the color settings to use for the UI
// colorConfigParser holds regex patterns for color configuration
type colorConfigParser struct {
	reColorBackground *regexp.Regexp
	reColorForeground *regexp.Regexp
	reColorBorder     *regexp.Regexp
	reColorSelected   *regexp.Regexp
	reColorAccent     *regexp.Regexp
	reColorDimmed     *regexp.Regexp
}

// newColorConfigParser creates a new color config parser
func newColorConfigParser() *colorConfigParser {
	return &colorConfigParser{
		reColorBackground: regexp.MustCompile(`^#\s*ColorBackground:\s*(.+)$`),
		reColorForeground: regexp.MustCompile(`^#\s*ColorForeground:\s*(.+)$`),
		reColorBorder:     regexp.MustCompile(`^#\s*ColorBorder:\s*(.+)$`),
		reColorSelected:   regexp.MustCompile(`^#\s*ColorSelected:\s*(.+)$`),
		reColorAccent:     regexp.MustCompile(`^#\s*ColorAccent:\s*(.+)$`),
		reColorDimmed:     regexp.MustCompile(`^#\s*ColorDimmed:\s*(.+)$`),
	}
}

// applyEnvColors applies environment variable colors to the config
func applyEnvColors(config *ColorConfig) {
	envMap := map[string]*string{
		EnvColorBackground: &config.Background,
		EnvColorForeground: &config.Foreground,
		EnvColorBorder:     &config.Border,
		EnvColorSelected:   &config.Selected,
		EnvColorAccent:     &config.Accent,
		EnvColorDimmed:     &config.Dimmed,
	}

	for envVar, field := range envMap {
		if envColor := os.Getenv(envVar); envColor != "" {
			*field = envColor
		}
	}
}

// parseColorLine processes a single line for color configuration
func (p *colorConfigParser) parseColorLine(line string, config *ColorConfig) {
	tests := []struct {
		re     *regexp.Regexp
		envKey string
		field  *string
	}{
		{p.reColorBackground, EnvColorBackground, &config.Background},
		{p.reColorForeground, EnvColorForeground, &config.Foreground},
		{p.reColorBorder, EnvColorBorder, &config.Border},
		{p.reColorSelected, EnvColorSelected, &config.Selected},
		{p.reColorAccent, EnvColorAccent, &config.Accent},
		{p.reColorDimmed, EnvColorDimmed, &config.Dimmed},
	}

	for _, test := range tests {
		if m := test.re.FindStringSubmatch(line); m != nil && os.Getenv(test.envKey) == "" {
			*test.field = m[1]
		}
	}
}

func getColorSettings(configPath string) ColorConfig {
	// Start with default color scheme (catppuccin-mocha inspired)
	colorConfig := ColorConfig{
		Background: "#1e1e2e", // Dark blue/purple
		Foreground: "#cdd6f4", // Light gray
		Border:     "#9399b2", // Medium gray
		Selected:   "#a6e3a1", // Green
		Accent:     "#89dceb", // Cyan
		Dimmed:     "#585b70", // Dark gray
	}

	// Apply environment variables (highest priority)
	applyEnvColors(&colorConfig)

	// Check SSH config for defaults (only applies if env vars aren't set)
	file, err := os.Open(configPath)
	if err != nil {
		return colorConfig // Return current defaults if can't open config
	}
	defer file.Close()

	parser := newColorConfigParser()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parser.parseColorLine(line, &colorConfig)
	}

	return colorConfig
}
