package main

import (
	"flag"
	"fmt"
	"github.com/alonexy/acps/conf"
	"github.com/alonexy/acps/packets"
	"github.com/alonexy/acps/tests"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init()  {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	if conf.Conf.LogName != "" {
		file := "./logs/" + conf.Conf.LogName
		logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
		if err != nil {
			panic(err)
		}
		log.SetOutput(logFile) // 将文件设置为log输出的文件
	}
	log.SetPrefix("[GoSocket ==>] ")
	log.SetFlags(log.LstdFlags | log.Ldate | log.Lmicroseconds | log.Lshortfile)
}
func main() {
	if (conf.ClientTest) {
		tests.ClienTest()
	} else {

		signChan := make(chan os.Signal)
		signal.Notify(signChan, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		localAddress, err := net.ResolveTCPAddr("tcp4", conf.Conf.TCP.Bind)
		if (err != nil) {
			panic(err)
		}
		log.Printf("TCP Listen Port --->>%v",localAddress)
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
					fmt.Println("Quit[1]", s)
					os.Exit(1)
				default:
					fmt.Println("Quit[2]", s)
					os.Exit(2)
				}
			}
		}()
		readerChannel := make(chan packets.ReaderMsgType, 100)
		go func() {
			go packets.ServerReader(readerChannel)
		}()
		for {
			if TcpConn, err := tcpListener.AcceptTCP(); err != nil {
				log.Println("AcceptFail:", err)
				continue
			} else {
				go handleAccepts(TcpConn, readerChannel)
			}

		}

		log.Println("Service exit")
	}

}

func handleAccepts(conn *net.TCPConn, readerChannel chan packets.ReaderMsgType) {
	log.Printf("%v--->>connect success", conn.RemoteAddr())
	conn.Write(packets.Pong())
	conn.SetDeadline(time.Now().Add(time.Duration(conf.Conf.TCP.Timeout) * time.Second))
	ch := packets.ServerHeartTimer(conn)
	packets.ServerRetDataHandle(conn, readerChannel)
	ch <- true
	return
}
