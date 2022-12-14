FROM golang:1.19-alpine

ENV apiKey=""
ENV telegram=""
ENV wechat=""

RUN export GOPRIVATE=github.com/houko/wechatgpt

WORKDIR /app

COPY . /app

RUN go mod download && go build -o server main.go

CMD ./server