package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"strings"
)

type ForwardMode int

const (
	ForwardModeRtl ForwardMode = 0
	ForwardModeLtr ForwardMode = 1
)

// return whether the forwarding is successful or not
func forward(client *ssh.Client, rAddr *net.TCPAddr, lAddr *net.TCPAddr, mode ForwardMode) bool {

	switch mode {
	case ForwardModeRtl:
		if !checkRemotePortAvailable(client, rAddr) {
			log.Println("remote port is occupied")
			return false
		}
		tl, err := client.ListenTCP(rAddr)

		if err != nil {
			log.Printf("Failed to forward port: %s%s --> %s%s, because %s", RemoteMark, rAddr, LocalMark, lAddr, err.Error())
			return false
		}
		log.Printf("Forward port: %s%s --> %s%s", RemoteMark, rAddr, LocalMark, lAddr)

		go func() {
			defer func(tl net.Listener) {
				err := tl.Close()
				if err != nil {
					log.Println("unable to close listener: ", err)
				}
			}(tl)
			failedCount := 0
			for {
				conn, err := tl.Accept()
				if err != nil {
					failedCount++
					log.Printf("unable to accept incoming connection: %v", err)
					if failedCount > 10 {
						log.Printf("too many failed attempts, closing listener")
						err := client.Close()
						if err != nil {
							return
						}
						return
					}
					continue
				}
				log.Printf("Accepted incoming connection: %s%s --> %s%s --> %s%s", RemoteMark, conn.RemoteAddr(), RemoteMark, conn.LocalAddr(), LocalMark, lAddr)
				go copyConnToLocalAddr(conn, lAddr)
			}
		}()
	case ForwardModeLtr:
		tl, err := net.ListenTCP("tcp", lAddr)

		if err != nil {
			log.Printf("Failed to forward port: %s%s --> %s%s, because %s", RemoteMark, rAddr, LocalMark, lAddr, err.Error())
		}
		log.Printf("Forward port: %s%s --> %s%s", LocalMark, lAddr, RemoteMark, rAddr)
		go func() {
			defer func(tl net.Listener) {
				err := tl.Close()
				if err != nil {
					log.Println("unable to close listener: ", err)
				}
			}(tl)
			failedCount := 0
			for {
				conn, err := tl.Accept()
				if err != nil {
					failedCount++
					log.Printf("unable to accept incoming connection: %v", err)
					if failedCount > 10 {
						log.Printf("too many failed attempts, closing listener")
						err := client.Close()
						if err != nil {
							return
						}
						return
					}
					continue
				}
				log.Printf("Accepted incoming connection: %s%s --> %s%s --> %s%s", LocalMark, conn.RemoteAddr(), LocalMark, conn.LocalAddr(), RemoteMark, rAddr)
				go copyConnToRemoteAddr(client, conn, rAddr)
			}
		}()

	}

	return true
}

// check the remote port is occupied or not
func checkRemotePortAvailable(client *ssh.Client, rAddr *net.TCPAddr) bool {
	if rAddr.Port == 0 {
		return true
	}

	session, err := client.NewSession()
	if err != nil {
		log.Println("unable to create session during check port: ", err)
	}
	cmd := fmt.Sprintf("netstat -ltn|grep ':%d '", rAddr.Port)
	//cmd := "netstat -ltn"
	_, err = session.CombinedOutput(cmd)
	if err != nil {
		// if no port match, the status code is 1
		if strings.Contains(err.Error(), "exited with status 1") {
			return true
		}
		log.Fatal("unable to run netstat: ", err)
	}
	return true
}

func copyBetween(srcConn net.Conn, dstConn net.Conn) {
	go func() {
		size, err := io.Copy(dstConn, srcConn)
		if err != nil {
			log.Println("unable to copy: ", err)
		}
		log.Printf("Req-Copied %d bytes", size)
	}()
	go func() {
		size, err := io.Copy(srcConn, dstConn)
		if err != nil {
			log.Println("unable to copy: ", err)
		}
		log.Printf("Res-Copied %d bytes", size)
	}()
}

/**
 * DialTCP to rAddr as dstConn on remote host, and copy data between conn and dstConn
 */
func copyConnToRemoteAddr(client *ssh.Client, conn net.Conn, rAddr *net.TCPAddr) {
	if client == nil {
		log.Println("client is nil")
		return
	}

	dstConn, err := client.DialTCP("tcp", nil, rAddr)
	if err != nil {
		log.Println("unable to dial: ", err)
		return
	}
	copyBetween(conn, dstConn)
}

/**
 * DialTCP to lAddr as dstConn,and copy data between conn and dstConn
 */
func copyConnToLocalAddr(conn net.Conn, lAddr *net.TCPAddr) {
	dstConn, err := net.DialTCP("tcp", nil, lAddr)
	if err != nil {
		log.Println("unable to dial: ", err)
		return
	}
	copyBetween(conn, dstConn)
}
