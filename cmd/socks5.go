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

const defaultAddr = "127.0.0.1:1080"

var (
	listenAddr string
	username   string
	password   string

	socks5Cmd = &cobra.Command{
		Use:   "socks5 [host:port]",
		Short: "socks5 protocol implement",
		Long: `socks5 protocol implement
see https://datatracker.ietf.org/doc/html/rfc1928.`,
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

			if username != "" && password != "" {
				protocol.UserAuthInfo = protocol.Socks5UserAuthInfo{
					Username: []byte(username),
					Password: []byte(password),
				}
			}
			Socks5Serve(listenAddr)
		},
	}
)

func Socks5Serve(listenAddr string) {
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

		if secretKey != "" {
			protocol.SecretKey = secretKey
			protocol.InitSecretKey()
			protocol.Socks5DecryptProcess(client)
		} else {
			go protocol.Socks5Process(client)
		}
	}
}

func init() {
	RootCmd.AddCommand(socks5Cmd)
	socks5Cmd.Flags().StringVarP(&username, "username", "u", "", "socks proxy username.")
	socks5Cmd.Flags().StringVarP(&password, "password", "p", "", "socks proxy password.")
	socks5Cmd.Flags().StringVarP(&secretKey, "secretKey", "s", "", "socks proxy secretKey.")
}
