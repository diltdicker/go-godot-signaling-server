# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org.
#
# Copyright (c) 2026 Dillon Dickerson
#
@tool
extends EditorPlugin


func _enter_tree() -> void:
	add_custom_type("PeerToPeerClient", "Node", preload("res://addons/gaming_rtc_client/p2p_multiplayer.gd"), preload("res://addons/gaming_rtc_client/gaming_rtc_icon.svg"))


func _exit_tree() -> void:
	remove_custom_type("PeerToPeerClient")
