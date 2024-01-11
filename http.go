package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	connectedUsers   = make(map[*websocket.Conn]string)
	connectedUsersMu sync.Mutex
)

// Credentials struct to represent the JSON message containing username and password
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func handleWebSocket(c *gin.Context) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	// Read the initial message containing username and password
	var credentials Credentials
	if err := conn.ReadJSON(&credentials); err != nil {
		fmt.Println("Error reading credentials:", err)
		return
	}

	// Print username and password (for debugging purposes)
	fmt.Printf("User '%s' connected with password '%s'\n", credentials.Username, credentials.Password)

	// Authenticate user (replace with your own authentication logic)
	if authenticate(credentials.Username, credentials.Password) {
		// Lock the mutex before updating connected users map
		connectedUsersMu.Lock()
		connectedUsers[conn] = credentials.Username
		numUsers := len(connectedUsers)
		connectedUsersMu.Unlock()

		fmt.Printf("User '%s' authenticated. Total users: %d\n", credentials.Username, numUsers)

		// Send a welcome message to the user
		conn.WriteJSON(map[string]string{"message": "Welcome to the chat!"})

		// Start a Goroutine to handle incoming messages from this user
		go handleUserMessages(conn, credentials.Username)
	} else {
		// Authentication failed, close the connection
		fmt.Printf("User '%s' failed to authenticate.\n", credentials.Username)
		conn.WriteJSON(map[string]string{"error": "Authentication failed"})
		return
	}

	// Keep the connection open until the user disconnects
	for {
		// Read any incoming messages (implement based on your requirements)
		var message map[string]interface{}
		if err := conn.ReadJSON(&message); err != nil {
			fmt.Printf("User '%s' disconnected.\n", credentials.Username)
			// Lock the mutex before deleting the user from the connected users map
			connectedUsersMu.Lock()
			delete(connectedUsers, conn)
			connectedUsersMu.Unlock()
			break
		}
	}
}

func handleUserMessages(conn *websocket.Conn, username string) {
	for {
		// Read and process incoming messages from the user (implement based on your requirements)
		var message map[string]interface{}
		if err := conn.ReadJSON(&message); err != nil {
			fmt.Printf("Error reading message from user '%s': %v\n", username, err)
			break
		}

		// Handle the message (implement based on your requirements)
		fmt.Printf("User '%s' sent a message: %v\n", username, message)
	}
}

func authenticate(username, password string) bool {
	// Implement your own authentication logic here
	// For simplicity, this example always returns true
	return true
}

func main() {
	router := gin.Default()

	// WebSocket endpoint
	router.GET("/ws", handleWebSocket)

	// Serve static files from the "static" directory
	router.Static("/static", "./static")

	// Set up a route to render the HTML page
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Run the web server on port 8078
	router.Run(":8078")
}
