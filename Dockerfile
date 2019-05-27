FROM golang:alpine AS builder

RUN apk add --update --no-cache git ca-certificates && update-ca-certificates 2>/dev/null

RUN mkdir /app
WORKDIR /app

ADD ./go.mod .
ADD ./go.sum .

RUN go mod download

ADD ./main.go .

RUN set -ex && \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
        -a -installsuffix cgo \
        -ldflags="-w -s" \
        -o /usr/bin/app

FROM scratch

LABEL maintainer="Romuald Bulyshko <opensource@bulyshko.com>"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/bin/app /usr/local/bin/app

EXPOSE 54321/udp

ENTRYPOINT ["app"]
