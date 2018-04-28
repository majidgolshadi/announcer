package client_announcer

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type announceChannelRequest struct {
	Cluster   string `json:"cluster"`
	Message   string `json:"message" binding:"required"`
	ChannelId string `json:"channel_id" binding:"required"`
}

func AnnounceChannelHandler(c *gin.Context) {
	var input announceChannelRequest
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	users, err := onlineUserInq.GetOnlineUsers(input.ChannelId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	cluster, err := chatConnRepo.Get(input.Cluster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	go cluster.SendToUsers(input.Message, users)
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

type announceUserRequest struct {
	Cluster  string `json:"cluster"`
	Message  string `json:"message" binding:"required"`
	Username string `json:"username" binding:"required"`
}

func AnnounceUserHandler(c *gin.Context) {
	var input announceUserRequest
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	cluster, err := chatConnRepo.Get(input.Cluster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	go cluster.SendToUsers(input.Message, []string{input.Username})
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
