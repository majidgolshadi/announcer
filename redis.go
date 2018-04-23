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

	HashTable                 string
	CheckInterval             int
	connection                *redis.Client
	connectionStatus          bool
	keepConnectionAliveTicker *time.Ticker
}

func (r *Redis) keepConnectionAlive(duration time.Duration) {
	if r.keepConnectionAliveTicker != nil {
		r.keepConnectionAliveTicker.Stop()
	}

	r.keepConnectionAliveTicker = time.NewTicker(duration)
	go func() {
		for range r.keepConnectionAliveTicker.C {
			if r.connectionStatus == false {
				//log.Info("try to connect to redis...")
				r.connect()
			}

			if statusCmd := r.connection.Ping(); statusCmd.Err() != nil {
				//log.WithField("error", statusCmd.Err()).Warn("redis connection lost")
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

	r.keepConnectionAlive(time.Duration(r.CheckInterval) * time.Second)
	r.connectionStatus = true
}

func (r *Redis) usernameExists(username string) bool {
	if r.connectionStatus {
		result, _ := r.connection.HGet(r.HashTable, username).Result()
		//return result != ""
		println("redis result", result)
		return true
	}

	return true
}

func (r *Redis) getAllUsers() ([]string, error) {
	return r.connection.HKeys(r.HashTable).Result()
}

func (r *Redis) close() {
	log.Info("close redis connections to ", r.Address)
	r.connection.Close()
}
