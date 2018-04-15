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
	Message string `json:"message" binding:"required"`
	ChannelId int `json:"channel_id" binding:"required"`
}

func AnnounceHandler(c *gin.Context) {
	var input announceRequest
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}


}
