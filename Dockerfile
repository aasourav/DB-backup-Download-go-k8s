FROM golang:1.22.3-alpine3.19
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY main.go ./

RUN go build -o /db-dump-dwonload-helpler

ENTRYPOINT [ "/db-dump-dwonload-helpler" ]