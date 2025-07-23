package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func handleWebSocket(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("WebSocket upgrade error: %v", err)
        return
    }
    defer conn.Close()

    wsConnections[conn] = true
    defer delete(wsConnections, conn)

    log.Println("WebSocket client connected")

    for {
        messageType, message, err := conn.ReadMessage()
        if err != nil {
            log.Printf("WebSocket read error: %v", err)
            break
        }

        log.Printf("Received WebSocket message: %s", message)
        
        // Echo message back (in a real app, you'd process it)
        if err := conn.WriteMessage(messageType, message); err != nil {
            log.Printf("WebSocket write error: %v", err)
            break
        }
    }

    log.Println("WebSocket client disconnected")
}

func broadcastToWebSockets(message map[string]interface{}) {
    for conn := range wsConnections {
        if err := conn.WriteJSON(message); err != nil {
            log.Printf("WebSocket broadcast error: %v", err)
            conn.Close()
            delete(wsConnections, conn)
        }
    }
}
