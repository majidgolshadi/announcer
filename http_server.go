package client_announcer

import "github.com/gin-gonic/gin"

var (
	onlineUserInq *onlineUserInquiry
	chatConnRepo  *chatServerClusterRepository
)

func RunHttpServer(port string, inquiry *onlineUserInquiry, repository *chatServerClusterRepository) error {
	onlineUserInq = inquiry
	chatConnRepo = repository

	router := gin.Default()

	v1 := router.Group("/v1")
	{
		v1.POST("/register/announcer/", RegisterAnnouncerHandler)
		v1.DELETE("/register/announcer/", DeregisterAnnouncerHandler)

		// based on https://github.com/gin-gonic/gin/issues/205 issue we can't have something like /announcer/:announcer_name/send/
		v1.POST("/announce", AnnounceHandler)
	}

	return router.Run(port)
}
