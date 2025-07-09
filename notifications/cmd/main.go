package main

import (
	"log"
	"notifications/config"
	"notifications/grpc"
	"notifications/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()
	config.ConnectDatabase()

	go grpc.StartGRPCServer()
	r := gin.Default()

	// CORS libre con soporte para WebSockets
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Sec-WebSocket-Protocol, Sec-WebSocket-Key, Sec-WebSocket-Version, Upgrade, Connection")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Webhook Likes
	r.POST("/webhook/like", handlers.WebhookLike)

	// WEBSOCKET
	r.GET("/ws", handlers.WsHandler)

	// Endpoints para consultar notificaciones
	r.GET("/notifications/:userId", handlers.GetNotifications)
	r.PUT("/notifications/:notificationId/read", handlers.MarkNotificationAsRead)

	log.Println("HTTP server listening on :8001")
	r.Run(":8001")
}
