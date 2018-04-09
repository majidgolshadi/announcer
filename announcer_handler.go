package client_announcer

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func RegisterAnnouncerHandler(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"status": "do idi"})
}

func DeregisterAnnouncerHandler(c *gin.Context) {

}

type announceRequest struct {
	Cluster string `json:"cluster"`
	From string `json:"from"`
	Message string `json:"message"`
	Group string `json:"group"`
	Channel string `json:"channel"`
}

func AnnounceHandler(c *gin.Context) {
	var input announceRequest
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}


}