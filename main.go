package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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
		fmt.Print("\033[0;36mPROXY: \033[0m")
		scanner.Scan()
		command := strings.TrimSpace(scanner.Text())
		switch command {
		case "help":
			fmt.Println("Alle Verfügbaren Commands")
			fmt.Println("help: Show this Menu")
			fmt.Println("clear: Clear the Console")
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

func handleConnection(clientConn net.Conn, minecraftServer string, minecraftPort string) {
	fmt.Println("Verbindung von", clientConn.RemoteAddr().String())
	fmt.Print("\033[0;36mPROXY: \033[0m")

	serverConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", minecraftServer, minecraftPort))
	if err != nil {
		fmt.Println("Error connecting to Minecraft server:", err)
		fmt.Print("\033[0;36mPROXY: \033[0m")
		return
	}
	defer serverConn.Close()
	sendProxyProtocol(clientConn, serverConn)
	go func() {
		_, err := io.Copy(serverConn, clientConn)
		if err != nil {
			fmt.Println("Error forwarding data to Minecraft server:", err)
			fmt.Print("\033[0;36mPROXY: \033[0m")
			clientConn.Close()
		}
	}()

	_, err = io.Copy(clientConn, serverConn)
	if err != nil {
		fmt.Println("Client connection terminated.")
		fmt.Print("\033[0;36mPROXY: \033[0m")
	}
}

func main() {
	port := 25555
	fmt.Println("\u001B[34m  _   _ ______ _______      ________ ")
	fmt.Println("\u001B[34m | \\ | |  ____|  __ \\ \\    / /  ____|")
	fmt.Println("\u001B[34m |  \\| | |__  | |__) \\ \\  / /| |__  ")
	fmt.Println("\u001B[34m | . ` |  __| |  _  / \\ \\/ / |  __| ")
	fmt.Println("\u001B[34m | |\\  | |____| | \\ \\  \\  /  | |____ ")
	fmt.Println("\u001B[34m |_| \\_|______|_|  \\_\\  \\/   |______|")
	fmt.Println("")
	fmt.Println("")
	createProxyDirectoryIfNotExist()
	config, err := ReadConfig()
	minecraftServer := config["host"]
	minecraftPort := config["port"]
	if err != nil {
		log.Fatalf("Error reading configuration: %v", err)
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Printf("Listen to :%d\n", port)
	go readConsoleCommands()

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting client connection:", err)
			continue
		}

		go handleConnection(clientConn, minecraftServer, minecraftPort)
	}
}

func createProxyDirectoryIfNotExist() error {
	proxyDir := "Proxys"
	_, err := os.Stat("Proxys")
	if os.IsNotExist(err) {
		errDir := os.MkdirAll("Proxys", 0755)
		if errDir != nil {
			return errDir
		}
		fmt.Println("Der Ordner Proxy wurde erstellt")
	}
	configFile := filepath.Join(proxyDir, "config.yml")
	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		file, err := os.Create(configFile)
		if err != nil {
			return err
		}
		defer file.Close()
		defaultConfig := []byte(`Proxy01
host: 0.0.0.0
port: 25565
proxyprotocol: false
		`)
		_, err = file.Write(defaultConfig)
		if err != nil {
			return err
		}
		fmt.Println("Config 'config.yml' wurde erstellt. Bitte passe die Config an und starte die Proxy")
		os.Exit(0)
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
