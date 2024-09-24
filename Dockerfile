FROM golang:1.23.0-bookworm AS build

ARG upx_version=4.2.4

RUN apt-get update && apt-get install -y --no-install-recommends xz-utils && \
  curl -Ls https://github.com/upx/upx/releases/download/v${upx_version}/upx-${upx_version}-amd64_linux.tar.xz -o - | tar xvJf - -C /tmp && \
  cp /tmp/upx-${upx_version}-amd64_linux/upx /usr/local/bin/ && \
  chmod +x /usr/local/bin/upx && \
  apt-get remove -y xz-utils && \
  rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
# COPY operator_helm_packages /operator_helm_packages
# COPY ansible /ansible


RUN go mod download && go mod verify

# COPY *.go ./
COPY . .

RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o server -a -ldflags="-s -w" -installsuffix cgo
RUN upx --ultra-brute -qq server && upx -t server
FROM scratch

COPY --from=build /app/server /server

ENTRYPOINT ["/server"]

# FROM golang:1.22.3-alpine3.19
# WORKDIR /app

# COPY go.mod ./
# COPY go.sum ./
# COPY main.go ./

# RUN go build -o /db-dump-dwonload-helpler

# ENTRYPOINT [ "/db-dump-dwonload-helpler" ]