package client_announcer

import "github.com/gin-gonic/gin"

func RunHttpServer(port string) error {
	router := gin.Default()

	v1 := router.Group("/v1")
	{
		v1.POST("/register/announcer/", RegisterAnnouncerHandler)
		v1.DELETE("/register/announcer/", DeregisterAnnouncerHandler)

		// These two URI must be something like /announcer/:announcer_name/send/group but
		// based on https://github.com/gin-gonic/gin/issues/205 issue we can't have it
		v1.POST("/announce", AnnounceHandler)
	}

	return router.Run(port)
}
