package handlers

import (
	"fmt"
	"log"
	"net/http"
	"notifications/config"
	"notifications/dto"
	"notifications/models"
	"notifications/utils"
	"strconv"
	"strings"
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

	// Debug: Log de los datos recibidos
	log.Printf("=== WEBHOOK RECEIVED ===")
	log.Printf("Event: %s", req.Event)
	log.Printf("ActorId: %s", req.Data.ActorId)
	log.Printf("RecipientId: %s", req.Data.RecipientId)
	log.Printf("ResponsibleId: %s", req.Data.ResponsibleId)
	log.Printf("Type: %s", req.Data.Type)
	log.Printf("Content: %s", req.Data.Content)
	log.Printf("========================")

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

	responsibleUUID, err := uuid.Parse(req.Data.ResponsibleId)
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
		ActorID:       actorUUID,
		RecipientID:   recipientUUID,
		ResponsibleID: responsibleUUID,
		Type:          req.Data.Type,
		Content:       req.Data.Content,
		Read:          false,
		Timestamp:     parsedTime,
	}

	if err := config.DB.Create(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving notification"})
		return
	}

	// Enviar notificación por WebSocket si el usuario está conectado
	notificationMessage := fmt.Sprintf(`{
		"type": "notification",
		"id": "%s",
		"actorId": "%s",
		"recipientId": "%s",
		"notificationType": "%s",
		"content": "%s",
		"timestamp": "%s",
		"read": false
	}`, notification.ID, notification.ActorID, notification.RecipientID,
		notification.Type, notification.Content, notification.Timestamp.Format(time.RFC3339))

	// Intentar enviar por WebSocket (no falla si el usuario no está conectado)
	log.Printf("Attempting to send WebSocket notification to responsibleId: %s", req.Data.ResponsibleId)
	if err := SendNotification(req.Data.ResponsibleId, notificationMessage); err != nil {
		// Log pero no falla la operación
		log.Printf("Could not send WebSocket notification to user %s: %v", req.Data.ResponsibleId, err)
	} else {
		log.Printf("WebSocket notification sent successfully to user %s", req.Data.ResponsibleId)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification created successfully!",
		"id":      notification.ID,
	})
}

// GetNotifications obtiene las notificaciones de un usuario desde la base de datos
func GetNotifications(c *gin.Context) {
	userIdParam := c.Param("userId")

	// Validar que el userId sea un UUID válido
	userUUID, err := uuid.Parse(userIdParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userId format"})
		return
	}

	// Validar token de autorización
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
		return
	}

	tokenStr := tokenParts[1]
	claims, err := utils.ParseJWT(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	tokenUserId, ok := claims["userId"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token payload"})
		return
	}

	// Verificar que el usuario del token coincida con el solicitado
	if tokenUserId != userIdParam {
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized to access these notifications"})
		return
	}

	// Obtener parámetros de consulta opcionales
	limit := 50 // Límite por defecto
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	onlyUnread := c.Query("unread") == "true"

	// Construir la consulta usando ResponsibleID
	query := config.DB.Where(`"responsibleId" = ?`, userUUID)
	if onlyUnread {
		query = query.Where("read = ?", false)
	}

	var notifications []models.Notification
	if err := query.Order("timestamp DESC").Limit(limit).Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving notifications"})
		return
	}

	// Formatear las notificaciones para la respuesta
	var response []gin.H
	for _, notification := range notifications {
		response = append(response, gin.H{
			"id":            notification.ID,
			"actorId":       notification.ActorID,
			"recipientId":   notification.RecipientID,
			"responsibleId": notification.ResponsibleID,
			"type":          notification.Type,
			"content":       notification.Content,
			"read":          notification.Read,
			"timestamp":     notification.Timestamp.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": response,
		"count":         len(response),
	})
}

// MarkNotificationAsRead marca una notificación como leída
func MarkNotificationAsRead(c *gin.Context) {
	notificationIdParam := c.Param("notificationId")

	// Validar que el notificationId sea un UUID válido
	notificationUUID, err := uuid.Parse(notificationIdParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notificationId format"})
		return
	}

	// Validar token de autorización
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
		return
	}

	tokenStr := tokenParts[1]
	claims, err := utils.ParseJWT(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	tokenUserId, ok := claims["userId"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token payload"})
		return
	}

	// Buscar la notificación y verificar que pertenece al usuario
	var notification models.Notification
	if err := config.DB.Where(`id = ? AND "responsibleId" = ?`, notificationUUID, tokenUserId).First(&notification).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found or unauthorized"})
		return
	}

	// Marcar como leída
	if err := config.DB.Model(&notification).Update("read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read",
		"id":      notification.ID,
	})
}
