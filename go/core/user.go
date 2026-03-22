package core

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

type User struct {
	Mu       sync.Mutex      // 8 bytes
	Conn     *websocket.Conn // 8 bytes
	GameId   string          // 16 bytes (String header)
	CurLobby int64
	Id       int32 // 4 bytes
	PeerId   int32
	IsHost   bool // 1 byte
	// 3 bytes of padding here (Total 40 bytes)
}

// Function for sending JSON messages to client
// All messages contain:
// code: the message protocol code
// data: arbitrary struct related to the protocol operation
func (u *User) SendBufMessage(code MCode, data any) {

	d, err := json.Marshal(&WsSendMsg{
		Code: code,
		Data: data,
	})
	if err != nil {
		slog.Error("Unable to marshal JSON", "data", data)
		return
	}

	err2 := u.Conn.WriteMessage(websocket.TextMessage, d)

	if err2 != nil {
		slog.Error("Unable to write message to socket", "user", u.Id, "data", d)
	}

}

func (u *User) SendMessage(code MCode, data any) {
	// 1. LOCK: Essential for WebSocket safety
	u.Mu.Lock()
	defer u.Mu.Unlock()

	// 2. Prepare the stream (NextWriter)
	// This tells Gorilla WebSocket: "I'm about to start pouring data into the pipe"
	w, err := u.Conn.NextWriter(websocket.TextMessage)
	if err != nil {
		slog.Error("Socket write error", "user", u.Id, "error", err)
		return
	}

	// 3. Encode directly to the stream
	// Instead of creating a giant []byte in RAM with Marshal,
	// we stream your WsSendMsg struct directly into the network buffer.
	if err := json.NewEncoder(w).Encode(&WsSendMsg{
		Code: code,
		Data: data,
	}); err != nil {
		slog.Error("JSON Encode error", "user", u.Id, "error", err)
	}

	// 4. Close the writer to FLUSH the data to the client
	w.Close()
}

func (u *User) CloseConnection(r *Registry) {
	// 1. Lock the user to prevent concurrent writes during closing
	u.Mu.Lock()
	if u.Conn != nil {
		u.Conn.Close() // Shuts down the TCP socket
		u.Conn = nil   // Help the GC by clearing the reference
	}
	u.Mu.Unlock()

	// 2. Remove from Registry (This makes the User "eligible" for destruction)
	r.DeleteUser(u.Id)

	// 3. Logic: If they were in a lobby, remove them there too
	// r.RemoveUserFromLobby(u)
}
