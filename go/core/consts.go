package core

type MCode int8

const (
	ID        MCode = iota // 0:ID ([server] Initiates call. [user] Responds with game name (failure to respond will result in auto-disconnect))
	HOST                   // 1:HOST ([user] Initiates call to host game. [server] Responds with peer id and lobby code)
	JOIN                   // 2:JOIN ([user] Initiates call to join a lobby. [server] Responds with peer id)
	QUEUE                  // 3:QUEUE ([user] Initiates call to queue for a game. [server] Responds with peer id and lobby code)
	VIEW                   // 4:VIEW ([user] Initiates call to get details of lobby(s). [server] responds with lobby details and peer count)
	ADD                    // 5:ADD ([server] Initiates call to inform user of new peer connection)
	KICK                   // 6:RM ([server] Initiates call to inform user of peer disconnecting or lobby deletion. [user] initiates call with own id to signal intent to leave lobby)
	OFFER                  // 7:OFFER ([user] Initiates rtc offer to be relayed to desired user in lobby. [server] Relays call to desired user tagging the sending user's peer id)
	ANSWER                 // 8:ANSWER ([user] Initiates rtc answer to be relayed to desired user in lobby. [server] Relays call to desired user tagging the sending user's peer id)
	CANDIDATE              // 9:CANDIDATE ([user] Initiates rtc candidate to be relayed to desired user in lobby. [server] Relays call to desired user tagging the sending user's peer id)
	READY                  // 10:READY ([server] Initiates call to host to confirm all user connections. [user] Initiates call to server to send READY call to host (in case of not a queued lobby))
	START                  // 11:START ([user] Initiates call to server to disable lobby and close all user connections)
	ERR                    // 12:ERR ([server] Initiates call to inform user of server error)
	UPDATE                 // 13:UPDATE ([host] Initiaties call to updated maxPeer count or Meta of the Lobby). [server] replies to all the non-hosts with updated maxPeer count and Meta
	PING                   // 14:PING ([server] auto initiates call every 60 seconds. [user] sends back empty packet)
)

type MatchType int8

const (
	PRIVATE MatchType = iota // 0:PRIVATE not viewable without lobbyCode
	PUBLIC                   // 1:PUBLIC always viewable
	AUTO                     // 2:QUEUE lobby type auto checks for when full and auto-starts
)

const (
	StartGameCode    uint16 = 1000
	StartGameMsg     string = "Closing peer connection to start game"
	BadProtoCode     uint16 = 4005
	BadProtoMsg      string = "Recieved invalid message with unknown protocol"
	BadMsgCode       uint16 = 4022
	BadMsgMsg        string = "Received bad message Unable to process"
	LobbyMissingCode uint16 = 4004
	LobbyMissingMsg  string = "Lobby for given lobbyCode does not exist"
	BadViewCode      uint16 = 4000
	BadViewMsg       string = "Invalid message for viewing lobby"
	BadHostCode      uint16 = 4006
	BadHostMsg       string = "Invalid message for hosting lobby"
	BadJoinCode      uint16 = 4001
	BadJoinMsg       string = "Invalid message for joining lobby"
	BadQueueCode     uint16 = 4010
	BadQueueMsg      string = "Inavlid message for queueing"
	IdleSocketCode   uint16 = 4008
	IdleSocketMSg    string = "Idle socket connection for too long"
	UnknownErrCode   uint16 = 4017
	UnknownErrMsg    string = "Unknown error"
	UnknownPeerCode  uint16 = 4010
	UnknownPeerMsg   string = "Unknown peer"

	LobbyFullCode uint16 = 4023
	LobbyFullMsg  string = "Lobby is full"

	GameMismatchCode uint16 = 4009 // Mirrored from HTTP 409
	GameMismatchMsg  string = "Game ID does not match lobby"

	ServerBusyCode uint16 = 5003 // Mirrored from HTTP 503 (Service Unavailable)
	ServerBusyMsg  string = "Server at maximum capacity"

	RateLimitCode uint16 = 4029
	RateLimitMsg  string = "Message frequency too high (Rate Limited)"

	NotAllowedCode uint16 = 4003
	NotAllowedMsg  string = "Action not allowed"
)
