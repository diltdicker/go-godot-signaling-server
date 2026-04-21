// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org.
//
// Copyright (c) 2026 Dillon Dickerson
package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/diltdicker/go-godot-signaling-server/go/core"
	"github.com/diltdicker/go-godot-signaling-server/go/utils"
	"github.com/gorilla/websocket"
)

// Constant Envs
const memLimitKey = "GGSS_MEM_LIMIT"
const idleKickKey = "GGSS_IDLE_KICK"
const longKickKey = "GGSS_LONG_KICK"
const maxUsersKey = "GGSS_MAX_USERS"

// Global Envs
var memLimit int   // maximum amount of memory in MiB for the Go process to use
var idleKick int   // the amount of idle time to wait (seconds) before kicking user from the server
var longKick int   // time to wait (seconds) before kicking user for being connected too long
var maxUsers int64 // maximum number of concurrent users

// Global Consts
const maxLobbySize int8 = 16

// Env read from system
func init() {
	memLimit = utils.GetEnvInt(memLimitKey, -1)    // default is off
	idleKick = utils.GetEnvInt(idleKickKey, -1)    // default is off
	longKick = utils.GetEnvInt(longKickKey, 60*60) // default is 60 minutes
}

// Global Vars
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024, // Shrink to 1 KB
	WriteBufferSize: 1024, // Shrink to 1 KB
	CheckOrigin:     func(r *http.Request) bool { return true },
}
var idPool = utils.NewIdPool(2)
var registry = core.NewRegistry()

// Global Errors - optimization no new mem allocation for error messages
var (
	// Generic Protocol Errors
	ErrBadProto   = &core.ErrMessage{ErrCode: int16(core.BadProtoCode), ErrReason: core.BadProtoMsg}
	ErrBadMessage = &core.ErrMessage{ErrCode: int16(core.BadMsgCode), ErrReason: core.BadMsgMsg}
	ErrUnknown    = &core.ErrMessage{ErrCode: int16(core.UnknownErrCode), ErrReason: core.UnknownErrMsg}

	// Lobby & Peer Errors
	ErrLobbyMissing = &core.ErrMessage{ErrCode: int16(core.LobbyMissingCode), ErrReason: core.LobbyMissingMsg}
	ErrLobbyFull    = &core.ErrMessage{ErrCode: int16(core.LobbyFullCode), ErrReason: core.LobbyFullMsg}
	ErrUnknownPeer  = &core.ErrMessage{ErrCode: int16(core.UnknownPeerCode), ErrReason: core.UnknownPeerMsg}

	// Action Specific Errors
	ErrBadView   = &core.ErrMessage{ErrCode: int16(core.BadViewCode), ErrReason: core.BadViewMsg}
	ErrBadHost   = &core.ErrMessage{ErrCode: int16(core.BadHostCode), ErrReason: core.BadHostMsg}
	ErrBadJoin   = &core.ErrMessage{ErrCode: int16(core.BadJoinCode), ErrReason: core.BadJoinMsg}
	ErrBadQueue  = &core.ErrMessage{ErrCode: int16(core.BadQueueCode), ErrReason: core.BadQueueMsg}
	ErrForbidden = &core.ErrMessage{ErrCode: int16(core.NotAllowedCode), ErrReason: core.NotAllowedMsg}

	// State/Validation Errors
	ErrGameMismatch = &core.ErrMessage{ErrCode: int16(core.GameMismatchCode), ErrReason: core.GameMismatchMsg}

	// Connection Management
	ErrIdleSocket = &core.ErrMessage{ErrCode: int16(core.IdleSocketCode), ErrReason: core.IdleSocketMSg}
	ErrRateLimit  = &core.ErrMessage{ErrCode: int16(core.RateLimitCode), ErrReason: core.RateLimitMsg}
	ErrServerBusy = &core.ErrMessage{ErrCode: int16(core.ServerBusyCode), ErrReason: core.ServerBusyMsg}
)

// Websocket Logic
// ===============
func handleMessage(u *core.User, p []byte) (err error) {
	// marshal json to WsMessage
	m := &core.WsRecieveMsg{}
	if err := json.Unmarshal(p, m); err != nil {
		slog.Error("Unable to unmarshal WsMessage from:", "user", u.Id)
		u.SendMessage(core.ERR, ErrBadMessage)
		return err
	}

	slog.Info("recieved msg:", "code", m.Code, "data", m.Data)

	switch m.Code {
	case core.PING:
		{
			// do nothing

		}
	case core.ID:
		{
			gameId := strings.TrimSpace(m.Data.GameId)
			if gameId == "" {
				u.SendMessage(core.ERR, ErrBadMessage)
				return
			}
			u.Mu.Lock()
			u.GameId = gameId // assign user's game id to profile
			u.Mu.Unlock()
		}

	case core.HOST:
		{
			// validate needed field zero values
			if u.GameId == "" || m.Data.MaxPeers <= 0 || m.Data.MaxPeers > maxLobbySize ||
				m.Data.IsPublic == nil || m.Data.IsMesh == nil {
				u.SendMessage(core.ERR, ErrBadMessage)
				return
			}
			l := registry.CreateAddLobby()

			l.AddUser(u) // uses lock

			l.Mu.Lock()
			lobbyId := l.Id
			l.HostId = u.Id
			l.GameId = u.GameId
			l.MaxPeers = m.Data.MaxPeers
			l.Meta = m.Data.Meta // blind copy
			if *m.Data.IsPublic == true {
				l.LobbyType = core.PUBLIC
			} else {
				l.LobbyType = core.PRIVATE
			}
			l.IsMesh = *m.Data.IsMesh
			l.Mu.Unlock()

			u.Mu.Lock()
			u.IsHost = true
			u.CurLobby = lobbyId
			u.PeerId = 1
			u.Mu.Unlock()

			resp := &core.WsDataMessage{
				Id:        1,
				LobbyCode: l.LobbyCode,
				IsMesh:    &l.IsMesh,
			}

			u.SendMessage(core.HOST, resp)
		}
	case core.JOIN:
		{
			// validate inputs
			if u.GameId == "" || m.Data.LobbyCode == "" {
				u.SendMessage(core.ERR, ErrBadMessage)
				return
			}
			lobbyId := utils.StringToId(m.Data.LobbyCode)
			l, ok := registry.GetLobby(lobbyId)
			if !ok {
				u.SendMessage(core.ERR, ErrLobbyMissing)
				return
			}

			l.Mu.Lock() // lobby lock
			if u.GameId != l.GameId {
				l.Mu.Unlock()
				u.SendMessage(core.ERR, ErrGameMismatch)
				return
			}
			if len(l.Peers) >= int(l.MaxPeers) {
				l.Mu.Unlock()
				u.SendMessage(core.ERR, ErrLobbyFull)
				return
			}

			// Snapshot current peers before adding self (for the "ADD" loop)
			peersSnapshot := make([]*core.User, len(l.Peers))
			copy(peersSnapshot, l.Peers)
			l.Peers = append(l.Peers, u) // add user to lobby

			l.Mu.Unlock() // lobby unlock

			u.Mu.Lock()
			u.IsHost = false
			u.CurLobby = lobbyId
			u.PeerId = u.Id
			u.Mu.Unlock()

			resp := &core.WsDataMessage{
				Id:        u.Id,
				IsMesh:    &l.IsMesh,
				LobbyCode: l.LobbyCode,
			}

			u.SendMessage(core.JOIN, resp)

			// The Broadcast Loop
			for _, peer := range peersSnapshot {

				// Tell the existing peer about the NEW user
				peer.SendMessage(core.ADD, &core.WsDataMessage{PeerId: u.Id})

				// Tell the NEW user about the existing peer
				u.SendMessage(core.ADD, &core.WsDataMessage{PeerId: peer.Id})
			}
		}
		// Queue lobbies are effectively normal public lobbies that will auto-start when the lobby fill up
	case core.QUEUE:
		{

		}
	// Ongoing is different from queue in that the users never disconnect from lobby until the game ends (empty lobby)
	// Ongoing lobbies are always public, game starts immediately with the first host and players join throughout the game
	// Don't use mesh network -> needs a BACKUP call
	case core.ONGOING:
		{

		}
	case core.VIEW:
		{

		}
	case core.OFFER:
		{
			// validate input
			if m.Data.Offer == nil || m.Data.ToId <= 0 || u.PeerId <= 0 {
				u.SendMessage(core.ERR, ErrBadMessage)
				return
			}

			peerId := m.Data.ToId
			if m.Data.ToId == 1 {
				if l, ok := registry.GetLobby(u.CurLobby); ok {
					peerId = l.HostId
				} else {
					u.SendMessage(core.ERR, ErrLobbyMissing)
					return
				}
			}

			peer, ok := registry.GetUser(peerId)
			if !ok {
				u.SendMessage(core.ERR, ErrUnknownPeer)
				return
			}

			if u.CurLobby != peer.CurLobby {
				u.SendMessage(core.ERR, ErrLobbyMissing)
				return
			}

			peer.SendMessage(core.OFFER, &core.WsDataMessage{Offer: m.Data.Offer, FromId: u.PeerId})

		}
	case core.ANSWER:
		{
			// validate input
			if m.Data.Answer == nil || m.Data.ToId <= 0 || u.PeerId <= 0 {
				u.SendMessage(core.ERR, ErrBadMessage)
				return
			}

			peerId := m.Data.ToId
			if m.Data.ToId == 1 {
				if l, ok := registry.GetLobby(u.CurLobby); ok {
					peerId = l.HostId
				} else {
					u.SendMessage(core.ERR, ErrLobbyMissing)
					return
				}
			}

			peer, ok := registry.GetUser(peerId)
			if !ok {
				u.SendMessage(core.ERR, ErrUnknownPeer)
				return
			}

			if u.CurLobby != peer.CurLobby {
				u.SendMessage(core.ERR, ErrLobbyMissing)
				return
			}

			peer.SendMessage(core.ANSWER, &core.WsDataMessage{Answer: m.Data.Answer, FromId: u.PeerId})
		}
	case core.CANDIDATE:
		{
			// validate input
			if m.Data.SDP == nil || m.Data.Media == nil || m.Data.Index == nil || m.Data.ToId <= 0 || u.PeerId <= 0 {
				u.SendMessage(core.ERR, ErrBadMessage)
				return
			}

			peerId := m.Data.ToId
			if m.Data.ToId == 1 {
				if l, ok := registry.GetLobby(u.CurLobby); ok {
					peerId = l.HostId
				} else {
					u.SendMessage(core.ERR, ErrLobbyMissing)
					return
				}
			}

			peer, ok := registry.GetUser(peerId)
			if !ok {
				u.SendMessage(core.ERR, ErrUnknownPeer)
				return
			}

			if u.CurLobby != peer.CurLobby {
				u.SendMessage(core.ERR, ErrLobbyMissing)
				return
			}

			peer.SendMessage(core.CANDIDATE, &core.WsDataMessage{
				SDP:    m.Data.SDP,
				Media:  m.Data.Media,
				Index:  m.Data.Index,
				FromId: u.PeerId,
			})
		}
	case core.KICK:
		{
			// validate input
			if m.Data.Id == 0 {
				u.SendMessage(core.ERR, ErrBadMessage)
				return
			}
			l, ok := registry.GetLobby(u.CurLobby)
			if !ok {
				u.SendMessage(core.ERR, ErrLobbyMissing)
				return
			}

			var t *core.User
			var deleteLobby = false

			if u.Id == m.Data.Id {
				t = u // user is leaving lobby
				if u.IsHost {
					deleteLobby = true
				}
			} else {
				if !u.IsHost {
					u.SendMessage(core.ERR, ErrForbidden)
					return
				}
				t, _ = registry.GetUser(m.Data.Id) // host is kicking other user
			}

			if t == nil || u.CurLobby != t.CurLobby {
				u.SendMessage(core.ERR, ErrUnknownPeer)
				return
			}

			// remove peer from lobby
			peerId, remaining, _ := l.RemovePeer(t.Id)

			if !t.IsHost && u.IsHost {
				// force kick by host
				t.SendMessage(core.KICK, &core.WsDataMessage{Id: t.Id, LobbyAlive: utils.BoolPtr(true)})
			}

			// update all users or the remmoved peer
			for _, p := range remaining {
				if deleteLobby {
					p.SendMessage(core.KICK, &core.WsDataMessage{Id: peerId, LobbyAlive: utils.BoolPtr(false)})
					p.Mu.Lock()
					p.CurLobby = 0
					p.IsHost = false
					p.PeerId = 0
					p.Mu.Unlock()
				} else {
					p.SendMessage(core.KICK, &core.WsDataMessage{Id: peerId, LobbyAlive: utils.BoolPtr(true)})
				}
			}

			// update removed peer status
			t.Mu.Lock()
			t.CurLobby = 0
			t.IsHost = false
			t.PeerId = 0
			t.Mu.Unlock()

			if deleteLobby {
				registry.DeleteLobby(l.Id)
			}
		}
	case core.READY:
		{

			l, ok := registry.GetLobby(u.CurLobby)
			if !ok {
				u.SendMessage(core.ERR, ErrLobbyMissing)
				return
			}

			if u.IsHost {
				initId := m.Data.Id // can be zero

				l.Mu.Lock() // lock the lobby

				peerCount := int8(len(l.Peers) - 1)
				var targets []*core.User

				if initId == 0 { // initId was unset
					// Broadcast to all non-hosts
					for _, p := range l.Peers {
						if !p.IsHost {
							targets = append(targets, p)
						}
					}
				} else { // initId was set
					// Target a specific peer who joined late
					for _, p := range l.Peers {
						if p.Id == initId {
							targets = append(targets, p)
							break
						}
					}
				}
				l.Mu.Unlock()

				// send the READY Message
				// Handle the "Slight Delay" using a Goroutine
				// We pass the targets and count to the goroutine
				go func(peers []*core.User, count int8) {
					time.Sleep(1 * time.Second) // delay to allow users to finish connecting
					for _, p := range peers {
						p.SendMessage(core.READY, &core.WsDataMessage{
							Id:       p.PeerId,
							MaxPeers: peerCount,
						})
					}
				}(targets, peerCount)

			} else {
				// Case B: Peer signaling the Host

				host, ok := registry.GetUser(l.HostId)
				if ok && host.CurLobby == u.CurLobby {
					// Forward the peer's ready status to the host
					host.SendMessage(core.READY, m.Data)
				}
			}
		}
	case core.START:
		{
			l, ok := registry.GetLobby(u.CurLobby)
			if !u.IsHost || !ok {
				u.SendMessage(core.ERR, ErrForbidden)
				return
			}

			l.Mu.Lock()
			l.IsOpen = false
			targets := make([]*core.User, len(l.Peers))
			copy(targets, l.Peers)
			l.Mu.Unlock()
			slices.Reverse(targets) // host is now last

			// Tell everyone to start game
			for _, p := range targets {
				if !p.IsHost {
					p.SendMessage(core.START, nil)
				}
			}

			time.Sleep(250 * time.Millisecond) // wait for everyone else to recieve START

			u.SendMessage(core.START, nil)

			// The Clean Exit
			go func(lobbyId int64, players []*core.User) {
				time.Sleep(500 * time.Millisecond) // Wait for WebRTC finalization
				for _, p := range players {
					registry.DisconnectUser(p.Id) // Atomic cleanup of everything
				}
				registry.DeleteLobby(lobbyId)
			}(l.Id, targets)
		}
	case core.UPDATE:
		{

		}
	default:
		{
			u.SendMessage(core.ERR, ErrBadProto)
		}
	}
	return nil // no err
}

// ===============

// Websocket Server Code
// =====================

func handler(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Upgrade error:", "error", err)
	}

	// setting connection details
	conn.SetReadLimit(32 * 1024)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start a heartbeat ticker for this specific user
	go func() {
		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return // Trigger closure if ping fails
			}
		}
	}()

	// assign new client id
	u := &core.User{
		Id:       idPool.Borrow(),
		CurLobby: 0,
		Conn:     conn,
	}
	registry.AddUser(u)
	slog.Info("Client connected", "id", u.Id)

	defer func() {
		slog.Info("Cleaning up user", "id", u.Id)

		registry.DisconnectUser(u.Id)
		idPool.Return(u.Id)
	}()

	// server asks user which game id
	u.SendMessage(core.ID, nil)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			// Loop breaks here! Function exits, defers run.
			break
		}

		if messageType == websocket.TextMessage {
			handleMessage(u, p)
		}
	}
}

func main() {
	// var _ = GetNextID(3499)
	http.HandleFunc("/ws", handler)
	// fmt.Println("Server started on :8080")
	slog.Info("Server started")
	http.ListenAndServe(":8080", nil)

	// println(utils.GenLobbyCode(34543))
}
