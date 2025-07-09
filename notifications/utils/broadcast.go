package utils

import (
	"encoding/json"
	"log"
	"notifications/models"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Canal para pasar notificaciones desde handlers
var NotificationChan = make(chan models.Notification, 100)

// Mapa y mutex para conexiones WebSocket
var Connections = make(map[string]*websocket.Conn)
var ConnectionsMu = &sync.Mutex{}

// Envía la notificación solo al destinatario conectado
func BroadcastNotification(noti models.Notification) {
	payload := map[string]interface{}{
		"actorId":     noti.ActorID.String(),
		"recipientId": noti.RecipientID.String(),
		"type":        noti.Type,
		"content":     noti.Content,
		"timestamp":   noti.Timestamp.Format(time.RFC3339),
	}
	data, _ := json.Marshal(payload)

	ConnectionsMu.Lock()
	defer ConnectionsMu.Unlock()

	userID := noti.RecipientID.String()
	if conn, ok := Connections[userID]; ok {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Println("WS write error:", err)
			conn.Close()
			delete(Connections, userID)
		}
	}
}
