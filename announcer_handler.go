package client_announcer

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func RegisterAnnouncerHandler(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"status": "doidi"})
}

func DeregisterAnnouncerHandler(c *gin.Context) {

}

func AnnounceHandler(c *gin.Context) {

}