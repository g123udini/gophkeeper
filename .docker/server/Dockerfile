FROM golang:1.25

WORKDIR /app

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
RUN go mod download

COPY ./cmd/server ./cmd/server
COPY ./internal/server ./internal/server
COPY ./internal/common ./internal/common

RUN CGO_ENABLED=0 GOOS=linux go build -C ./cmd/server -o server

CMD ["/app/cmd/server/server"]
