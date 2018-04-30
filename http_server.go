package client_announcer

import (
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"time"
)

var (
	onlineUserInq *onlineUserInquiry
	chatConnRepo  *ChatServerClusterRepository
)

const API_NOT_FOUND_MESSAGE = "404 api not found"

// based on https://github.com/gin-gonic/gin/issues/205 issue we can't have something like /announcer/:announcer_name/send/
func RunHttpServer(port string, inquiry *onlineUserInquiry, repository *ChatServerClusterRepository) error {
	onlineUserInq = inquiry
	chatConnRepo = repository

	log.Info("rest api server listening on port ", port)

	return fasthttp.ListenAndServe(port, func(ctx *fasthttp.RequestCtx) {
		ctx.SetContentType("application/json")
		start := time.Now().UnixNano()

		switch string(ctx.Path()) {
		case "/v1/announce/channel":
			V1AnnounceChannelHandler(ctx)
		case "/v1/announce/user":
			V1AnnounceUserHandler(ctx)
		default:
			ctx.Error(API_NOT_FOUND_MESSAGE, fasthttp.StatusNotFound)
		}

		log.Infof("%s | %s | %d | %dns",
			ctx.Method(), ctx.Path(), ctx.Response.Header.StatusCode(), time.Now().UnixNano()-start)
	})
}
