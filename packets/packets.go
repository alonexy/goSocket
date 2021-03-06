package packets

import (
	"bytes"
	"fmt"
	"github.com/alonexy/acps/conf"
	"github.com/google/uuid"
	"github.com/syyongx/php2go"
	"log"
	"net"
	"reflect"
	"strings"
	"time"
	"unsafe"
)

var (
	// client send heart hearder
	heart1 = []byte{2, 1, 3, 0, 0}
	// client ／ server  -crontab heart header
	heart2 = []byte{1, 1, 3, 0, 0}
	// login suc send header
	heart3 = []byte{2, 2, 3, 0, 66}

	LoginDatas = Login{}
)

type Login struct {
	DevId     string
	LoginTime string
	Token     string
	UserName  string
}

func ByteString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func StringByte(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}
// packet header handler
func HeadJoin(body []byte) []byte {
	var buffer bytes.Buffer
	defer func() {
		buffer.Reset()
	}()
	bodyExcept := len(body) / 256
	bodyFmod := len(body) % 256
	Headers := []byte{1, 2, 3, byte(bodyExcept), byte(bodyFmod)}
	buffer.Write(Headers)
	buffer.Write(body)
	return buffer.Bytes()
}

type ReaderMsgType struct {
	Tc *net.TCPConn
	Msg string
}


// data pack 
func PackData(data interface{}) ([]byte, error) {
	BytesData, err := php2go.JSONEncode(data)
	if (err != nil) {
		return []byte{}, err
	}
	return HeadJoin(BytesData), nil
}

// data unpack
func UnPackData(bytes []byte) (string, error) {

	return ByteString(bytes), nil
}

func GetServerHeaertData() []byte {
	return heart2;
}
func GetLoginSucData() []byte {
	return heart3;
}
func GetClientHeaertData() []byte {
	return heart2;
}
func GetClientPassiveHeaertData() []byte {
	return heart1;
}

// Is Server Heart
func IsHeartS(heart []byte) bool {
	heartRes1 := bytes.Compare(heart, heart1)
	heartRes2 := bytes.Compare(heart, heart2)
	if (heartRes1 == 0) {
		return true
	}
	if (heartRes2 == 0) {
		return true
	}
	return false
}
func IsHeartC(heart []byte) bool {
	heartRes1 := bytes.Compare(heart, heart2)
	if (heartRes1 == 0) {
		return true
	}
	return false
}

//check login  
func ClientCheckSuclogin(login []byte) bool {
	//2 2 3 0 66
	loginHeader := []byte{2, 2, 3, 0, 66}
	res := bytes.Compare(login, loginHeader)
	if (res == 0) {
		return true
	} else {
		return false
	}
}

// server retData Hander
func ServerRetDataHandle(conn *net.TCPConn, readerChannel chan<- ReaderMsgType) {
	buf := make([]byte, 1024)
	Buffers := bytes.NewBuffer([]byte{})
	defer func() {
		Buffers.Reset()
		conn.Close()
	}()
	for {
		if n, err := conn.Read(buf); err != nil {
			log.Printf("recvData Fail >>>>> ErrMsg :%v", err)
			return
		} else {
			//binary.Read(Buffers,binary.BigEndian,buf)
			if (len(buf) == 0) {
				continue
			}
			if (IsHeartS(buf[:5])) {
				conn.SetDeadline(time.Now().Add(time.Duration(conf.Conf.TCP.Timeout) * time.Second)) // set 5s timeout
			}
			//log.Println(buf[:n])
			Buffers.Write(buf[:n])
			tmpBytes := Buffers.Bytes()[:5] // 临时查看body长度 不进行读取
			// header：1-3-3-bodysize/256-bodySize%256 body
			// 被除数 = 除数 x 商 + 余数
			bodySize := (int(tmpBytes[3]))*int(256) + int(tmpBytes[4])
			//log.Println("bodySize==>", bodySize)
			//log.Println("Buffers==>", Buffers.Bytes())
			// 如果超过Buffer的长度
			if (bodySize > Buffers.Len()) {

			} else if (bodySize == 0) {
				//读取头信息
				DataHeaders := make([]byte, 5)
				Buffers.Read(DataHeaders)
				continue
			} else {
				//读出头信息
				DataHeaders := make([]byte, 5)
				Buffers.Read(DataHeaders)
				BodyDatas := make([]byte, bodySize)
				Buffers.Read(BodyDatas)
				if RetStr, err := UnPackData(BodyDatas); err != nil {
					log.Printf("UnPackData Err[%v]", err)
				} else {
					// TODO 简单写
					if (strings.Contains(strings.ToLower(RetStr), "devid")) {
						err := php2go.JSONDecode(StringByte(RetStr), &LoginDatas)
						if (err != nil) {
							log.Println("DevId:[%v]---->Auth Is Err [%v]", LoginDatas.DevId, err)
							return
						}
						if (LoginDatas.Token == Token(conf.Conf.Auth.PassWd, LoginDatas.LoginTime, conf.Conf.Auth.UserName, LoginDatas.DevId)) {
							conn.Write(GetLoginSucData())
							log.Printf("DevId:[%v]---->Auth SUC", LoginDatas.DevId)
						} else {
							log.Printf("DevId:[%v]---->Auth Fail", LoginDatas.DevId)
							return
						}
					}
					ReaderMsgData := ReaderMsgType{}
					ReaderMsgData.Tc = conn
					ReaderMsgData.Msg = RetStr

					readerChannel<- ReaderMsgData
				}
			}

		}

	}
}
// Client Test Func---
func ClientRetDataHandle(conn net.Conn, readerChannel chan string) {
	buf := make([]byte, 1024)
	Buffers := bytes.NewBuffer([]byte{})
	ch := ClientHeartTimer(conn)
	defer func() {
		Buffers.Reset()
		conn.Close()
		ch <- true
	}()
	for {
		if n, err := conn.Read(buf); err != nil {
			log.Printf("Client Recv Data >>>>> ErrMsg :%v", err)
			return
		} else {
			if (len(buf) == 0) {
				continue
			}
			// Client
			if (ClientCheckSuclogin(buf[:5])) {
				log.Println("Login Is Suc")
				//conn.Write(GetClientPassiveHeaertData())
				tticker := time.NewTicker(time.Duration(conf.Conf.Timeout) * time.Second)
				go func(tticker *time.Ticker) {
					defer tticker.Stop()
					for {
						select {
						case <-tticker.C:
							conn.Write(GetClientPassiveHeaertData())
						}
					}
				}(tticker)
				continue
			}
			if (IsHeartC(buf[:5])) {
				conn.Write(GetClientPassiveHeaertData())
				continue
			}
			Buffers.Write(buf[:n])
			tmpBytes := Buffers.Bytes()[:5] // 临时查看body长度 不进行读取
			// header：1-3-3-bodysize/256-bodySize%256 body
			// 被除数 = 除数 x 商 + 余数
			bodySize := (int(tmpBytes[3]))*int(256) + int(tmpBytes[4])
			//log.Println("bodySize==>", bodySize)
			//log.Println("Buffers==>", Buffers.Bytes())
			// 如果超过Buffer的长度
			if (bodySize > Buffers.Len()) {

			} else if (bodySize == 0) {
				//读取头信息
				DataHeaders := make([]byte, 5)
				Buffers.Read(DataHeaders)
				continue
			} else {
				//读出头信息
				DataHeaders := make([]byte, 5)
				Buffers.Read(DataHeaders)
				BodyDatas := make([]byte, bodySize)
				Buffers.Read(BodyDatas)
				if RetStr, err := UnPackData(BodyDatas); err != nil {
					log.Printf("UnPackData Err[%v]", err)
				} else {
					readerChannel <- RetStr
				}
			}

		}

	}
}
func ClientReader(readerChannel <-chan string) {
	for {
		select {
		case data := <-readerChannel:
			if (!php2go.Empty(data)) {
				log.Println(data)
			}
		}
	}
}
// Server整合数据读取
func ServerReader(readerChannel chan ReaderMsgType) {
	for {
		select {
		case data := <-readerChannel:
			if (!php2go.Empty(data)) {
				log.Println(data.Msg)

			}
		}
	}
}
func GetUid() string {
	UUid, _ := uuid.NewRandom()
	return fmt.Sprintf("%s", UUid)
}
func Token(passwd string, loginTime string, uname string, DevId string) string {
	return php2go.Md5(passwd + loginTime + uname + DevId);
}

// get login data
func GetLoginData(uname string, passwd string) []byte {
	loginTime := fmt.Sprintf("%v", time.Now().Format("2006-01-02 15:04:05"))
	DevId := GetUid()
	data := make(map[string]string)
	data["UserName"] = uname
	data["LoginTime"] = loginTime
	data["DevId"] = DevId
	data["Token"] = Token(passwd, loginTime, uname, DevId)
	data["Action"] = "login"

	BytesData, err := PackData(data)
	if (err != nil) {
		log.Println("GetLoginData", err)
		return []byte{}
	}
	return BytesData
}

//心跳定时器
func ServerHeartTimer(conn *net.TCPConn) chan bool {
	ticker := time.NewTicker(time.Duration(conf.Conf.Timeout) * time.Second)
	stopChan := make(chan bool)
	go func(ticker *time.Ticker) {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				conn.Write(GetServerHeaertData())
				log.Printf("Server Timer Send heart at %v\r\n", time.Now().Format("2006-01-02 15:04:05"))
			case stop := <-stopChan:
				if stop {
					log.Println("Server Timer Stop")
					return
				}
			}
		}
	}(ticker)
	return stopChan
}
func ClientHeartTimer(conn net.Conn) chan bool {
	ticker := time.NewTicker(time.Duration(conf.Conf.Timeout) * time.Second)
	stopChan := make(chan bool)
	go func(ticker *time.Ticker) {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				conn.Write(GetClientHeaertData())
				log.Printf("Client Timer Send heart at %v\r\n", time.Now().Format("2006-01-02 15:04:05"))
			case stop := <-stopChan:
				if stop {
					log.Println("Client Timer Stop")
					return
				}
			}
		}
	}(ticker)
	return stopChan
}

func Ping() []byte {
	str := "PING"
	BytesData, _ := PackData(str)
	return BytesData
}

func Pong() []byte {
	str := "PONG"
	BytesData, _ := PackData(str)
	return BytesData
}
func T() []byte {
	str := "T"
	BytesData, _ := PackData(str)
	return BytesData
}
