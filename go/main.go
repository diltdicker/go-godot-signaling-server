package main

import (
	"container/list"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Lobby struct {
	Game      string
	LobbyName string
	ShortDesc string
	LobbyCode string
	LobbyType uint8
	MaxPeers  uint8
	IsMesh    bool
	Tags      string
	IsActive  bool
	IsHidden  bool
	PeerList  list.List
}

type User struct {
	Id       int
	LobbyId  int
	IsHost   bool
	CurLobby Lobby
	Socket   *websocket.Conn
	Game     string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var users = make(map[int]*User) // Global
var userIdCounter int
var mu sync.Mutex // mutex for thread safety

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// Create a new User
	mu.Lock()
	userIdCounter++
	user := &User{
		Id:     userIdCounter,
		Socket: conn,
		// Initialize other fields as needed (e.g., Game from query params)
	}
	users[user.Id] = user
	mu.Unlock()

	fmt.Println("Client connected!")

	// Set initial read deadline (e.g., 60 seconds from now)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// event loop
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read error: ", err)
			// clean up disconnect
			mu.Lock()
			delete(users, user.Id)
			mu.Unlock()
			break
		}

		// Reset deadline on activity
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		fmt.Printf("Recieved: %s\n", string(p))

		if err := conn.WriteMessage(messageType, p); err != nil {
			fmt.Println("write error:", err)
			break
		}
	}

}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
