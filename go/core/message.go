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
	// --- 24-BYTE FIELDS (Slice headers: Data, Len, Cap) ---
	Offer  json.RawMessage `json:"offer,omitempty"`
	Answer json.RawMessage `json:"answer,omitempty"`
	Media  json.RawMessage `json:"media,omitempty"`
	SDP    json.RawMessage `json:"sdp,omitempty"`
	Meta   json.RawMessage `json:"meta,omitempty"`

	// --- 16-BYTE FIELDS (String headers: Data, Len) ---
	LobbyCode string `json:"lobbyCode,omitempty"`
	GameId    string `json:"gameId,omitempty"`

	// --- 8-BYTE FIELDS (Pointers and int64) ---
	IsPublic   *bool  `json:"isPublic,omitempty"`
	IsMesh     *bool  `json:"isMesh,omitempty"`
	LobbyAlive *bool  `json:"lobbyAlive,omitempty"`
	IsHost     *bool  `json:"isHost,omitempty"`
	Index      *int32 `json:"index,omitempty"`

	// --- 4-BYTE FIELDS (Grouped to fill 8-byte slots) ---
	Id     int32 `json:"id,omitempty"`
	ToId   int32 `json:"toId,omitempty"`
	FromId int32 `json:"fromId,omitempty"`
	PeerId int32 `json:"peerId,omitempty"`

	// --- 1-BYTE FIELDS (Grouped at the very end) ---
	MaxPeers int8 `json:"maxPeers,omitempty"`

	// Total Padding added by Go: ~7 bytes at the end to round to 8.
	// Previous "Mixed" version likely had 16-24 bytes of internal padding.
}
