PREFIX ?= /usr/local

build:
	go build .

install:
	sudo cp module $(PREFIX)/bin
