package core

import (
	"encoding/json"
	"sync"
)

type Lobby struct {
	Meta      json.RawMessage
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
	// If MatchType is 4 bytes, Go adds 4 bytes of padding here.
}

func (l *Lobby) AddUser(u *User) {
	l.Mu.Lock()
	defer l.Mu.Unlock()

	l.Peers = append(l.Peers, u)
}

func (l *Lobby) KickUser(id int32) {
	l.Mu.Lock()
	defer l.Mu.Lock()

	for i, u := range l.Peers {
		if u.Id == id {
			l.Peers[i] = l.Peers[len(l.Peers)-1]

			l.Peers[len(l.Peers)-1] = nil

			l.Peers = l.Peers[:len(l.Peers)-1]
			return
		}
	}
}

func (l *Lobby) Destruct() {
	l.Mu.Lock()
	defer l.Mu.Unlock()

	for i := range l.Peers {
		l.Peers[i] = nil
	}
}
