package client_announcer

import (
	"github.com/go-redis/redis"
	"time"
)

type Redis struct {
	Address  string
	Password string
	Database int

	connection                *redis.Client
	connectionStatus          bool
	keepConnectionAliveTicker *time.Ticker
}

func redisClientFactory(address string, password string, db int, checkInternal int) *Redis {
	r := &Redis{}
	r.connection = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})

	r.keepConnectionAlive(time.Duration(checkInternal) * time.Second)

	return r
}

func (r *Redis) keepConnectionAlive(duration time.Duration) {
	r.keepConnectionAliveTicker = time.NewTicker(duration)
	go func() {
		for range r.keepConnectionAliveTicker.C {
			if r.connectionStatus == false {
				r.connect()
			}

			if statusCmd := r.connection.Ping(); statusCmd.Err() != nil {
				r.connectionStatus = false
			}
		}
	}()
}

func (r *Redis) connect() {
	r.connection = redis.NewClient(&redis.Options{
		Addr:     r.Address,
		Password: r.Password,
		DB:       r.Database,
	})
}

func (r *Redis) usernameExists(username string) bool {
	if r.connectionStatus {
		result, _ := r.connection.HLen(username).Result()
		return result > 0
	}

	return true
}

func (r *Redis) getAllUsers() ([]string, error) {
	return r.connection.Keys("*").Result()
}

func (r *Redis) close() {
	r.connection.Close()
}
