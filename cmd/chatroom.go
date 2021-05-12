package cmd

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var (
	clientMap = sync.Map{}
	chatCmd   = &cobra.Command{
		Use:   "chat [host:port]",
		Short: "chatroom implement",
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

			ChatroomServer(listenAddr)
		},
	}
)

func ChatroomServer(listenAddr string) {
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
		clientMap.Store(remoteAddr, client)

		go ChatroomProcess(client)
	}
}

func ChatroomProcess(client net.Conn) {
	// goroutine panic will panic all
	remoteAddr := client.RemoteAddr().String()

	for {
		message := make([]byte, 1024)
		n, err := client.Read(message)
		if err != nil {
			// if client close will raise EOF
			log.Printf("Receive failed: %v\n", err)
			clientMap.Delete(remoteAddr)
			return
		}
		log.Printf("【Receive】【%s】:%s", remoteAddr, message[:n])
		clientMap.Range(func(key, value interface{}) bool {
			target, ok := value.(net.Conn)
			if !ok {
				// return false will break circulation
				return false
			}
			if client == target {
				return true
			}

			_, err = target.Write(append([]byte(fmt.Sprintf("【%s】", remoteAddr)), message...))
			if err != nil {
				log.Printf("Send To 【%s】 failed: %v\n", target.RemoteAddr().String(), err)
				return false
			}
			return true
		})
	}
}

func init() {
	RootCmd.AddCommand(chatCmd)
}
