package logic

import (
	"errors"
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
	Address     string
	Password    string
	Database    int
	SetPrefix   string
	ReadTimeout time.Duration
	MaxRetries  int
}

func (opt *RedisOpt) init() error {
	if opt.SetPrefix == "" {
		return errors.New("set prefix is empty")
	}

	if opt.Address == "" {
		opt.Address = "127.0.0.1:6379"
	}

	if opt.ReadTimeout == 0 {
		opt.ReadTimeout = 1
	}

	opt.ReadTimeout = opt.ReadTimeout * time.Millisecond

	return nil
}

func NewRedisUserDataStore(opt *RedisOpt) (*redis, error) {
	if err := opt.init(); err != nil {
		return nil, err
	}

	r := &redis{
		opt: opt,
	}

	r.connect()
	return r, nil
}

func (r *redis) connect() (err error) {
	r.conn = redisCli.NewClient(&redisCli.Options{
		Addr:        r.opt.Address,
		Password:    r.opt.Password,
		DB:          r.opt.Database,
		MaxRetries:  r.opt.MaxRetries,
		ReadTimeout: r.opt.ReadTimeout,
		OnConnect: func(conn *redisCli.Conn) error {
			log.Info("redis connection established to ", r.opt.Address)
			return nil
		},
	})

	_, err = r.conn.Ping().Result()
	return err
}

func (r *redis) GetAllOnlineUsers() (<-chan string, error) {
	usersChan := make(chan string)
	ips, err := r.conn.Keys(r.opt.SetPrefix).Result()
	if err != nil {
		return nil, err
	}

	go func() {
		for _, ip := range ips {
			members, _ := r.conn.SMembers(ip).Result()
			for _, member := range members {
				usersChan <- member
			}
		}

		close(usersChan)
	}()

	return usersChan, nil
}

func (r *redis) IsHeOnline(username string) bool {
	result, _ := r.conn.Get(username).Result()
	return result != ""
}

func (r *redis) WhichOneIsOnline(usernames []string) []interface{} {
	result, _ := r.conn.MGet(usernames...).Result()
	return result
}

func (r *redis) Close() {
	log.Warn("redis close connection to ", r.opt.Address)
	if r.conn != nil {
		r.conn.Close()
	}
}
