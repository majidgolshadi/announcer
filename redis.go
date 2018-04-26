package client_announcer

import (
	"time"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

type redisDs struct {
	Address       string
	Password      string
	Database      int
	CheckInterval int

	connStatus      bool
	conn            *redis.Client
	checkConnTicker *time.Ticker
	logger          *log.Entry
}

func (r *redisDs) connectAlways() (err error) {
	if err = r.connect(); err != nil {
		return err
	}

	r.keepConnectionAlive(time.Duration(r.CheckInterval) * time.Second)
	return
}

func (r *redisDs) retryToConnect() {
	r.connect()
	r.keepConnectionAlive(time.Duration(r.CheckInterval) * time.Second)
}

func (r *redisDs) connect() (err error) {
	r.conn = redis.NewClient(&redis.Options{
		Addr:     r.Address,
		Password: r.Password,
		DB:       r.Database,
	})

	_, err = r.conn.Ping().Result()
	r.connStatus = bool(err == nil)
	return err
}

func (r *redisDs) keepConnectionAlive(duration time.Duration) {
	r.checkConnTicker = time.NewTicker(duration)
	go func() {
		for range r.checkConnTicker.C {
			if !r.connStatus {
				log.Info("try to connect to redis...")
				r.connect()
			}

			if statusCmd := r.conn.Ping(); statusCmd.Err() != nil {
				log.WithField("error", statusCmd.Err()).Warn("redis connection lost")
				r.close()
				r.connStatus = false
			}
		}
	}()
}

func (r *redisDs) connectionStatus() bool {
	return r.connStatus
}

func (r *redisDs) connection() *redis.Client {
	return r.conn
}

func (r *redisDs) close() {
	log.Info("close redis connections to ", r.Address)
	r.checkConnTicker.Stop()
	r.conn.Close()
}
