package cmd

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"
	protocol "github.com/windard/socks/protocol"
)

var (
	socks4Cmd = &cobra.Command{
		Use:   "socks4 [host:port]",
		Short: "socks4 protocol implement",
		Long: `socks4 protocol implement
see https://www.openssh.com/txt/socks4.protocol`,
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

			Socks4Serve(listenAddr)
		},
	}
)

func Socks4Serve(listenAddr string) {
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

		// 正向代理服务器-socks4
		go protocol.Socks4Process(client)
	}
}

func init() {
	RootCmd.AddCommand(socks4Cmd)
}
