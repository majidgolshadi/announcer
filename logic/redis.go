package logic

import (
	redisCli "github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"time"
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
		opt.CheckInterval = 5
	}

	opt.CheckInterval = opt.CheckInterval * time.Second
}

func NewRedis(opt *RedisOpt) *redis {
	opt.init()

	return &redis{
		opt:             opt,
		checkConnTicker: time.NewTicker(opt.CheckInterval),
	}
}

func (r *redis) connectAndKeep() {
	r.connect()
	go r.keepConnectionAlive()
}

func (r *redis) connect() (err error) {
	r.conn = redisCli.NewClient(&redisCli.Options{
		Addr:       r.opt.Address,
		Password:   r.opt.Password,
		DB:         r.opt.Database,
		MaxRetries: 10,
		OnConnect: func(conn *redisCli.Conn) error {
			log.Info("redis connection established to ", r.opt.Address)
			return nil
		},
	})

	_, err = r.conn.Ping().Result()
	return err
}

func (r *redis) keepConnectionAlive() {
	for range r.checkConnTicker.C {
		if statusCmd := r.conn.Ping(); statusCmd.Err() != nil {
			log.WithField("error", statusCmd.Err()).Warn("redis connection lost")
			r.conn.Close()
			r.connect()
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
	log.Warn("close redis connections to ", r.opt.Address)
	r.checkConnTicker.Stop()

	if r.conn != nil {
		r.conn.Close()
	}
}
