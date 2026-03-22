package core

import (
	"encoding/json"
)

// u.SendMessage(CODE_LOBBY, &WsDataMessage{ Id: 123 })
type WsSendMsg struct {
	Data any   // 16 bytes (Interface header)
	Code MCode // 2-4 bytes (depending on MCode type)
}

type WsRecieveMsg struct {
	Data WsDataMessage // Large (Size of WsDataMessage)
	Code MCode         // 2-4 bytes
}

type ErrMessage struct {
	ErrReason string // 16 bytes
	ErrCode   int16  // 2 bytes
}

type WsDataMessage struct {
	// --- 24-BYTE FIELDS (Slices / json.RawMessage) ---
	Offer  json.RawMessage `json:"offer,omitempty"`
	Answer json.RawMessage `json:"answer,omitempty"`
	Media  json.RawMessage `json:"media,omitempty"`
	SDP    json.RawMessage `json:"sdp,omitempty"`
	Meta   json.RawMessage `json:"meta,omitempty"`

	// --- 16-BYTE FIELDS (Strings) ---
	LobbyCode string `json:"lobbyCode,omitempty"`
	GameId    string `json:"gameId,omitempty"`

	// --- 8-BYTE FIELDS (Pointers & int64) ---
	IsPublic   *bool `json:"isPublic,omitempty"`
	IsMesh     *bool `json:"isMesh,omitempty"`
	LobbyAlive *bool `json:"lobbyAlive,omitempty"`
	IsHost     *bool `json:"isHost,omitempty"`
	Index      int64 `json:"index,omitempty"`

	// --- 4-BYTE FIELDS (Packed together = 8 bytes) ---
	Id     int32 `json:"id,omitempty"`
	ToId   int32 `json:"toId,omitempty"`
	PeerId int32 `json:"peerId,omitempty"`

	// --- 1-BYTE FIELDS ---
	MaxPeers int8 `json:"maxPeers,omitempty"`

	// PADDING: Go will automatically add 7 bytes here at the end
	// to make the struct size a multiple of 8.
}
