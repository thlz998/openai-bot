FROM golang:1.19

ADD . /code

RUN export GO111MODULE=on && \
    export GOPROXY=https://goproxy.cn && \
    cd /code && \
    make && \
    mkdir -p /app && \
    mv /code/dist/linux/main /app/server && \
    mkdir /app/data && \
    rm -rf /code

WORKDIR /app
CMD /app/server