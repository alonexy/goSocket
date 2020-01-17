package pool

import (
	"github.com/alonexy/acps/conf"
	"github.com/garyburd/redigo/redis"
	"time"
)

var RedisClient *redis.Pool

func init() {
	// 建立连接池
	RedisClient = &redis.Pool{
		// 从配置文件获取maxidle以及maxactive，取不到则用后面的默认值
		MaxIdle: 50, //最大空闲连接数
		MaxActive:300,    //最大连接数量
		//MaxActive:   0,                 //连接池最大连接数量,不确定可以用0（0表示自动定义），按需分配
		//IdleTimeout: 300 * time.Second, //连接关闭时间 300秒 （300秒不使用自动关闭）
		Wait:        true, //超过 最大限制 等待
		IdleTimeout: 0, //一直有效
		Dial: func() (redis.Conn, error) { //要连接的redis数据库
			c, err := redis.Dial(conf.Conf.Redis.Type, conf.Conf.Redis.Address)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", conf.Conf.Redis.Auth); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}