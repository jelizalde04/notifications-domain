package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Permitir todas las conexiones por ahora
		return true
	},
}

// Gin handler para manejar /ws
func HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	log.Println("WebSocket connection established")

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		log.Printf("Received message: %s\n", message)

		// Responder pong de prueba
		response := []byte("pong: " + string(message))
		err = conn.WriteMessage(messageType, response)
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}
