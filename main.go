package main

import (
	"flag"
	"fmt"
	"github.com/alonexy/acps/conf"
	"github.com/alonexy/acps/logger"
	"github.com/alonexy/acps/packets"
	"github.com/alonexy/acps/tests"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// WaitGroupWrapper WaitGroupWrapper
type WaitGroupWrapper struct {
	sync.WaitGroup
}

// Wrap Wrap
func (w *WaitGroupWrapper) Wrap(fn func()) {
	w.Add(1)
	go func() {
		defer w.Done()
		fn()
	}()
}
func init() {
	logger.InitLogging("test_server.log", 0)
}
func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	if (conf.ClientTest) {
		tests.ClienTest()
	} else {
		//ctx, cancel := context.WithCancel(context.Background())
		//
		//wg := &WaitGroupWrapper{}
		signChan := make(chan os.Signal)
		signal.Notify(signChan, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		localAddress, err := net.ResolveTCPAddr("tcp4", conf.Conf.TCP.Bind)
		if (err != nil) {
			panic(err)
		}
		log.Printf("Listen --->>%v",localAddress)
		tcpListener, err := net.ListenTCP("tcp4", localAddress)
		if err != nil {
			panic(err)
		}
		defer func() {
			tcpListener.Close()
		}()

		go func() {
			for s := range signChan {
				switch s {
				case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
					fmt.Println("退出", s)
					os.Exit(0)
				default:
					fmt.Println("other", s)
					os.Exit(0)
				}
			}
		}()
		//声明一个管道用于接收解包的数据
		readerChannel := make(chan string, 1)
		go func() {
			go packets.Reader(readerChannel)
		}()
		for {
			if TcpConn, err := tcpListener.AcceptTCP(); err != nil {
				log.Println("接受连接失败", err)
				continue
			} else {
				go handleAccepts(TcpConn, readerChannel)
			}

		}

		log.Println("Service exit")
	}

}

func handleAccepts(conn *net.TCPConn, readerChannel chan string) {
	log.Printf("%v--->>connect success", conn.RemoteAddr())
	conn.Write(packets.Pong())
	//设置了超时，当一定时间内客户端无请求发送，conn便会自动关闭，
	conn.SetDeadline(time.Now().Add(time.Duration(conf.Conf.TCP.Timeout) * time.Second))
	ch := packets.ServerHeartTimer(conn)
	packets.ServerRetDataHandle(conn, readerChannel)
	ch <- true
	return
}
