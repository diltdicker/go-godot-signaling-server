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