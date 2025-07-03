package handlers

import (
	"net/http"
	"notifications/config"
	"notifications/dto"
	"notifications/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func WebhookLike(c *gin.Context) {
	var req dto.LikeWebhookRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorUUID, err := uuid.Parse(req.Data.ActorId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid actorId UUID"})
		return
	}

	recipientUUID, err := uuid.Parse(req.Data.RecipientId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipientId UUID"})
		return
	}

	_, err = uuid.Parse(req.Data.ResponsibleId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid responsibleId UUID"})
		return
	}

	// Parse timestamp compatible with ISO format from Python
	parsedTime, err := time.Parse("2006-01-02T15:04:05.999999", req.Data.Timestamp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timestamp format"})
		return
	}

	notification := models.Notification{
		ActorID:     actorUUID,
		RecipientID: recipientUUID,
		Type:        req.Data.Type,
		Content:     req.Data.Content,
		Read:        false,
		Timestamp:   parsedTime,
	}

	if err := config.DB.Create(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification created successfully!",
	})
}
