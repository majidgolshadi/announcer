package client_announcer

import (
	"time"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

type Redis struct {
	Address  string
	Password string
	Database int

	hashTable                 string
	connection                *redis.Client
	connectionStatus          bool
	keepConnectionAliveTicker *time.Ticker
}

func redisClientFactory(address string, password string, db int, hashTable string, checkInternal int) *Redis {
	r := &Redis{}
	r.connection = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})

	r.hashTable = hashTable
	r.keepConnectionAlive(time.Duration(checkInternal) * time.Second)

	return r
}

func (r *Redis) keepConnectionAlive(duration time.Duration) {
	r.keepConnectionAliveTicker = time.NewTicker(duration)
	go func() {
		for range r.keepConnectionAliveTicker.C {
			if r.connectionStatus == false {
				log.Info("try to connect to redis...")
				r.connect()
			}

			if statusCmd := r.connection.Ping(); statusCmd.Err() != nil {
				log.WithField("error", statusCmd.Err()).Warn("redis connection lost")
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
	return r.connection.HKeys(r.hashTable).Result()
}

func (r *Redis) close() {
	r.connection.Close()
}
