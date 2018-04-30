package client_announcer

import (
	"encoding/base64"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"net/http"
)

type announceChannelRequest struct {
	Cluster   string `json:"cluster"`
	Message   string `json:"message" binding:"required"`
	ChannelId string `json:"channel_id" binding:"required"`
}

func V1AnnounceChannelHandler(ctx *fasthttp.RequestCtx) {
	if string(ctx.Method()) != "POST" {
		ctx.Error(API_NOT_FOUND_MESSAGE, fasthttp.StatusNotFound)
		return
	}

	var input announceChannelRequest
	if err := json.Unmarshal(ctx.Request.Body(), &input); err != nil {
		log.WithField("bad request", ctx.Request.Body()).Error()
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	msg, err := base64.StdEncoding.DecodeString(input.Message)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte(err.Error()))
		return
	}

	users, err := onlineUserInq.GetOnlineUsers(input.ChannelId)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBody([]byte(err.Error()))
		return
	}

	cluster, err := chatConnRepo.Get(input.Cluster)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBody([]byte(err.Error()))
		return
	}

	go cluster.SendToUsers(string(msg), users)
	ctx.SetStatusCode(http.StatusOK)
}

type announceUserRequest struct {
	Cluster  string `json:"cluster"`
	Message  string `json:"message" binding:"required"`
	Username string `json:"username" binding:"required"`
}

func V1AnnounceUserHandler(ctx *fasthttp.RequestCtx) {
	if string(ctx.Method()) != "POST" {
		ctx.Error(API_NOT_FOUND_MESSAGE, fasthttp.StatusNotFound)
		return
	}

	var input announceUserRequest
	if err := json.Unmarshal(ctx.Request.Body(), &input); err != nil {
		log.WithField("bad request", ctx.Request.Body()).Error()
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	msg, err := base64.StdEncoding.DecodeString(input.Message)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte(err.Error()))
		return
	}

	cluster, err := chatConnRepo.Get(input.Cluster)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBody([]byte(err.Error()))
		return
	}

	go cluster.SendToUsers(string(msg), []string{input.Username})
	ctx.SetStatusCode(http.StatusOK)
}
