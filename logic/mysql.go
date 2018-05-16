package logic

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
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
	PageLength    int
	CheckInterval time.Duration
}

const ChannelIDQuery = `select channel_id from ws_channel_data where channel_channelid="%s"`
const UsersChannelUsernameQuery = `select member_username from ws_channel_members where member_channelid="%s" limit %d,%d`

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
		opt.CheckInterval = 5
	}

	opt.CheckInterval = opt.CheckInterval * time.Second

	return nil
}

func NewMysqlChannelDataStore(opt *MysqlOpt) (ChannelDataStore, error) {
	if err := opt.init(); err != nil {
		return nil, err
	}

	ms := &mysql{
		opt:             opt,
		checkConnTicker: time.NewTicker(opt.CheckInterval),
	}

	if err := ms.connect(); err != nil {
		return nil, err
	}

	return ms, nil
}

func (ms *mysql) connect() (err error) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8",
		ms.opt.Username, ms.opt.Password, ms.opt.Address, ms.opt.Database)

	if ms.conn, err = sql.Open("mysql", dataSourceName); err != nil {
		return err
	}

	log.Info("connect to mysql ", ms.opt.Address)

	go ms.keepConnectionAlive()
	return ms.conn.Ping()
}

func (ms *mysql) keepConnectionAlive() {
	for range ms.checkConnTicker.C {
		if err := ms.conn.Ping(); err != nil {
			log.WithField("error", err.Error()).Warn("Mysql connection lost")
			ms.conn.Close()
			ms.connect()
		}
	}
}

func (ms *mysql) GetChannelMembers(channelID string) (username <-chan string, err error) {
	usernameChan := make(chan string)

	row := ms.conn.QueryRow(fmt.Sprintf(ChannelIDQuery, channelID))
	if row == nil {
		return nil, errors.New("channel not found")
	}

	var id string
	if err := row.Scan(&id); err != nil {
		return nil, err
	}

	go func() error {
		var username string
		var emptyFlag bool
		for offset := 0; ; offset = offset + ms.opt.PageLength {
			emptyFlag = true
			rows, err := ms.conn.Query(fmt.Sprintf(UsersChannelUsernameQuery, id, offset, ms.opt.PageLength))
			if err != nil {
				log.Error("query execution error: ", err.Error())
				break
			}

			for rows.Next() {
				emptyFlag = false
				if err := rows.Scan(&username); err != nil {
					log.Error("scan mysql row error: ", err.Error())
					continue
				}

				usernameChan <- username
			}

			rows.Close()
			if emptyFlag {
				break
			}
		}

		close(usernameChan)
		return nil
	}()

	return usernameChan, nil

}

func (ms *mysql) Close() {
	log.Warn("close mysql connections to ", ms.opt.Address)
	ms.checkConnTicker.Stop()

	if ms.conn != nil {
		ms.conn.Close()
	}
}
