package cmd

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	echoCmd = &cobra.Command{
		Use:   "echo [host:port]",
		Short: "echo protocol implement",
		Long: `echo protocol implement
see https://tools.ietf.org/html/rfc862.
`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				listenAddr = defaultAddr
			} else if len(args) == 1 {
				listenAddr = args[0]
			} else if len(args) == 2 {
				listenAddr = strings.Join(args, ":")
			} else if len(args) > 2 {
				fmt.Println("listen addr is invalid.")
				os.Exit(1)
			}

			EchoServe(listenAddr)
		},
	}
)

func EchoServe(listenAddr string) {
	server, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Printf("Listen failed: %v\n", err)
		return
	}
	log.Printf("Start Listen on %s ...", listenAddr)
	for {
		client, err := server.Accept()
		if err != nil {
			log.Printf("Accept failed: %v\n", err)
			continue
		}
		remoteAddr := client.RemoteAddr().String()
		remoteNetwork := client.RemoteAddr().Network()
		log.Printf("Connection from: [%s]%s\n", remoteNetwork, remoteAddr)

		// echo-server
		go EchoProcess(client)
	}
}

func EchoProcess(client net.Conn) {
	// goroutine panic will panic all

	for {
		message := make([]byte, 1024)
		n, err := client.Read(message)
		if err != nil {
			// if client close will raise EOF
			log.Printf("Receive failed: %v\n", err)
			return
		}
		log.Printf("【Receive】:%s", message[:n])
		_, err = client.Write(message)
		if err != nil {
			// if client close will raise EOF
			log.Printf("Send failed: %v\n", err)
			return
		}
	}
}

func init() {
	RootCmd.AddCommand(echoCmd)
}
