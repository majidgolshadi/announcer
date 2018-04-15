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
	Cluster   string `json:"cluster"`
	Message   string `json:"message" binding:"required"`
	ChannelId int    `json:"channel_id" binding:"required"`
}

func AnnounceHandler(c *gin.Context) {
	var input announceRequest
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	users, err := onlineUserInq.GetOnlineUsers(input.ChannelId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
	}

	cluster, err := chatConnRepo.GetCluster(input.Cluster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
	}

	cluster.SendToUsers(input.Message, users)
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
