from golang:1.13 as builder

WORKDIR $GOPATH/src/github.com/shanehowearth/ratelimiter/limiter
ADD . $GOPATH/src/github.com/shanehowearth/ratelimiter/limiter

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -v -o /go/bin/ratelimiter cmd/main.go

from scratch

WORKDIR /root/
COPY --from=0 /go/bin/ratelimiter .

ENTRYPOINT ["./ratelimiter"]

