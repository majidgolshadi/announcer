package logic

import (
	"database/sql"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

type mysql struct {
	opt             *MysqlOpt
	checkConnTicker *time.Ticker
	conn            *sql.DB
	connStatus      bool
}

type MysqlOpt struct {
	Address       string
	Username      string
	Password      string
	Database      string
	CheckInterval time.Duration
}

func (opt *MysqlOpt) init() error {
	if opt.Database == "" {
		return errors.New("database name does not set")
	}

	if opt.Address == "" {
		opt.Address = "127.0.0.1:3306"
	}

	if opt.Username == "" {
		opt.Username = "root"
	}

	if opt.CheckInterval == 0 {
		opt.CheckInterval = 5 * time.Second
	}

	return nil
}

func NewMysql(opt *MysqlOpt) (*mysql, error) {
	if err := opt.init(); err != nil {
		return nil, err
	}

	return &mysql{
		opt:             opt,
		checkConnTicker: time.NewTicker(opt.CheckInterval),
	}, nil
}

func (ms *mysql) connect() (err error) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8",
		ms.opt.Username, ms.opt.Password, ms.opt.Address, ms.opt.Database)

	if ms.conn, err = sql.Open("mysql", dataSourceName); err != nil {
		return err
	}

	go ms.keepConnectionAlive()
	return ms.conn.Ping()
}

func (ms *mysql) keepConnectionAlive() {
	for range ms.checkConnTicker.C {
		if !ms.connStatus {
			log.Info("try to connect to redis...")
			ms.connect()
		}

		if err := ms.conn.Ping(); err != nil {
			log.WithField("error", err.Error()).Warn("Mysql connection lost")

			ms.conn.Close()
			ms.connStatus = false
		}
	}
}

func (ms *mysql) GetConnection() *sql.DB {
	return ms.conn
}

func (ms *mysql) close() {
	log.Info("close mysql connections to ", ms.opt.Address)
	ms.checkConnTicker.Stop()

	if ms.conn != nil {
		ms.conn.Close()
	}
}
