FROM golang:1.19-alpine

ENV api_key=""

RUN export GOPRIVATE=github.com/WhaleSu/wechatgpt

WORKDIR /app

COPY . /app

RUN go env -w GOPROXY=https://goproxy.cn,direct && go env -w GO111MODULE=on && go mod download && go build -o server main.go

CMD ./server
