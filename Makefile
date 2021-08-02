# Makefile f

GOCMD=go
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
RM=rm


# Main target
build: harmovis_controller

harmovis_controller: harmovis_controller.go
	$(GOCMD) build


docker-image:
	docker build -t harmovis_controller .

docker-push:
	docker tag harmovis_controller:latest synerex/harmovis-demo:latest
	docker push synerex/harmovis-demo:latest

