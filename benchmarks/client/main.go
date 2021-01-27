package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"time"
)

var addr = flag.String("a", "localhost:1834", "address")
var num = flag.Int("c", 1, "connection number")
var timeOut = flag.Int("t", 2, "timeout second")
var msgLen = flag.Int("m", 4096, "message length")

var msg []byte

func main() {
	flag.Parse()
	msg = make([]byte, *msgLen)
	rand.Read(msg)

	// startC 和 closeC 都是空接口通道
	startC := make(chan interface{})
	closeC := make(chan interface{})
	// result 是结果数据通道，数量为 *num 个
	result := make(chan int64, *num)

	for i := 0; i < *num; i++ {
		conn, err := net.Dial("tcp", *addr)
		if err != nil {
			panic(err)
		}
		go handler(conn, startC, closeC, result)
	}

	// start
	close(startC)
	// 进行休眠
	time.Sleep(time.Duration(*timeOut) * time.Second)
	// stop
	close(closeC)

	var totalMessagesRead int64
	for i := 0; i < *num; i++ {
		// 接收全部的包个数，因为知道包固定大小，所以可以这么个统计法
		totalMessagesRead += <-result
	}
	// 进行统计，1MiB = 1024 * 1204B
	fmt.Println(totalMessagesRead/int64(*timeOut*1024*1024), " MiB/s throughput")
}

// handler：处理方法
func handler(conn net.Conn, startC chan interface{}, closeC chan interface{}, result chan int64) {
	var count int64
	buf := make([]byte, 2*(*msgLen))
	<-startC

	_, e := conn.Write(msg)
	if e != nil {
		fmt.Println("Error to send message because of ", e.Error())
	}
	// 死循环进行统计
	for {
		select {
		case <-closeC:
			// 接收到 closeC 时，Client 进行关闭，则将剩余所有 count 写入通道 result
			result <- count
			conn.Close()
			return
		default:
			n, err := conn.Read(buf)
			if n > 0 {
				count += int64(n)
			}
			// 发生错误时
			if err != nil {
				fmt.Print("Error to read message because of ", err)
				result <- count
				conn.Close()
				return
			}
			// 读取 buf 中所有数据
			_, err = conn.Write(buf[:n])
			if err != nil {
				fmt.Println("Error to send message because of ", e.Error())
			}
		}
	}
}
