BIN_NAME=rtc_server

build:
	cd go && go build -o ../out/$(BIN_NAME) main.go

run: build
	./out/$(BIN_NAME)

clean:
	go clean
	rm out/*

test:
	cd tests && source venv/bin/activate && python3 client_test.py

# push main code to example game for live testing
update-addons:
	rm -rf example/example-game/addons
	cp -r godot/addons example/example-game/

# pull live edits back to main code
clone-addons:
	rm -rf godot/addons
	cp -r  example/example-game/addons godot