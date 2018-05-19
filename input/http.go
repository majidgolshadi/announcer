package input

import (
	"encoding/base64"
	"encoding/json"
	"github.com/majidgolshadi/client-announcer/logic"
	"github.com/majidgolshadi/client-announcer/output"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"net/http"
	"time"
)

const API_NOT_FOUND_MESSAGE = "404 api not found"

// based on https://github.com/gin-gonic/gin/issues/205 issue we can't have something like /announcer/:announcer_name/send/
func RunHttpServer(port string, inputChannel chan<- *logic.ChannelAct, outputChannel chan<- *output.Msg) error {
	log.Info("rest api server listening on port ", port)

	return fasthttp.ListenAndServe(port, func(ctx *fasthttp.RequestCtx) {
		ctx.SetContentType("application/json")
		start := time.Now()

		switch string(ctx.Path()) {
		case "/v1/announce/channel":
			v1AnnounceChannelHandler(ctx, inputChannel)
		case "/v1/announce/users":
			v1AnnounceUsersHandler(ctx, outputChannel)
		default:
			ctx.Error(API_NOT_FOUND_MESSAGE, fasthttp.StatusNotFound)
		}

		log.Infof("%s | %s | %d | %fs",
			ctx.Method(), ctx.Path(), ctx.Response.Header.StatusCode(), time.Now().Sub(start).Seconds())
	})
}

type announceChannelRequest struct {
	Message   string `json:"message"`
	ChannelId string `json:"channel_id"`
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

type announceUsersRequest struct {
	Message   string   `json:"message"`
	Usernames []string `json:"usernames"`
}

func v1AnnounceUsersHandler(ctx *fasthttp.RequestCtx, outputChannel chan<- *output.Msg) {
	if string(ctx.Method()) != "POST" {
		ctx.Error(API_NOT_FOUND_MESSAGE, fasthttp.StatusNotFound)
		return
	}

	var input announceUsersRequest
	if err := json.Unmarshal(ctx.Request.Body(), &input); err != nil {
		log.WithField("bad request", ctx.Request.Body()).Error()
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	if len(input.Usernames) < 1 {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	msgTmp, err := base64.StdEncoding.DecodeString(input.Message)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte(err.Error()))
		return
	}

	go func() {
		messageTemplate := string(msgTmp)
		for _, username := range input.Usernames {
			outputChannel <- &output.Msg{
				Temp: messageTemplate,
				User: username,
			}
		}
	}()

	ctx.SetStatusCode(http.StatusOK)
}
