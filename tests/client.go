package tests

import (
	"github.com/alonexy/acps/conf"
	"github.com/alonexy/acps/packets"
	"log"
	"net"
	"sync"
)

func ClienTest(){
	log.Println("clients  start")
	wg := &sync.WaitGroup{}
	for ii:=0;ii<10;ii++ {
		wg.Add(1)
		go func(wgg *sync.WaitGroup) {
			defer wgg.Done()
			conn, e := net.Dial("tcp4", conf.Conf.TCP.Bind)
			if e != nil{
				panic(e)
			}
			conn.Write(packets.Ping())
			// 登录
			conn.Write(packets.GetLoginData(conf.Conf.Auth.UserName,conf.Conf.Auth.PassWd))
			//声明一个管道用于接收解包的数据
			readerChannel := make(chan string, 1)
			go packets.ClientRetDataHandle(conn,readerChannel)
			packets.ClientReader(readerChannel)
			return
		}(wg)
	}
	ch := make(chan string,1);
	<-ch
}
