package config

import (
	"os"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

var (
	AsynqClient *asynq.Client
	RedisClient *redis.Client
)

var (
	GoogleClientID     = os.Getenv("GOOGLE_CLIENT_ID")
	GoogleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	RedirectURI        = os.Getenv("GOOGLE_REDIRECT_URI")
)

func InitRedisAndAsynq() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	AsynqClient = asynq.NewClient(asynq.RedisClientOpt{
		Addr: os.Getenv("REDIS_ADDR"),
	})
}
