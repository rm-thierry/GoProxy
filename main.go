package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var Proxys []ProxyConfig

type ProxyConfig struct {
	Port       int
	Address    string
	TargetPort int
}

func getProxys(folderPath string) ([]ProxyConfig, error) {
	var proxyConfigs []ProxyConfig

	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yml") {
			configFilePath := filepath.Join(folderPath, file.Name())
			proxyConfig, err := readProxyConfigFromFile(configFilePath)
			if err != nil {
				fmt.Printf("Error reading '%s': %v\n", file.Name(), err)
				continue
			}
			proxyConfigs = append(proxyConfigs, proxyConfig)
		}
	}
	return proxyConfigs, nil
}

func readProxyConfigFromFile(filePath string) (ProxyConfig, error) {
	var proxyConfig ProxyConfig

	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return proxyConfig, err
	}

	lines := strings.Split(string(fileData), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) != 2 {
			continue
		}

		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])

		switch key {
		case "Port":
			port, err := strconv.Atoi(value)
			if err != nil {
				return proxyConfig, fmt.Errorf("Falscher Port '%s'", filePath)
			}
			proxyConfig.Port = port
		case "Address":
			proxyConfig.Address = value
		case "TargetPort":
			targetPort, err := strconv.Atoi(value)
			if err != nil {
				return proxyConfig, fmt.Errorf("Falscher target Port '%s'", filePath)
			}
			proxyConfig.TargetPort = targetPort
		}
	}

	if proxyConfig.Port == 0 || proxyConfig.Address == "" || proxyConfig.TargetPort == 0 {
		return proxyConfig, fmt.Errorf("Wrong Config '%s'", filePath)
	}

	return proxyConfig, nil
}

func sendProxyProtocol(clientConn net.Conn, serverOutput io.Writer) {
	sourceAddress := clientConn.RemoteAddr().(*net.TCPAddr)
	destAddress := clientConn.LocalAddr().(*net.TCPAddr)

	proxyProtocolHeader := fmt.Sprintf("PROXY TCP4 %s %s %d %d\r\n",
		sourceAddress.IP.String(), destAddress.IP.String(), sourceAddress.Port, destAddress.Port)

	serverOutput.Write([]byte(proxyProtocolHeader))
}

func readConsoleCommands() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\033[0m", time.TimeOnly, " \033[0;36mPROXY: \033[0m")
		scanner.Scan()
		command := strings.TrimSpace(scanner.Text())
		switch command {
		case "list":
			folderPath := "Proxys"
			proxyConfigs, err := getProxys(folderPath)
			if err != nil {
				fmt.Println("Fehler beim Lesen der Proxy-Konfigurationen:", err)
				return
			}
			var count int = 0
			for _, config := range proxyConfigs {
				count = count + 1
				fmt.Println("Proxy", count, config.Port, config.Address, config.TargetPort)
			}
		case "update":
			updateProxies()
		case "stop":
			os.Exit(0)
		case "clear":
			clearConsole()
		default:
		}
	}
}

func clearConsole() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func handleConnection(clientConn net.Conn, minecraftServer string, minecraftPort int) {

	serverConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", minecraftServer, fmt.Sprint(minecraftPort)))
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Print("\033[0m", time.TimeOnly, " \033[0;36mPROXY: \033[0m")
		return
	}
	defer serverConn.Close()
	sendProxyProtocol(clientConn, serverConn)
	go func() {
		_, err := io.Copy(serverConn, clientConn)
		if err != nil {
			fmt.Println("Error:", err)
			fmt.Print("\033[0m", time.TimeOnly, " \033[0;36mPROXY: \033[0m")
			clientConn.Close()
		}
	}()

	_, err = io.Copy(clientConn, serverConn)
	if err != nil {
	}
}

func startProxyListener(port int, minecraftServer string, minecraftPort int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()
	fmt.Printf("Listening :%d\n", port)
	go readConsoleCommands()

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		go handleConnection(clientConn, minecraftServer, minecraftPort)
	}
}

func updateProxies() {
	folderPath := "Proxys"

	existingPorts := make(map[int]bool)
	for _, proxy := range Proxys {
		existingPorts[proxy.Port] = true
	}

	newProxies, err := getNewProxies(folderPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, proxy := range newProxies {
		if existingPorts[proxy.Port] {
			continue
		}

		fmt.Printf("Neue Proxy - Port: %d, Address: %s, TargetPort: %d\n", proxy.Port, proxy.Address, proxy.TargetPort)
		Proxys = append(newProxies)
		go startProxyListener(proxy.Port, proxy.Address, proxy.TargetPort)
	}
	fmt.Println("Updated Successfully")
}

func getNewProxies(folderPath string) ([]ProxyConfig, error) {
	var newProxies []ProxyConfig

	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yml") {
			configFilePath := filepath.Join(folderPath, file.Name())
			proxyConfig, err := readProxyConfigFromFile(configFilePath)
			if err != nil {
				fmt.Printf("Error '%s': %v\n", file.Name(), err)
				continue
			}
			newProxies = append(newProxies, proxyConfig)
		}
	}
	return newProxies, nil
}

func main() {
	folderPath := "Proxys"
	fmt.Println("\u001B[34m  _   _ ______ _______      ________ ")
	fmt.Println("\u001B[34m | \\ | |  ____|  __ \\ \\    / /  ____|")
	fmt.Println("\u001B[34m |  \\| | |__  | |__) \\ \\  / /| |__  ")
	fmt.Println("\u001B[34m | . ` |  __| |  _  / \\ \\/ / |  __| ")
	fmt.Println("\u001B[34m | |\\  | |____| | \\ \\  \\  /  | |____ ")
	fmt.Println("\u001B[34m |_| \\_|______|_|  \\_\\  \\/   |______|")
	fmt.Println("")
	fmt.Println("")
	createProxyDirectoryIfNotExist()
	proxyConfigs, err := getProxys(folderPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	Proxys = append(proxyConfigs)

	for _, config := range proxyConfigs {
		go startProxyListener(config.Port, config.Address, config.TargetPort)
	}
	readConsoleCommands()
}

func createProxyDirectoryIfNotExist() error {
	_, err := os.Stat("Proxys")
	if os.IsNotExist(err) {
		errDir := os.MkdirAll("Proxys", 0755)
		if errDir != nil {
			return errDir
		}
		fmt.Println("Created Proxy Folder")
	}
	return nil
}

func ReadConfig() (map[string]string, error) {
	config := make(map[string]string)

	configFile := "Proxys/config.yml"
	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		config[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return config, nil
}
