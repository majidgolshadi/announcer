package logic

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type userActivityRestApi struct {
	conn         *http.Client
	opt          *UserActivityRestApiOpt
	getOnlineUrl string
}

type UserActivityRestApiOpt struct {
	RestApiBaseUrl    string
	RequestTimeout    time.Duration
	IdleConnTimeout   time.Duration
	MaxIdleConnection int
	MaxRetries        int
}

const JsonContentType = "application/json"

func (opt *UserActivityRestApiOpt) init() error {
	if opt.RestApiBaseUrl == "" {
		return errors.New("user activity base url is empty")
	}

	if opt.MaxRetries < 0 {
		return errors.New("max retry must be at least 1")
	}

	if opt.MaxRetries == 0 {
		opt.MaxRetries = 3
	}

	if opt.RequestTimeout == 0 {
		opt.RequestTimeout = 500
	}

	if opt.MaxIdleConnection == 0 {
		opt.MaxIdleConnection = 100
	}

	if opt.IdleConnTimeout == 0 {
		opt.IdleConnTimeout = 90
	}

	opt.IdleConnTimeout = opt.IdleConnTimeout * time.Second
	opt.RequestTimeout = opt.RequestTimeout * time.Millisecond

	return nil
}

func NewUserActivityRestApi(opt *UserActivityRestApiOpt) (*userActivityRestApi, error) {
	if err := opt.init(); err != nil {
		return nil, err
	}

	r := &userActivityRestApi{
		conn: &http.Client{
			Timeout: opt.RequestTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        opt.MaxIdleConnection,
				MaxIdleConnsPerHost: opt.MaxIdleConnection,
				IdleConnTimeout:     opt.IdleConnTimeout,
			},
		},
		opt:          opt,
		getOnlineUrl: fmt.Sprintf("http://%s/v1/users/online", opt.RestApiBaseUrl),
	}

	return r, nil
}

func (r *userActivityRestApi) IsHeOnline(username string) bool {
	var result *http.Response
	err := errors.New("not call")

	for i := 0; i < r.opt.MaxRetries && err != nil; i++ {
		result, err = r.conn.Post(r.getOnlineUrl, JsonContentType,
			strings.NewReader(fmt.Sprintf("[\"%s\"]", username)))
	}

	defer result.Body.Close() // to make connection reusable

	res, _ := ioutil.ReadAll(result.Body)
	return len(res) > 2
}

func (r *userActivityRestApi) FilterOnlineUsers(usernames []string) []string {
	var result *http.Response
	err := errors.New("not call")

	for i := 0; i < r.opt.MaxRetries && err != nil; i++ {
		result, err = r.conn.Post(r.getOnlineUrl, JsonContentType,
			strings.NewReader(fmt.Sprintf("[\"%s\"]", strings.Join(usernames, "\",\""))))
	}

	defer result.Body.Close() // to make connection reusable

	body, _ := ioutil.ReadAll(result.Body)
	trimmed := strings.Trim(string(body), "]")
	trimmed = strings.Trim(trimmed, "[")

	return strings.Split(trimmed, ",")
}

func (r *userActivityRestApi) Close() {
	log.Warn("user activity rest api close connection to ", r.opt.RestApiBaseUrl)
}
