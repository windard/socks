package protocol

import (
	"crypto/md5"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/chacha20"
)

type Chacha20Stream struct {
	key     []byte
	encoder *chacha20.Cipher
	decoder *chacha20.Cipher

	conn net.Conn
}

type CipherStream interface {
	Close() error
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
}

var secretKey []byte

func NewChacha20Stream(key []byte, conn net.Conn) (*Chacha20Stream, error) {
	s := &Chacha20Stream{
		key:  key,
		conn: conn,
	}

	var err error

	nonce := make([]byte, chacha20.NonceSizeX)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	s.encoder, err = chacha20.NewUnauthenticatedCipher(s.key, nonce)
	if err != nil {
		return nil, err
	}

	if n, err := s.conn.Write(nonce); err != nil || n != len(nonce) {
		return nil, errors.New("write nonce failed:" + err.Error())
	}
	log.Printf("Send nonce:%v", nonce)
	return s, nil
}

func (s *Chacha20Stream) Read(p []byte) (int, error) {
	if s.decoder == nil {
		nonce := make([]byte, chacha20.NonceSizeX)
		if n, err := io.ReadAtLeast(s.conn, nonce, len(nonce)); err != nil || n != len(nonce) {
			return n, errors.New("can't read nonce from stream:" + err.Error())
		}
		log.Printf("Receive nonce:%v", nonce)
		decoder, err := chacha20.NewUnauthenticatedCipher(s.key, nonce)
		if err != nil {
			return 0, errors.New("generate decoder failed:" + err.Error())
		}

		s.decoder = decoder
	}

	n, err := s.conn.Read(p)
	if err != nil || n == 0 {
		return n, err
	}

	dst := make([]byte, n)
	pn := p[:n]
	s.decoder.XORKeyStream(dst, pn)
	copy(pn, dst)
	return n, nil
}

func (s *Chacha20Stream) Write(p []byte) (int, error) {
	dst := make([]byte, len(p))
	s.encoder.XORKeyStream(dst, p)

	return s.conn.Write(dst)
}

func (s *Chacha20Stream) Close() error {
	return s.conn.Close()
}

func RelayProcess(client net.Conn, remoteAddr string, role string) {
	target, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		client.Close()
		log.Printf("Relay failed: %v\n", err)
		return
	}

	var src, dst CipherStream
	if role == "local" {
		log.Printf("local role is:%s", role)
		src = client                                    // source request client
		dst, err = NewChacha20Stream(secretKey, target) // local relay server
	} else {
		log.Printf("remote role is:%s", role)
		src, err = NewChacha20Stream(secretKey, client) // remote relay server
		dst = target                                    // source request target
	}

	if err != nil {
		client.Close()
		target.Close()
		log.Printf("new stream error:" + err.Error())
		return
	}

	log.Printf("receive from:%s relay to:%s", client.RemoteAddr().String(), target.RemoteAddr().String())
	StreamForward(src, dst)
}

func StreamForward(client, target CipherStream) {
	forward := func(src, dst CipherStream) {
		defer src.Close()
		defer dst.Close()
		_, err := io.Copy(src, dst)
		if err != nil {
			log.Printf("copy error:" + err.Error())
			return
		}
	}
	go forward(client, target)
	go forward(target, client)
}

func RelayServe() {
	//rand.Seed(time.Now().UnixNano())
	listenAddr := flag.String("listenAddr", "127.0.0.1:9080", "listen addr")
	remoteAddr := flag.String("remoteAddr", "127.0.0.1:9081", "remote addr")
	role := flag.String("role", "local", "role: local or remote")
	secret := flag.String("secret", "", "secret key")
	flag.Parse()

	if *secret == "" {
		fmt.Printf("Please specify a secret.\n")
		return
	}

	secretKey = []byte(fmt.Sprintf("%x", md5.Sum([]byte(*secret))))
	fmt.Printf("%s: [%s] -> [%s]\n", *role, *listenAddr, *remoteAddr)

	server, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		fmt.Printf("Listen failed: %v\n", err)
		return
	}

	for {
		client, err := server.Accept()
		if err != nil {
			fmt.Printf("Accept failed: %v\n", err)
			continue
		}

		clientAddr := client.RemoteAddr().String()
		clientNetwork := client.RemoteAddr().Network()
		log.Printf("Connection from: [%s]%s\n", clientNetwork, clientAddr)

		go RelayProcess(client, *remoteAddr, *role)
	}
}
