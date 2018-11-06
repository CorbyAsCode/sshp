package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

// Structure to hold host info and output
type hostStruct struct {
	Hostname string
	Output   string
	Error    string
	Command  string
}

// Execute ssh command.  Push results to channel and signal to WaitGroup that we're finished
func sshWorker(user string, sshKeyPath string, cmd string, debug bool) {
	for host := range inputs {
		var output []byte
		start := time.Now()

		client := &sshClient{
			IP:   host.Hostname,
			User: user,
			Port: 22,
			Cert: sshKeyPath,
		}

		if debug {
			fmt.Printf("INFO: sshWorker() starting %s\n", host.Hostname)
		}
		err := client.Connect()
		if err != nil {
			//fmt.Printf("ERROR: %s: %s\n", host.Hostname, err.Error())
			host.Error = err.Error()
			outputs <- host
			continue
		}
		output, err = client.RunCmd(cmd)
		if err != nil {
			//fmt.Printf("ERROR: %s: %s\n", host.Hostname, err.Error())
			host.Error = err.Error()
			outputs <- host
			continue
		}
		client.Close()

		if debug {
			fmt.Printf("INFO: Finished %s in \t%s\n", host.Hostname, time.Since(start))
		}
		host.Output = fmt.Sprintf("%s", output)
		outputs <- host
	}
}

// Read a file and return an array of strings
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// Parse a config file and return strings
func parseConfig(contents []string) (string, string) {
	var (
		user       string
		sshKeyPath string
	)

	for _, line := range contents {
		if strings.HasPrefix(line, "user") {
			user = strings.Trim(strings.Split(line, "=")[1], " ")
		}
		if strings.HasPrefix(line, "sshKeyPath") {
			sshKeyPath = strings.Trim(strings.Split(line, "=")[1], " ")
		}
	}
	return user, sshKeyPath
}

// Check if file exists
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// Return an array of strings from an ini file section
func getIniSection(section string, sections map[string][]string) (value []string, ok bool) {
	if s, ok := sections[section]; ok == true {
		return s, true
	}
	return nil, false
}

// Read ini file and return an array of bytes
func parseIniFile(filename string) ([]byte, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return contents, err
}

// Parse array of bytes from ini file returning a map of the sections
func parseINI(data []byte, lineSep string) map[string][]string {
	var sectionName string
	sections := make(map[string][]string)
	lines := bytes.Split(data, []byte(lineSep))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		size := len(line)
		if size == 0 {
			// Skip blank lines
			continue
		}
		if line[0] == ';' || line[0] == '#' {
			// Skip comments
			continue
		}
		if line[0] == '[' && line[size-1] == ']' {
			// Parse INI-Section
			sectionName = string(line[1 : size-1])
			continue
		}
		val := bytes.TrimSpace(line)
		sections[sectionName] = append(sections[sectionName], string(val))
	}
	return sections
}

// Set up channels.
var inputs = make(chan hostStruct, 5)
var outputs = make(chan hostStruct, 5)

// main function
func main() {

	// Default config file
	defaultConfig := "/etc/sshp.conf"

    // Global vars
	var (
		configLines    []string
		err            error
		hosts          []string
		sshOutputs     hostStruct
		ok             bool
		userConf       *string
		sshKeyPathConf *string
	)
	numWorkers := 10
	iniLineSeparator := "\n"
	haveHostlist, haveHostFile, haveIni := false, false, false

	// Parse CLI flags.
	hostFile := flag.String("f", "", "Host file to use, one hostname per line.")
	hostList := flag.String("l", "", "Comma-separated list of hosts.")
	cmd := flag.String("c", "", "Command to execute over ssh.")
	debug := flag.Bool("d", false, "Print debug messages.")
	sudo := flag.Bool("s", false, "Use sudo to become root on hosts.")
	config := flag.String("config", defaultConfig, "Config file to use.")
	user := flag.String("u", "", "User to login as.")
	sshKeyPath := flag.String("k", "", "Path to ssh key.")
	help := flag.Bool("h", false, "Print help.")
	iniFile := flag.String("iniFile", "", "Path to ini file.")
	iniSection := flag.String("iniSection", "", "ini section to retrieve.")
	flag.Parse()

	// Print usage if the help flag is passed
	if *help {
		flag.Usage()
		os.Exit(0)
	}

    // Logic to handle host lists
	if *hostFile == "" && *hostList == "" && *iniFile != "" && *iniSection != "" {
		haveIni = true
	} else if *hostFile != "" && *hostList == "" && *iniFile == "" && *iniSection == "" {
		haveHostFile = true
	} else if *hostFile == "" && *hostList != "" && *iniFile == "" && *iniSection == "" {
		haveHostlist = true
	} else if *hostFile == "" && *hostList == "" && *iniFile != "" && *iniSection == "" {
		fmt.Printf("ERROR: If using an ini file you must include 'iniSection'!\n")
		flag.Usage()
		os.Exit(1)
	} else if *hostFile == "" && *hostList == "" && *iniFile == "" && *iniSection != "" {
		fmt.Printf("ERROR: If using an ini file you must include 'iniFile'!\n")
		flag.Usage()
		os.Exit(1)
	}

	// Parse ini file if it's passed in
	if haveIni {
		iniContents, err := parseIniFile(*iniFile)
		if err != nil {
			fmt.Printf("ERROR: Could not parse ini file '%s': %s\n", *iniFile, err.Error())
			os.Exit(1)
		}
		hosts, ok = getIniSection(*iniSection, parseINI(iniContents, iniLineSeparator))
		if !ok {
			fmt.Printf("ERROR: iniSection '%s' not found!\n", *iniSection)
			os.Exit(1)
		}
    // Parse host list from CLI
	} else if haveHostlist {
		hosts = strings.Split(*hostList, ",")
	// Parse the given host file
	} else if haveHostFile {
		hosts, err = readLines(*hostFile)
		if err != nil {
			fmt.Printf("ERROR reading file: %s\n", err)
			flag.Usage()
			os.Exit(1)
		}
	// Throw error if no host list, host file, or ini file is passed in
	} else {
		fmt.Printf("ERROR: No hosts were passed in.\n")
		flag.Usage()
		os.Exit(1)
	}

	// Read the config file that was passed in
	if *config != "" {
		if exists(*config) {
			configLines, err = readLines(*config)
			if err != nil {
				fmt.Printf("Error: Could not read config file '%s': %s\n", *config, err.Error())
				flag.Usage()
				os.Exit(1)
			}
			usr, key := parseConfig(configLines)
			userConf = &usr
			sshKeyPathConf = &key
		}
	}

	// CLI params should take precedence over config file params that were just recently parsed
	if *user != "" {
		*userConf = *user
	}
	if *sshKeyPath != "" {
		if exists(*sshKeyPath) {
			*sshKeyPathConf = *sshKeyPath
		} else {
			fmt.Printf("ERROR: sshKeyPath '%s' is not readable or does not exist.\n", *sshKeyPath)
			flag.Usage()
			os.Exit(1)
		}
	// Check if the key path from the config file exists
	} else {
		if !exists(*sshKeyPathConf) {
			fmt.Printf("ERROR: sshKeyPath '%s' is not readable or does not exist.\n", *sshKeyPathConf)
			flag.Usage()
			os.Exit(1)
		}
	}

	// Prepend the sudo options if requested
	if *sudo {
		*cmd = fmt.Sprintf("sudo su - root -c '%s'", *cmd)
	}

	if *debug {
		fmt.Printf("Number of hosts to process: %d\n", len(hosts))
	}
	
    // Create ssh workers.
	for i := 1; i <= numWorkers; i++ {
		go sshWorker(*userConf, *sshKeyPathConf, *cmd, *debug)
	}

	// Push hosts onto inputs channel.
	// Close the channel to signal that this channel is finished.
	go func() {
		for _, host := range hosts {
			var hostSt hostStruct
			hostSt.Hostname = host
			inputs <- hostSt
		}
		close(inputs)
	}()

	// Print output as each goroutine finishes
	for a := 1; a <= len(hosts); a++ {
		sshOutputs = <-outputs
		if sshOutputs.Error == "" {
			fmt.Printf("%s: %s", sshOutputs.Hostname, sshOutputs.Output)
		} else {
			fmt.Printf("%s: %s\n", sshOutputs.Hostname, sshOutputs.Error)
		}
	}
}
