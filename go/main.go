package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Client connected!")

	// event loop
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read error: ", err)
			break
		}

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
