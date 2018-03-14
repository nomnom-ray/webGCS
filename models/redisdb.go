package models

import (
	"github.com/go-redis/redis"
)

//Client is globally accessable; used in lotparking; would like to privatize sometimes
var client *redis.Client
var tileClient *redis.Client

//InitRedis serves clients from redis ??? not sure advantage over direct
func InitRedis() {
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", //default port of redis-server; lo-host when same machine
	})

	tileClient = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:9851",
	})

}
