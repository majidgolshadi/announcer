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
	MaxIdealConn  int
	MaxOpenConn   int
	CheckInterval time.Duration
}

const ChannelIDQuery = `SELECT channel_id FROM ws_channel_data WHERE channel_state="ACCEPTED" AND channel_channelid="%s"`
const UsersChannelUsernameQuery = `SELECT member_id,member_username FROM ws_channel_members WHERE member_channelid="%s" AND member_id > %d LIMIT %d`

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

	if opt.MaxOpenConn == 0 {
		opt.MaxOpenConn = 10
	}

	if opt.MaxIdealConn == 0 {
		opt.MaxIdealConn = 10
	}

	if opt.CheckInterval == 0 {
		opt.CheckInterval = 5
	}

	opt.CheckInterval = opt.CheckInterval * time.Second

	return nil
}

func NewMysqlChannelDataStore(opt *MysqlOpt) (*mysql, error) {
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

	ms.conn.SetMaxIdleConns(ms.opt.MaxIdealConn)
	ms.conn.SetMaxOpenConns(ms.opt.MaxOpenConn)

	log.Info("mysql connect established to ", ms.opt.Address)

	go ms.keepConnectionAlive()
	return ms.conn.Ping()
}

func (ms *mysql) keepConnectionAlive() {
	for range ms.checkConnTicker.C {
		if err := ms.conn.Ping(); err != nil {
			log.Warn("mysql connection lost ", err.Error())
			ms.conn.Close()
			ms.connect()
		}
	}
}

func (ms *mysql) GetChannelMembers(channelID string) (username <-chan string, err error) {
	// get channel id
	row := ms.conn.QueryRow(fmt.Sprintf(ChannelIDQuery, channelID))
	if row == nil {
		return nil, errors.New("mysql channel not found")
	}

	var id string
	if err := row.Scan(&id); err != nil {
		return nil, err
	}

	usernameChan := make(chan string)

	go func() {
		lastSeenMemberId := 0
		var monitMysqlTotalTime float64

		for rowCount := ms.opt.PageLength; rowCount >= ms.opt.PageLength; {
			start := time.Now()
			rows, err := ms.conn.Query(fmt.Sprintf(UsersChannelUsernameQuery, id, lastSeenMemberId, ms.opt.PageLength))
			monitMysqlTotalTime = monitMysqlTotalTime + time.Now().Sub(start).Seconds()

			if err != nil {
				log.Error("mysql query execution error: ", err.Error())
				break
			}

			lastSeenMemberId, rowCount = scanRows(rows, usernameChan)
			rows.Close()
		}

		close(usernameChan)
		log.WithFields(log.Fields{
			"channel_id":   channelID,
			"process_time": monitMysqlTotalTime,
		}).Info("mysql query")
	}()

	return usernameChan, nil
}

func scanRows(rows *sql.Rows, usernameChan chan string) (lastMemberId int, count int) {
	username := ""
	for rows.Next() {

		if err := rows.Scan(&lastMemberId, &username); err != nil {
			log.Error("mysql scan row error: ", err.Error())
		}

		count++
		usernameChan <- username
	}

	return
}

func (ms *mysql) Close() {
	log.Warn("mysql close connection to ", ms.opt.Address)
	ms.checkConnTicker.Stop()

	if ms.conn != nil {
		ms.conn.Close()
	}
}
