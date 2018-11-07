FROM golang:1.11-alpine

WORKDIR $GOPATH/src/github.com/haakonleg/go-e2ee-chat-engine
COPY . .

RUN apk update && apk upgrade && \
    apk add --no-cache git && \
    GO111MODULE="on" CGO_ENABLED=0 go build -o chat_server cmd/server/main.go

CMD ["./chat_server"]
