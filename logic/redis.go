package logic

import (
	"time"

	redisCli "github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

type redis struct {
	connStatus      bool
	conn            *redisCli.Client
	checkConnTicker *time.Ticker
	opt             *RedisOpt
}

type RedisOpt struct {
	Address       string
	Password      string
	Database      int
	CheckInterval time.Duration
}

func (opt *RedisOpt) init() {
	if opt.Address == "" {
		opt.Address = "127.0.0.1:6379"
	}

	if opt.CheckInterval == 0 {
		opt.CheckInterval = 5 * time.Second
	}
}

func NewRedis(opt *RedisOpt) *redis {
	opt.init()

	return &redis{
		opt: opt,
	}
}

func (r *redis) ConnectAndKeep() {
	r.connect()
	go r.keepConnectionAlive(r.opt.CheckInterval)
}

func (r *redis) connect() (err error) {
	r.conn = redisCli.NewClient(&redisCli.Options{
		Addr:     r.opt.Address,
		Password: r.opt.Password,
		DB:       r.opt.Database,
		OnConnect: func(conn *redisCli.Conn) error {
			log.Info("redis connection established")
			return nil
		},
	})

	_, err = r.conn.Ping().Result()
	r.connStatus = bool(err == nil)
	return err
}

func (r *redis) keepConnectionAlive(duration time.Duration) {
	r.checkConnTicker = time.NewTicker(duration)

	for range r.checkConnTicker.C {
		if !r.connStatus {
			log.Info("try to connect to redis...")
			r.connect()
		}

		if statusCmd := r.conn.Ping(); statusCmd.Err() != nil {
			log.WithField("error", statusCmd.Err()).Warn("redis connection lost")

			r.conn.Close()
			r.connStatus = false
		}
	}
}

func (r *redis) HKeys(HTable string) ([]string, error) {
	return r.conn.HKeys(HTable).Result()
}

func (r *redis) HGet(key string, filed string) (string, error) {
	return r.conn.HGet(key, filed).Result()
}

func (r *redis) close() {
	log.Info("close redis connections to ", r.opt.Address)
	r.checkConnTicker.Stop()
	r.conn.Close()
}
