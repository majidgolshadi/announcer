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

const ApiNotFoundMessage = "404 api not found"

// based on https://github.com/gin-gonic/gin/issues/205 issue we can't have something like /announcer/:announcer_name/send/
func RunHttpServer(port string, inputChannel chan<- *logic.ChannelAct, inputChat chan<- *output.Message) error {
	log.Info("rest api server listening on port ", port)

	return fasthttp.ListenAndServe(port, func(ctx *fasthttp.RequestCtx) {
		ctx.SetContentType("application/json")
		start := time.Now()

		switch string(ctx.Path()) {
		case "/v1/announce/channel":
			v1PostAnnounceChannelHandler(ctx, inputChannel)
		case "/v1/announce/users":
			v1PostAnnounceUsersHandler(ctx, inputChat)
		default:
			ctx.Error(ApiNotFoundMessage, fasthttp.StatusNotFound)
		}

		log.WithFields(log.Fields{
			"method":        string(ctx.Method()),
			"URI":           string(ctx.Path()),
			"status":        ctx.Response.Header.StatusCode(),
			"response_time": time.Now().Sub(start).Seconds(),
		}).Info("rest api")
	})
}

type announceChannelRequest struct {
	Message   string `json:"message"`
	ChannelId string `json:"channel_id"`
}

func v1PostAnnounceChannelHandler(ctx *fasthttp.RequestCtx, inputChannel chan<- *logic.ChannelAct) {
	if string(ctx.Method()) != "POST" {
		ctx.Error(ApiNotFoundMessage, fasthttp.StatusNotFound)
		return
	}

	var input announceChannelRequest
	if err := json.Unmarshal(ctx.Request.Body(), &input); err != nil {
		log.WithField("bad request", err.Error()).Error("channel rest api")
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	if input.ChannelId == "" {
		log.WithField("bad request", "channel id is empty").Error("channel rest api")
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

func v1PostAnnounceUsersHandler(ctx *fasthttp.RequestCtx, inputChat chan<- *output.Message) {
	if string(ctx.Method()) != "POST" {
		ctx.Error(ApiNotFoundMessage, fasthttp.StatusNotFound)
		return
	}

	var input announceUsersRequest
	if err := json.Unmarshal(ctx.Request.Body(), &input); err != nil {
		log.WithField("bad request", err.Error()).Error("users rest api")
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

	for _, username := range input.Usernames {
		inputChat <- &output.Message{
			Template: string(msgTmp),
			Username: username,
			Loggable: true,
		}
	}

	ctx.SetStatusCode(http.StatusOK)
}
