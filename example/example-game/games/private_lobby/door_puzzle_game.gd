extends Node

@onready var p2p_client: PeerToPeerClient = $%PeerToPeerClient
var multiplayer_menu: PrivateMultiplayerMenu = null

# Called when the node enters the scene tree for the first time.
func _ready() -> void:
	var multiplayer_menu_scene = load("res://games/private_lobby/PrivateMultiplayerMenu.tscn")
	multiplayer_menu = multiplayer_menu_scene.instantiate()
	add_child(multiplayer_menu)
	multiplayer_menu.visible = false
	multiplayer_menu.exit_menu.connect(_exit_multiplayer)

func _exit_multiplayer() -> void:
	$%GameMenu.visible = true
	multiplayer_menu.visible = false
	
	# disconnect multiplayer
	p2p_client.end_multiplayer()
	# disconnect from signalling server
	p2p_client.disconnect_from_server()

func _on_new_game_button_pressed() -> void:
	$%GameMenu.visible = false
	multiplayer_menu._reset_host_mode()
	multiplayer_menu.visible = true
	
	# connect to signalling server and setup multiplayer
	#p2p_client.connect_to_server("ws://127.0.0.1:8080/ws") 
	# defined in the override.cfg file
	p2p_client.connect_to_server(ProjectSettings.get_setting("network/rtc/signal_server_url")) 

func _on_peer_to_peer_client_socket_connected() -> void:
	print("p2p_client_connected")


func _on_peer_to_peer_client_socket_disconnected(code: int, reason: String) -> void:
	print("p2p_client_disconnected ", str(code), reason)


func _on_peer_to_peer_client_socket_error(err_code: int, err_message: String) -> void:
	print("p2p_client_error ", str(err_code), err_message)
