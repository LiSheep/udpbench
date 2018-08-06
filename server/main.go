package main

import (
	"flag"
	"net"
	"fmt"
	"time"
	"github.com/tenchlee/udpbench"
)

var (
	listenAddr = flag.String("listenAddr", "0.0.0.0:12345", "server address (ip:port)")
	gos = flag.Int("gos", 1, "max go routine num")
)

var listener *net.UDPConn

func recv_udp() {
	data := make([]byte, 1500)
	for {
		n, addr, err := listener.ReadFrom(data)
		if err != nil {
			fmt.Println("ReadFrom", err)
			return
		}
		ok, _, _, length := udpbench.Check_package(data, n)
		if !ok {
			fmt.Println("check data fail")
			continue
		}
		snd_data := data[:length+14]
		listener.WriteTo(snd_data, addr)
	}
}

func main() {

	var err error
	flag.Parse()
	addr,err := net.ResolveUDPAddr("udp",*listenAddr)
	if err != nil  {
		fmt.Println("ResolveUDPAddr", err)
		return
	}

	listener, err = net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("ListenUDP", err)
		return
	}
	defer listener.Close()

	i := *gos
	for i > 0 {
		i--
		go recv_udp()
	}
	for {
		time.Sleep(time.Hour)
	}
}