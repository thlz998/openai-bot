FROM golang:1.19-alpine

ADD . /code

RUN export GO111MODULE=on && \
    export GOPROXY=https://goproxy.cn && \
    cd /code && \
    go mod download && \
    go build -o server main.go && \
    mkdir -p /app && \
    mv /code/server /app && \
    rm -rf /code

WORKDIR /app
CMD ./server