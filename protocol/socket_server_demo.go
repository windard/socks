package protocol

import (
	"fmt"
	"log"
	"net"

	stream "github.com/nknorg/encrypted-stream"
)

func Serve() {
	server, err := net.Listen("tcp", ":6423")
	if err != nil {
		log.Printf("Listen  failed: %v\n", err)
		return
	}

	for {
		client, err := server.Accept()
		if err != nil {
			log.Printf("Accept failed: %v\n", err)
			continue
		}
		remoteAddr := client.RemoteAddr().String()
		remoteNetwork := client.RemoteAddr().Network()
		log.Printf("Connection from: [%s]%s\n", remoteNetwork, remoteAddr)

		// 反向代理服务器
		//go ReverseProxyProcess(client)
		// 正向代理服务器-socks5
		//go Socks5Process(client)
		// 正向代理服务器-socks4
		//Socks4Process(client)
		// 反向加密代码服务器
		go ReverseEncryptProxyProcess(client)
		// 正向解密代理服务器-socks5
		//go Socks5DecryptProcess(client)
	}
}

func echoProcess(client net.Conn) {
	// goroutine panic will panic all
	remoteAddr := client.RemoteAddr().String()
	remoteNetwork := client.RemoteAddr().Network()
	fmt.Printf("Connection from: [%s]%s\n", remoteNetwork, remoteAddr)

	for {
		message := make([]byte, 1024)
		_, err := client.Read(message)
		if err != nil {
			// if client close will raise EOF
			log.Printf("Receive failed: %v\n", err)
			// No need to close
			client.Close()
			return
		}
		log.Printf("【Receive】:%s", message)
		client.Write(append([]byte("【echo】"), message...))
	}
}

func ReverseProxyProcess(client net.Conn) {
	dstConn, err := net.Dial("tcp", "127.0.0.1:5002")
	if err != nil {
		client.Close()
		log.Printf("Connect Internal failed: %v\n", err)
		return
	}

	Socks5Forward(client, dstConn)
}

func ReverseEncryptProxyProcess(client net.Conn) {
	dstConn, err := net.Dial("tcp", "127.0.0.1:6543")
	if err != nil {
		client.Close()
		log.Printf("Connect Internal failed: %v\n", err)
		return
	}

	dstConn, err = stream.NewEncryptedStream(dstConn, &stream.Config{
		Cipher: stream.NewXSalsa20Poly1305Cipher(&SecretKeyByte),
	})
	if err != nil {
		log.Printf("create encrypted stream error:%+v\n", err)
		return
	}
	Socks5Forward(client, dstConn)
}
