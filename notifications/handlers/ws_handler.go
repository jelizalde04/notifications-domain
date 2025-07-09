package handlers

import (
	"fmt"
	"log"
	"net/http"
	"notifications/config"
	"notifications/models"
	"notifications/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// En desarrollo, permite cualquier origen
		// En producción, deberías verificar el origen específico
		origin := r.Header.Get("Origin")
		log.Println("WebSocket connection attempt from origin:", origin)
		return true // Cambia esto en producción para verificar orígenes específicos
	},
	// Buffer sizes opcionales para mejorar rendimiento
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var Connections = make(map[string]*websocket.Conn)

func WsHandler(c *gin.Context) {
	// Validación de autenticación ANTES del upgrade
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		log.Println("Missing Authorization header")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		log.Println("Invalid Authorization header format:", authHeader)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
		return
	}

	tokenStr := tokenParts[1]
	log.Println("Token received for WebSocket connection")

	claims, err := utils.ParseJWT(tokenStr)
	if err != nil {
		log.Println("Token parse error:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	userId, ok := claims["userId"].(string)
	if !ok {
		log.Println("Invalid token payload: missing or wrong userId type")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token payload"})
		return
	}

	log.Println("Token valid for user:", userId, "- Proceeding with WebSocket upgrade")

	// Solo hacer upgrade después de validar autenticación
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	// Cerrar conexión existente si el usuario ya está conectado
	if existingConn, exists := Connections[userId]; exists {
		log.Println("Closing existing connection for user:", userId)
		existingConn.Close()
	}

	Connections[userId] = conn
	log.Printf("New WebSocket connection established for user: %s", userId)
	log.Printf("Total active connections: %d", len(Connections))
	log.Printf("Connected users: %v", getConnectedUsersList())

	defer func() {
		conn.Close()
		delete(Connections, userId)
		log.Println("Connection closed for user:", userId)
	}()

	// Enviar mensaje de confirmación
	welcomeMsg := fmt.Sprintf(`{"type":"welcome","message":"Connected successfully","userId":"%s"}`, userId)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(welcomeMsg)); err != nil {
		log.Println("Failed to send welcome message:", err)
		return
	}

	// Enviar notificaciones pendientes
	go sendPendingNotifications(userId, conn)

	// Loop de lectura de mensajes
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("WebSocket unexpected close error:", err)
			} else {
				log.Println("WebSocket connection closed normally for user:", userId)
			}
			break
		}
		log.Println("Message received from user", userId, ":", string(msg))

		// Echo del mensaje como confirmación
		response := fmt.Sprintf(`{"type":"echo","message":"Message received","original":"%s"}`, string(msg))
		if err := conn.WriteMessage(websocket.TextMessage, []byte(response)); err != nil {
			log.Println("Failed to send echo response:", err)
			break
		}
	}
}

// sendPendingNotifications envía las notificaciones no leídas al usuario cuando se conecta
func sendPendingNotifications(userId string, conn *websocket.Conn) {
	userUUID, err := uuid.Parse(userId)
	if err != nil {
		log.Printf("Invalid userId format for pending notifications: %s", userId)
		return
	}

	// Obtener notificaciones no leídas de los últimos 30 días
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var notifications []models.Notification

	if err := config.DB.Where(`"responsibleId" = ? AND read = ? AND timestamp >= ?`,
		userUUID, false, thirtyDaysAgo).
		Order("timestamp DESC").
		Limit(50). // Limitar a 50 notificaciones para evitar sobrecarga
		Find(&notifications).Error; err != nil {
		log.Printf("Error fetching pending notifications for user %s: %v", userId, err)
		return
	}

	if len(notifications) == 0 {
		log.Printf("No pending notifications for user %s", userId)
		return
	}

	log.Printf("Sending %d pending notifications to user %s", len(notifications), userId)

	// Enviar cada notificación
	for _, notification := range notifications {
		notificationMessage := fmt.Sprintf(`{
			"type": "notification",
			"id": "%s",
			"actorId": "%s",
			"recipientId": "%s",
			"responsibleId": "%s",
			"notificationType": "%s",
			"content": "%s",
			"timestamp": "%s",
			"read": false,
			"pending": true
		}`, notification.ID, notification.ActorID, notification.RecipientID, notification.ResponsibleID,
			notification.Type, notification.Content, notification.Timestamp.Format(time.RFC3339))

		if err := conn.WriteMessage(websocket.TextMessage, []byte(notificationMessage)); err != nil {
			log.Printf("Failed to send pending notification %s to user %s: %v",
				notification.ID, userId, err)
			// Si falla el envío de una notificación, paramos para no saturar el log
			break
		}
	}

	log.Printf("Finished sending pending notifications to user %s", userId)
}

func SendNotification(userId string, message string) error {
	log.Printf("SendNotification called for userId: %s", userId)
	log.Printf("Current active connections: %v", getConnectedUsersList())

	conn, ok := Connections[userId]
	if !ok {
		log.Printf("User %s is not connected. Notification will be stored for later delivery.", userId)
		return fmt.Errorf("user not connected")
	}

	err := conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Printf("Failed to send notification to user %s: %v", userId, err)
		return err
	}

	log.Printf("Notification sent successfully to user %s", userId)
	return nil
}

// Función auxiliar para debug - obtener lista de usuarios conectados
func getConnectedUsersList() []string {
	var users []string
	for userId := range Connections {
		users = append(users, userId)
	}
	return users
}
