package protocol

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	stream "github.com/nknorg/encrypted-stream"
)

type Socks5UserAuthInfo struct {
	Username []byte
	Password []byte
}

var UserAuthInfo = Socks5UserAuthInfo{
	Username: []byte("YOUR_PROXY_LOGIN"),
	Password: []byte("YOUR_PROXY_PASSWORD"),
}

var SecretKey string
var SecretKeyByte [32]byte

func InitSecretKey() {
	b := md5.Sum([]byte(SecretKey))
	copy(SecretKeyByte[:16], b[:])
	copy(SecretKeyByte[16:], b[:])

	//hex.EncodeToString(b[:])
	//SecretKeyByte = []byte(fmt.Sprintf("%x", md5.Sum([]byte(SecretKey))))
}

func Socks5Connect(client net.Conn) (dstConn net.Conn, err error) {
	buf := make([]byte, 256)

	// 读取请求命令
	n, err := client.Read(buf[:4])
	if n != 4 || err != nil {
		return nil, errors.New("reading header: " + err.Error())
	}

	ver, cmd, _, atyp := buf[0], buf[1], buf[2], buf[3]
	if ver != 5 || cmd != 1 {
		return nil, errors.New("invalid ver/cmd")
	}

	// 读取请求地址
	addr := ""
	switch atyp {
	case 1:
		log.Printf("client addr type:IPv4\n")
		n, err = io.ReadFull(client, buf[:4])
		if n != 4 || err != nil {
			return nil, errors.New("reading IPv4 addr error:" + err.Error())
		}
		addr = net.IP(buf[:4]).String()
	case 3:
		log.Printf("client addr type:URL\n")
		n, err = io.ReadFull(client, buf[:1])
		if n != 1 || err != nil {
			return nil, errors.New("reading addr Len error:" + err.Error())
		}
		addrLen := int(buf[0])
		n, err = io.ReadFull(client, buf[:addrLen])
		if n != addrLen || err != nil {
			return nil, errors.New("reading addr URL error:" + err.Error())
		}
		addr = string(buf[:addrLen])
	case 4:
		log.Printf("client addr type:IPv6\n")
		n, err = io.ReadFull(client, buf[:16])
		if n != 16 || err != nil {
			return nil, errors.New("reading IPv4 addr error:" + err.Error())
		}
		addr = net.IP(buf[:16]).String()
	default:
		log.Printf("client addr type:Unknown\n")
		return nil, errors.New("invalid atyp")
	}

	n, err = io.ReadFull(client, buf[:2])
	if n != 2 || err != nil {
		return nil, errors.New("reading port error:" + err.Error())
	}
	port := binary.BigEndian.Uint16(buf[:2])

	network := ""
	dst := ""
	switch atyp {
	case 1:
		network = "tcp"
		dst = fmt.Sprintf("%s:%d", addr, port)
	case 3:
		network = "tcp"
		dst = fmt.Sprintf("%s:%d", addr, port)
	case 4:
		network = "tcp6"
		dst = fmt.Sprintf("[%s]:%d", addr, port)
	}

	log.Printf("client addr:%s\n", dst)
	dstConn, err = net.Dial(network, dst)
	if err != nil {
		_, err2 := client.Write([]byte{0x05, 0x01})
		if err2 != nil {
			return nil, errors.New("writing error:" + err2.Error())
		}
		return nil, errors.New("dial dst error:" + err.Error())
	}

	// 返回连接状态
	n, err = client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	if err != nil {
		return nil, errors.New("writing error:" + err.Error())
	}

	return dstConn, nil
}

func Socks5Auth(client net.Conn) (err error) {
	buf := make([]byte, 256)

	// 读取 VER 和 NMETHODS
	//n, err := io.ReadFull(client, buf[:2])
	n, err := client.Read(buf[:2])
	if n != 2 || err != nil {
		return errors.New("reading header: " + err.Error())
	}

	ver, nMethods := int(buf[0]), int(buf[1])
	if ver != 5 {
		return errors.New("invalid version")
	}

	// 读取 METHODS 列表
	n, err = io.ReadFull(client, buf[:nMethods])
	if n != nMethods {
		return errors.New("reading methods:" + err.Error())
	}

	log.Printf("client accept methods:%+v\n", buf[:nMethods])

	if bytes.IndexByte(buf[:nMethods], 0x02) < 0 {
		// 如果没有用户名密码，则返回无需认证
		n, err = client.Write([]byte{0x05, 0x00})
		if n != 2 || err != nil {
			return errors.New("writing error:" + err.Error())
		}
	} else {
		// 返回用户名密码认证
		n, err = client.Write([]byte{0x05, 0x02})
		if n != 2 || err != nil {
			return errors.New("writing error:" + err.Error())
		}
		return Socks5UserAuth(client)
	}

	return nil
}

func Socks5UserAuth(client net.Conn) (err error) {
	buf := make([]byte, 256)
	var username, password []byte

	// 读取用户名
	n, err := client.Read(buf[:2])
	if n != 2 || err != nil {
		return errors.New("reading header: " + err.Error())
	}

	ver, nUsername := int(buf[0]), int(buf[1])
	if ver != 1 {
		return errors.New("invalid version")
	}

	n, err = client.Read(buf[:nUsername])
	if n != nUsername {
		return errors.New("reading username error:" + err.Error())
	}

	username = make([]byte, nUsername)
	copy(username, buf[:nUsername])

	// 读取密码
	n, err = client.Read(buf[:1])
	if n != 1 || err != nil {
		return errors.New("reading header: " + err.Error())
	}

	nPassword := int(buf[0])
	n, err = client.Read(buf[:nPassword])
	if n != nPassword {
		return errors.New("reading password error:" + err.Error())
	}

	password = make([]byte, nPassword)
	copy(password, buf[:nPassword])

	log.Printf("user auth username:%s, password:%s\n", username, password)

	if bytes.Equal(username, UserAuthInfo.Username) && bytes.Equal(password, UserAuthInfo.Password) {
		// 认证成功
		n, err = client.Write([]byte{0x01, 0x00})
		if n != 2 || err != nil {
			return errors.New("writing error:" + err.Error())
		}
	} else {
		// 认证失败
		n, err = client.Write([]byte{0x01, 0x01})
		if n != 2 || err != nil {
			return errors.New("writing error:" + err.Error())
		}
		return errors.New("user auth fail")
	}
	return nil
}

func Socks5Forward(client, target net.Conn) {
	forward := func(src, dst net.Conn) {
		defer src.Close()
		defer dst.Close()
		io.Copy(src, dst)
	}
	go forward(client, target)
	go forward(target, client)
}

func Socks4Connect(client net.Conn) (dstConn net.Conn, err error) {
	buf := make([]byte, 256)

	// 读取请求命令
	n, err := client.Read(buf[:2])
	if n != 2 || err != nil {
		return nil, errors.New("reading header: " + err.Error())
	}
	vn, cd := buf[0], buf[1]
	if vn != 4 || cd != 1 {
		return nil, errors.New("invalid vn/cd")
	}

	// 读取请求端口
	n, err = io.ReadFull(client, buf[:2])
	if n != 2 || err != nil {
		return nil, errors.New("reading port error:" + err.Error())
	}
	port := binary.BigEndian.Uint16(buf[:2])

	// 读取请求地址
	n, err = io.ReadFull(client, buf[:4])
	if n != 4 || err != nil {
		return nil, errors.New("reading IPv4 addr error:" + err.Error())
	}
	addr := net.IP(buf[:4]).String()

	n, err = client.Read(buf)
	if err != nil {
		return nil, errors.New("reading userid error: " + err.Error())
	}
	log.Printf("socks4 user id:%+v\n", buf[:n])

	network := "tcp"
	dst := fmt.Sprintf("%s:%d", addr, port)
	log.Printf("remote addr:%s\n", dst)
	dstConn, err = net.Dial(network, dst)
	if err != nil {
		return nil, errors.New("dial dst error:" + err.Error())
	}

	// 返回连接状态
	n, err = client.Write([]byte{0x00, 0x5a, 0, 0, 0, 0, 0, 0})
	if err != nil {
		dstConn.Close()
		return nil, errors.New("writing error:" + err.Error())
	}

	return dstConn, nil
}

func Socks4Process(client net.Conn) {
	log.Printf("start socks4 auth handle")

	target, err := Socks4Connect(client)
	if err != nil {
		log.Printf("connect error:%+v\n", err)
		client.Close()
		return
	}

	log.Printf("start socks4 forward handle")
	Socks5Forward(client, target)
}

func Socks5DecryptProcess(client net.Conn) {
	log.Printf("start socks5 auth handle")

	client, err := stream.NewEncryptedStream(client, &stream.Config{
		Cipher: stream.NewXSalsa20Poly1305Cipher(&SecretKeyByte),
	})
	if err != nil {
		log.Printf("create encrypted stream error:%+v\n", err)
		return
	}

	if err := Socks5Auth(client); err != nil {
		log.Printf("auth error:%+v\n", err)
		client.Close()
		return
	}

	log.Printf("start socks5 connect handle")
	target, err := Socks5Connect(client)
	if err != nil {
		log.Printf("connect error:%+v\n", err)
		client.Close()
		return
	}

	log.Printf("start socks5 forward handle")
	Socks5Forward(client, target)
}

func Socks5Process(client net.Conn) {
	log.Printf("start socks5 auth handle")

	if err := Socks5Auth(client); err != nil {
		log.Printf("auth error:%+v\n", err)
		client.Close()
		return
	}

	log.Printf("start socks5 connect handle")
	target, err := Socks5Connect(client)
	if err != nil {
		log.Printf("connect error:%+v\n", err)
		client.Close()
		return
	}

	log.Printf("start socks5 forward handle")
	Socks5Forward(client, target)
}
