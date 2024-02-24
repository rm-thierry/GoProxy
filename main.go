package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
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
			fmt.Println("/help: Show this Menu")
			fmt.Println("/clear: Clear the Console")
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
	serverConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", minecraftServer, minecraftPort))
	if err != nil {
		fmt.Println("Error connecting to Minecraft server:", err)
		return
	}
	defer serverConn.Close()
	sendProxyProtocol(clientConn, serverConn)
	go func() {
		_, err := io.Copy(serverConn, clientConn)
		if err != nil {
			fmt.Println("Error forwarding data to Minecraft server:", err)
		}
	}()

	_, err = io.Copy(clientConn, serverConn)
}

func main() {
	minecraftServer := "78.46.208.71"
	minecraftPort := "25565"
	port := 25555
	fmt.Println("\u001B[34m  _   _ ______ _______      ________ ")
	fmt.Println("\u001B[34m | \\ | |  ____|  __ \\ \\    / /  ____|")
	fmt.Println("\u001B[34m |  \\| | |__  | |__) \\ \\  / /| |__  ")
	fmt.Println("\u001B[34m | . ` |  __| |  _  / \\ \\/ / |  __| ")
	fmt.Println("\u001B[34m | |\\  | |____| | \\ \\  \\  /  | |____ ")
	fmt.Println("\u001B[34m |_| \\_|______|_|  \\_\\  \\/   |______|")
	fmt.Println("")
	fmt.Println("")
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("\033[0;36mPROXY: \033[0m Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Printf("Listen to :%d\n", port)
	go readConsoleCommands()

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			fmt.Println("\033[0;36mPROXY: \033[0m Error accepting client connection:", err)
			fmt.Print("\033[0;36mPROXY: \033[0m")
			continue
		}

		go handleConnection(clientConn, minecraftServer, minecraftPort)
		fmt.Println("Ein Client ist gejoint ", clientConn.RemoteAddr().String())
		fmt.Print("\033[0;36mPROXY: \033[0m")
	}
}
