package main

import (
	"io"
	"fmt"
	"log"
	"net"
	"os"
	"bufio"
	"bytes"
	"strings"

	proxyproto "github.com/exosite/proxyprotov2"
)

func handleConn(conn net.Conn) {
	silence := false
	connInfo := fmt.Sprintf("%s->%s", conn.RemoteAddr().String(), conn.LocalAddr().String())
	defer func() {
		conn.Close()
		if !silence {
			log.Printf("[%s] Connection closed.", connInfo)
		}
	}()
	connBuf := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	proxyInfo, bytesToWrite, err := proxyproto.HandleProxy(connBuf)
	if err != nil {
		if err == io.EOF {
			// EOF?  Just return.  Screw it.
			silence = true
			return
		}
		log.Printf("[%s] Failed to handle proxy protocol: %s", connInfo, err.Error())
		return
	}
	if proxyInfo != nil {
		for _, tlv := range proxyInfo.TLVs {
			// log.Printf("[%s] TLV 0x%x: %#v", connInfo, tlv.Type, tlv.Value)
			tlsInfo, isTls := tlv.(*proxyproto.TlsTLV)
			if isTls {
				log.Printf("[%s] TLS Version: %s", connInfo, tlsInfo.Version())
				log.Printf("[%s] CN: %s", connInfo, tlsInfo.CN())
				log.Printf("[%s] SNI: %s", connInfo, tlsInfo.SNI())
				if tlsInfo.Certs != nil {
					certs, err := tlsInfo.Certs()
					if err != nil {
						log.Printf("[%s] Failed to parse certificates: %s", err.Error())
					} else {
						for i, cert := range certs {
							log.Printf("[%s] Certificate %d: %s", connInfo, i, cert.Subject.CommonName)
						}
					}
				}
				fp := tlsInfo.Fingerprint()
				if fp != nil {
					var shabuf bytes.Buffer
					for _, part := range fp {
						shabuf.WriteString(fmt.Sprintf("%x", part))
					}
					log.Printf("[%s] Cert SHA1: %s", connInfo, shabuf.String())
				}
			}
		}
	}
	connBuf.Flush()
	connBuf = nil
	mw := io.MultiWriter(conn, os.Stdout)
	if bytesToWrite != nil {
		_, err = mw.Write(bytesToWrite)
		if err != nil {
			log.Printf("[%s] Failed to write back: %s", connInfo, err.Error())
			return
		}
	}
	bytesCopied, err := io.Copy(mw, conn)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "connection reset by peer") {
			silence = true
		} else {
			log.Printf("[%s] Error while echoing input: %s", connInfo, err.Error())
		}
	}
	if bytesCopied == 0 {
		silence = true
	} else {
		log.Printf("[%s] Copied %d bytes", connInfo, bytesCopied)
	}
	//connBuf.Flush()
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %s", err.Error())
			continue
		}

		go handleConn(conn)
	}
}
