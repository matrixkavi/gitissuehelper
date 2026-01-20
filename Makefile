all: build

build:
	go mod download
	go build -o gitissuehelper

run: build
	./gitissuehelper create -h

clean:
	rm -f gitissuehelper

.PHONY: all build run clean
