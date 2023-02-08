all: clean build install

build:
	GOOS=linux GOARCH=amd64 go build -v main.go
	mkdir -p ./dist/linux
	mv main ./dist/linux/
	chmod +x ./dist/linux/
	go build -v main.go
	mkdir -p ./dist/mac
	mv main ./dist/mac/
	chmod +x ./dist/mac/

install:
	cd ./dist/ && tar cvfz chatgpt.tar.gz linux

clean:
	rm -rf ./dist