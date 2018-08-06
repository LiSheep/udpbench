package main

import (
	"gopkg.in/fatih/pool.v2"
	"net"
	"net/http"
	_ "net/http/pprof"
	"flag"
	"runtime"
	"fmt"
	"os"
	"time"
	"github.com/tenchlee/udpbench"
	"github.com/willf/bitset"
	"sync"
)

var (
	serverAddr = flag.String("serverAddr", "192.168.4.223:12345", "server address (ip:port)")
	connections = flag.Int("conn", 1, "concurrent connection")
	cpuprofile = flag.String("prof", "", "write cpu profile to file")
	packageSize = flag.Int("size", 100, "per package size")
	qps = flag.Int("qps", 50, "request per second")
	loopTime = flag.Int("time", 100, "loop times")
	check = flag.Bool("check", true, "check package")
	duration = flag.Int("duration", 20, "send duration (ms)")
)
var conn_pool pool.Pool
var data []byte
var fin_cnt int

func print_ttls(ttls [101]int) {
	fmt.Printf("acks: ")
	for i, cnt := range(ttls) {
		if cnt > 0 {
			fmt.Printf("<%d:%d ", i*10, cnt)
		}
	}
	fmt.Println("")
}

func send_udp() {
	conn, err := conn_pool.Get()
	if err != nil {
		fmt.Println(err)
		return
	}
	var mutex sync.Mutex

	go func() {
		var sets bitset.BitSet
		avg := 0
		min := 1000
		max := 0
		total := 0
		first_ttl := 0
		total_cnt := 0
		var ttls [101]int
		for {
			var data = make([]byte, 1500)
			conn.SetReadDeadline(time.Now().Add(time.Second * 5))
			n, err := conn.Read(data)
			if err != nil {
				if total_cnt > 0 {
					avg = total/total_cnt
				}
				fmt.Printf("%d recv addr:%v avg:%d min:%d max:%d total:%d first_ttl:%d loss:%d%% \n",
					fin_cnt, conn.LocalAddr(), avg, min, max, total_cnt, first_ttl, (*loopTime-total_cnt)*100/(*loopTime))
				print_ttls(ttls)
				return
			}
			ok, id, ts, _ := udpbench.Check_package(data, n)
			if !ok {
				fmt.Println("data format error")
				continue
			}
			total_cnt++
			ttl := udpbench.Iclock() - ts
			ittl := int(ttl)
			if id != 0 {
				total += ittl
				if ittl < min {
					min = ittl
				}
				if ittl > max {
					max = ittl
				}
			} else {
				first_ttl = ittl
			}
			if ittl < 1000 {
				ttls[ittl/10]++
			} else {
				ttls[100]++
			}
			sets.Set(uint(id))
			mutex.Lock()
			if sets.Count() == uint(*loopTime) {
				avg = total/(*loopTime)
				fin_cnt++
				fmt.Printf("%d recv addr:%v avg:%d min:%d max:%d total:%d first_ttl:%d \n",
					fin_cnt, conn.LocalAddr(), avg, min, max, total_cnt, first_ttl)
				print_ttls(ttls)
				if fin_cnt == *connections {
					fmt.Printf("all %d connections finish successs\n", *connections)
				}

				break
			}
			mutex.Unlock()
		}
	}()
	first_pkg := true
	i := 0
	for i < *loopTime {
		conn.Write(udpbench.Encode_package(uint32(i), data))
		if first_pkg {
			time.Sleep(time.Second)
			first_pkg = false
		} else {
			time.Sleep(time.Millisecond * time.Duration(*duration))
		}
		i++
	}
	conn.Close()

}


func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	// if *cpuprofile != "" {
	// 	f, err := os.Create(*cpuprofile)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		return
	// 	}
	// 	fmt.Println("start pro in", *cpuprofile)
	// 	pprof.StartCPUProfile(f)

	// 	defer pprof.StopCPUProfile()
	// }

	go func() {
        http.ListenAndServe("localhost:8080", nil)
    }()

	data = make([]byte, *packageSize)
	for i := 0; i < *packageSize; i++ {
		data[i] = 'a'
	}
	fmt.Printf("server address: %s, connections: %d, packageSize: %d, qps %d, loopTime %d\n",
		*serverAddr, *connections, *packageSize, *qps, *loopTime)

	conn_cb := func() (net.Conn, error) { return net.Dial("udp", *serverAddr) }
	var err error
	conn_pool, err = pool.NewChannelPool(*connections, *connections*2, conn_cb)
	if err != nil {
		fmt.Println(err)
		return
	}

	i := *connections
	for i > 0 {
		i--
		go send_udp()
		time.Sleep(10*time.Millisecond)
	}

	for {
		time.Sleep(time.Hour)
	}
}