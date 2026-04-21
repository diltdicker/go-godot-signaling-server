package core

import (
	"encoding/json"
	"sync"
)

type Lobby struct {
	Meta      json.RawMessage
	Backup    json.RawMessage
	Mu        sync.RWMutex // 24 bytes
	Peers     []*User      // 24 bytes
	LobbyCode string       // 16 bytes
	GameId    string
	Id        int64
	HostId    int32

	LobbyType MatchType // 4 or 8 bytes
	MaxPeers  int8
	IsMesh    bool
	IsOpen    bool
	IsOngoing bool
	// If MatchType is 4 bytes, Go adds 4 bytes of padding here.
}

func (l *Lobby) AddUser(u *User) {
	l.Mu.Lock()
	defer l.Mu.Unlock()

	l.Peers = append(l.Peers, u)
}

func (l *Lobby) Destruct() {
	l.Mu.Lock()
	defer l.Mu.Unlock()

	for i := range l.Peers {
		l.Peers[i] = nil
	}
}

// RemovePeer removes a user by ID and returns their PeerId and a slice of remaining peers.
// We return these so the caller can broadcast the REMOVE message without holding the lock.
func (l *Lobby) RemovePeer(userId int32) (peerId int32, remaining []*User, ok bool) {
	l.Mu.Lock()
	defer l.Mu.Unlock()

	for i, p := range l.Peers {
		if p.Id == userId {
			peerId = p.PeerId // Capture the "Virtual ID" (1 or Global)

			// --- SWAP AND POP (The Fastest Way) ---
			lastIdx := len(l.Peers) - 1

			// 1. Move the last person into the gap
			l.Peers[i] = l.Peers[lastIdx]

			// 2. CRITICAL: Null out the old last slot for the GC
			// If you don't do this, the User stays in RAM forever!
			l.Peers[lastIdx] = nil

			// 3. Shrink the slice
			l.Peers = l.Peers[:lastIdx]

			// 4. Create a snapshot for the broadcast loop
			remaining = make([]*User, len(l.Peers))
			copy(remaining, l.Peers)

			return peerId, remaining, true
		}
	}
	return 0, nil, false
}


func (l *Lobby) UpdateBackup(newData []byte) {
    l.Mu.Lock()
    defer l.Mu.Unlock()
    
    // Efficient reuse of the existing slice capacity
    if cap(l.Backup) >= len(newData) {
        l.Backup = l.Backup[:len(newData)]
        copy(l.Backup, newData)
    } else {
        l.Backup = json.RawMessage(newData)
    }
}