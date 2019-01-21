package conf

import (
	"github.com/alonexy/acps/logger"
	"github.com/garyburd/redigo/redis"
)

// 获取redis 连接
func GetRedisCon() (redis.Conn, error) {
	c, err := redis.Dial("tcp", Conf.Redis.Host + Conf.Redis.Port, redis.DialPassword(Conf.Redis.Passwd), redis.DialDatabase(Conf.Redis.Db))
	if err != nil {
		logger.Errorln("GetRedisCon %v",err)
		return nil, err
	}
	return c, nil
}
