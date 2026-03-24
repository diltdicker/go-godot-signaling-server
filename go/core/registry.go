package core

import (
	"sync"

	"github.com/diltdicker/go-godot-signaling-server/go/utils"
)

type Registry struct {
	mu     sync.RWMutex     // 24-32 bytes
	Users  map[int32]*User  // 8 bytes (Pointer to map)
	Lobbys map[int64]*Lobby // 8 bytes (Pointer to map)
}

func NewRegistry() *Registry {
	return &Registry{
		// Initializing with a small capacity (e.g., 100) saves CPU
		// by preventing early resizing without wasting much RAM.
		Users:  make(map[int32]*User, 100),
		Lobbys: make(map[int64]*Lobby, 100),
	}
}

func (r *Registry) UserExists(id int32) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.Users[id]
	return ok
}

func (r *Registry) AddUser(u *User) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Users[u.Id] = u
}

func (r *Registry) GetUser(id int32) (*User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.Users[id]
	return u, ok
}

func (r *Registry) DeleteUser(id int32) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Deleting a key that doesn't exist is safe in Go (it does nothing)
	delete(r.Users, id)
}

func (r *Registry) DisconnectUser(userId int32) {
	// 1. Fetch user and remove from global registry immediately
	// This prevents any new messages from being routed to a dying connection
	r.mu.Lock()
	u, ok := r.Users[userId]
	if ok {
		delete(r.Users, userId)
	}
	r.mu.Unlock()

	if !ok {
		return // User already gone
	}

	// 2. Cleanup Lobby State
	if u.CurLobby != 0 {
		if l, ok := r.GetLobby(u.CurLobby); ok {
			// Use the thread-safe RemovePeer we built
			peerId, remaining, removed := l.RemovePeer(u.Id)

			if removed {
				// If the user was the host, the lobby should likely die
				if u.IsHost {
					// Notify others the lobby is closing
					for _, p := range remaining {
						p.SendMessage(KICK, &WsDataMessage{
							Id:         peerId,
							LobbyAlive: utils.BoolPtr(false),
						})
						p.Mu.Lock()
						p.CurLobby = 0
						p.Mu.Unlock()
					}
					r.DeleteLobby(l.Id)
				} else {
					// Just notify others that this specific peer left
					for _, p := range remaining {
						p.SendMessage(KICK, &WsDataMessage{Id: peerId})
					}
				}
			}
		}
	}

	// 3. Final Resource Release
	u.Mu.Lock()
	if u.Conn != nil {
		u.Conn.Close()
	}
	// Null out the connection to ensure GC can reclaim the websocket buffers
	u.Conn = nil
	u.Mu.Unlock()
}

func (r *Registry) GetLobby(id int64) (*Lobby, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	l, ok := r.Lobbys[id]
	return l, ok
}

func (r *Registry) addLobby(l *Lobby) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Lobbys[l.Id] = l
}

func (r *Registry) CreateLobby() (l *Lobby) {
	id := utils.GenerateLobbyId()
	for {
		if _, ok := r.GetLobby(id); ok {
			id = utils.GenerateLobbyId()
		} else {
			break
		}
	}
	lo := &Lobby{
		Id:        id,
		LobbyCode: utils.IdToString(id),
		Peers:     []*User{},
		IsOpen:    true,
	}
	return lo
}

func (r *Registry) CreateAddLobby() (l *Lobby) {
	id := utils.GenerateLobbyId()
	for {
		if _, ok := r.GetLobby(id); ok {
			id = utils.GenerateLobbyId()
		} else {
			break
		}
	}
	lo := &Lobby{
		Id:        id,
		LobbyCode: utils.IdToString(id),
		Peers:     []*User{},
		IsOpen:    true,
	}

	r.addLobby(lo)

	return lo
}

func (r *Registry) DeleteLobby(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if l, ok := r.Lobbys[id]; ok {
		l.Destruct()
		delete(r.Lobbys, id)
	}

}
