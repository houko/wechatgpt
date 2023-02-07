FROM golang:1.19-alpine

ENV api_key=""

ARG target=/app
ARG entrypoint=docker/docker-entrypoint.sh

RUN export GOPRIVATE=github.com/WhaleSu/wechatgpt

WORKDIR ${target}

COPY bootstrap/ ${target}/bootstrap
COPY config/ ${target}/config
COPY handler/ ${target}/handler
COPY local/ ${target}/local
COPY openai/ ${target}/openai
COPY utils/ ${target}/utils
COPY go.mod go.sum main.go ${target}/

RUN go env -w GOPROXY=https://goproxy.cn,direct && go env -w GO111MODULE=on && go mod download && go build -o server main.go

CMD ./server

COPY ${entrypoint} /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT [ "docker-entrypoint.sh" ]
