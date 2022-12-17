FROM golang:1.19-alpine

ENV api_key=""

# RUN export GOPRIVATE=github.com/houko/wechatgpt

WORKDIR /app

COPY . /app

ENV GO111MODULE=on
# ENV GOPROXY=https://goproxy.cn
ENV GOPROXY=https://goproxy.cn,direct


RUN go mod download && go build -o server main.go

CMD ./server
