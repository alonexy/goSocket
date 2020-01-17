package conf

import (
	"flag"
	"github.com/BurntSushi/toml"
)

var (
	confPath string
	// Conf config
	Conf       = &Config{}
	ClientTest bool
)


type Config struct {
	Debug     bool
	Timeout   int
	Redis     *Redis
	WebSocket *WebSocket
	TCP       *TCP
	Auth      *Auth
}
// auth
type Auth struct {
	UserName string
	PassWd   string
}

// Redis Conf
type Redis struct {
	Address string
	Auth    string
	Type    string
}

// Websocket Conf
type WebSocket struct {
	Bind        string
	TLSOpen     bool
	TLSBind     string
	CertFile    string
	PrivateFile string
}

//TCP server Conf
type TCP struct {
	Bind string
	Timeout int
}

func init() {
	flag.StringVar(&confPath, "conf", "conf-example.toml", "conf file path .")
	flag.BoolVar(&ClientTest, "tests", false, "Open client test .")
}

// Init init conf.
func Init() error {
	return local()
}

func local() (err error) {
	if _, err = toml.DecodeFile(confPath, &Conf); err != nil {
		return
	}
	Conf.fix()
	return
}

func (c *Config) fix() {
	return
}
