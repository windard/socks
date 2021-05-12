package cmd

import (
	"crypto/md5"
	"fmt"
	"log"
	"net"
	"os"

	stream "github.com/nknorg/encrypted-stream"

	"github.com/spf13/cobra"
	protocol "github.com/windard/socks/protocol"
)

var (
	clientAddr    string
	targetAddr    string
	role          string
	secretKey     string
	secretKeyByte [32]byte

	proxyCmd = &cobra.Command{
		Use:   "proxy [clientAddr] [targetAddr]",
		Short: "reverse proxy protocol implement",
		Long: `reverse proxy protocol implement
TCP socket forward, you can encrypt connection with secretKey.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 2 {
				clientAddr = args[0]
				targetAddr = args[1]
			} else {
				fmt.Println("Args is invalid.")
				os.Exit(1)
			}

			if secretKey != "" {
				b := md5.Sum([]byte(secretKey))
				copy(secretKeyByte[:16], b[:])
				copy(secretKeyByte[16:], b[:])
			}

			if role != "" && secretKey == "" {
				fmt.Printf("Please specify a secret.\n")
				os.Exit(1)
			}
			ProxyServe(clientAddr, targetAddr, role)
		},
	}
)

func ProxyServe(clientAddr, targetAddr, role string) {
	server, err := net.Listen("tcp", targetAddr)
	if err != nil {
		log.Printf("Listen failed: %v\n", err)
		return
	}
	log.Printf("Start Listen on %s ...", targetAddr)

	for {
		client, err := server.Accept()
		if err != nil {
			log.Printf("Accept failed: %v\n", err)
			continue
		}
		remoteAddr := client.RemoteAddr().String()
		remoteNetwork := client.RemoteAddr().Network()
		log.Printf("Connection from: [%s]%s\n", remoteNetwork, remoteAddr)

		if role == "" {
			// 反向代理服务器
			go ReverseProxyProcess(client, clientAddr)
		} else if role == "local" {
			// 本地加密代理服务器
			go ReverseLocalEncryptProxyProcess(client, clientAddr)
		} else {
			// 远程解密代理服务器
			go ReverseRemoteDecryptProxyProcess(client, clientAddr)
		}
	}
}

func ReverseProxyProcess(client net.Conn, clientAddr string) {
	dstConn, err := net.Dial("tcp", clientAddr)
	if err != nil {
		client.Close()
		log.Printf("Connect Internal failed: %v\n", err)
		return
	}

	protocol.Socks5Forward(client, dstConn)
}

func ReverseLocalEncryptProxyProcess(client net.Conn, clientAddr string) {
	dstConn, err := net.Dial("tcp", clientAddr)
	if err != nil {
		client.Close()
		log.Printf("Connect Internal failed: %v\n", err)
		return
	}

	dstConn, err = stream.NewEncryptedStream(dstConn, &stream.Config{
		Cipher: stream.NewXSalsa20Poly1305Cipher(&secretKeyByte),
	})
	if err != nil {
		log.Printf("create encrypted stream error:%+v\n", err)
		return
	}
	protocol.Socks5Forward(client, dstConn)
}

func ReverseRemoteDecryptProxyProcess(client net.Conn, clientAddr string) {
	dstConn, err := net.Dial("tcp", clientAddr)
	if err != nil {
		client.Close()
		log.Printf("Connect Internal failed: %v\n", err)
		return
	}

	client, err = stream.NewEncryptedStream(client, &stream.Config{
		Cipher: stream.NewXSalsa20Poly1305Cipher(&secretKeyByte),
	})
	if err != nil {
		log.Printf("create encrypted stream error:%+v\n", err)
		return
	}
	protocol.Socks5Forward(client, dstConn)
}

func init() {
	RootCmd.AddCommand(proxyCmd)
	proxyCmd.Flags().StringVarP(&role, "role", "r", "", "proxy role.(local|remote)")
	proxyCmd.Flags().StringVarP(&secretKey, "secretKey", "s", "", "proxy secretKey.")
}
