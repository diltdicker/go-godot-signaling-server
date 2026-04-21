extends CanvasLayer

class_name PrivateMultiplayerMenu

enum MultiplayerMode {HOST_SETUP, JOIN_SETUP, HOSTING, JOINING}

signal start_private_hosting(max_players: int)
signal start_private_game()
signal start_private_joining(code: String)
signal leave_private_lobby()
signal exit_menu()


var cur_state: MultiplayerMode

func update_player_cnt(cur_player_cnt: int, max_player_cnt: int):
	pass

func update_host_code(lobby_code: String):
	$%HostCodeEdit.text = lobby_code.to_upper()

func update_join_status(status_txt: String):
	pass

func _reset_host_mode():
	cur_state = MultiplayerMode.HOST_SETUP
	$%HostCheckButton.disabled = false
	$%JoinCheckButton.disabled = false
	$%HostVbox.visible = true
	$%JoinVbox.visible = false
	$%HostCodeEdit.text = ""
	$%HostCurrentPlayerCount.text = ""
	$%JoinCodeEdit.text = ""
	$%JoinCurrentPlayerCount.text = ""
	$%JoinStatusLabel.text = ""
	$%MaxPlayerSpin.editable = true
	$%StartButton.disabled = true
	$%JoinButton.visible = true
	$%HostButton.visible = true
	$%StartButton.disabled = true
	$%HostCancelButton.visible = false
	$%JoinCancelButton.visible = false

func _reset_join_mode():
	cur_state = MultiplayerMode.JOIN_SETUP

# Called when the node enters the scene tree for the first time.
func _ready() -> void:
	_reset_host_mode()


# Called every frame. 'delta' is the elapsed time since the previous frame.
#func _process(delta: float) -> void:
	#pass


func _on_join_button_pressed() -> void:
	cur_state = MultiplayerMode.JOINING
	emit_signal("start_private_joining", str($%JoinCodeEdit.text))


func _on_host_button_pressed() -> void:
	cur_state = MultiplayerMode.HOSTING
	emit_signal("start_private_hosting", $%MaxPlayerSpin.value)
	$%MaxPlayerSpin.editable = false
	
	$%StartButton.disabled = false


func _on_start_button_pressed() -> void:
	emit_signal("start_private_game")


func _on_join_check_button_pressed() -> void:
	print("join check")
	$%JoinCheckButton.button_pressed = true
	$%HostCheckButton.button_pressed = false
	$%HostVbox.visible = false
	$%JoinVbox.visible = true


func _on_host_check_button_pressed() -> void:
	print("host check")
	$%JoinCheckButton.button_pressed = false
	$%HostCheckButton.button_pressed = true
	$%HostVbox.visible = true
	$%JoinVbox.visible = false


func _on_exit_button_pressed() -> void:
	emit_signal("exit_menu")


func _on_join_cancel_button_pressed() -> void:
	emit_signal("leave_private_lobby")


func _on_host_cancel_button_pressed() -> void:
	emit_signal("leave_private_lobby")
