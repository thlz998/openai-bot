FROM golang:1.19

ADD . /code

RUN apt update && \
    apt install -y make vim curl wget zip

RUN export GO111MODULE=on && \
    export GOPROXY=https://goproxy.cn && \
    cd /code && \
    make && \
    mkdir -p /app && \
    mv /code/dist/linux/main /app && \
    rm -rf /code

WORKDIR /app
CMD ./server