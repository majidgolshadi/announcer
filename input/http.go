package input

import (
	"encoding/base64"
	"encoding/json"
	"github.com/majidgolshadi/client-announcer/logic"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"net/http"
	"time"
)

const API_NOT_FOUND_MESSAGE = "404 api not found"

// based on https://github.com/gin-gonic/gin/issues/205 issue we can't have something like /announcer/:announcer_name/send/
func RunHttpServer(port string, inputChannel chan<- *logic.ChannelAct, inputUser chan<- *logic.UserAct) error {
	log.Info("rest api server listening on port ", port)

	return fasthttp.ListenAndServe(port, func(ctx *fasthttp.RequestCtx) {
		ctx.SetContentType("application/json")
		start := time.Now()

		switch string(ctx.Path()) {
		case "/v1/announce/channel":
			v1AnnounceChannelHandler(ctx, inputChannel)
		case "/v1/announce/user":
			v1AnnounceUserHandler(ctx, inputUser)
		default:
			ctx.Error(API_NOT_FOUND_MESSAGE, fasthttp.StatusNotFound)
		}

		log.Infof("%s | %s | %d | %ds",
			ctx.Method(), ctx.Path(), ctx.Response.Header.StatusCode(), time.Now().Sub(start))
	})
}

type announceChannelRequest struct {
	Message   string `json:"message" binding:"required"`
	ChannelId string `json:"channel_id" binding:"required"`
}

func v1AnnounceChannelHandler(ctx *fasthttp.RequestCtx, inputChannel chan<- *logic.ChannelAct) {
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

	msgTmp, err := base64.StdEncoding.DecodeString(input.Message)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte(err.Error()))
		return
	}

	inputChannel <- &logic.ChannelAct{
		MessageTemplate: string(msgTmp),
		ChannelID:       input.ChannelId,
	}

	ctx.SetStatusCode(http.StatusOK)
}

type announceUserRequest struct {
	Message  string `json:"message" binding:"required"`
	Username string `json:"username" binding:"required"`
}

func v1AnnounceUserHandler(ctx *fasthttp.RequestCtx, inputUser chan<- *logic.UserAct) {
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

	msgTmp, err := base64.StdEncoding.DecodeString(input.Message)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte(err.Error()))
		return
	}

	inputUser <- &logic.UserAct{
		MessageTemplate: string(msgTmp),
		Username:        input.Username,
	}

	ctx.SetStatusCode(http.StatusOK)
}
