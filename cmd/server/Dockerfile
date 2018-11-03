FROM golang:1.11

WORKDIR $GOPATH/src/github.com/haakonleg/go-e2ee-chat-engine
COPY . .

RUN GO111MODULE="on" go build -o server cmd/server/main.go

CMD ["./server"]
