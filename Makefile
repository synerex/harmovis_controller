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