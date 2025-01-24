FROM golang:1.19-alpine

ARG DIR=app
COPY . /${DIR}
WORKDIR /${DIR}

RUN go mod download
RUN go build -o main ./cmd/main.go

CMD go run ./main