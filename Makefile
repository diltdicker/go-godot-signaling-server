BIN_NAME=rtc_server
DOCKER_NAME=rtc-server

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

# Declare phony targets to prevent conflicts with files of the same name
.PHONY: docker-kill docker-clean docker-build docker-run

# 1. Kill the running container
docker-kill:
	@echo "Stopping and removing container..."
	-docker stop $(DOCKER_NAME) 2>/dev/null || true
	-docker rm $(DOCKER_NAME) 2>/dev/null || true

# 2. Clean the docker container (and remove the image)
docker-clean: docker-kill
	@echo "Removing docker image..."
	-docker rmi $(DOCKER_NAME) 2>/dev/null || true

# 3. Build a new docker image
docker-build: docker-clean
	@echo "Building new image..."
	docker build -t $(DOCKER_NAME) .

# 5. Tail the docker logs
docker-logs:
	docker logs -f $(DOCKER_NAME)

# 4. Run the docker image
docker-run: docker-build
	@echo "Running container..."
	docker run -d --name $(DOCKER_NAME) -p 8080:8080 $(DOCKER_NAME)

docker-size:
	@echo "Image size for $(DOCKER_NAME):"
	@docker image ls $(DOCKER_NAME) --format "{{.Size}}"


godot-clean:
	-rm exports/linux/*
	-rm exports/windows/*
	-rm exports/web/*