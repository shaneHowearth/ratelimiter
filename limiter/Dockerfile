from golang:1.13 as builder

WORKDIR $GOPATH/src/github.com/shanehowearth/ratelimiter/limiter
ADD . $GOPATH/src/github.com/shanehowearth/ratelimiter/limiter

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -v -o /go/bin/ratelimiter cmd/main.go

from scratch
COPY init.sql /docker-entrypoint-initdb.d/
# ENV DBURI=postgres://accesswriter:hu8jmn3@access:5432/access_postgres_db
# ENV PORTNUM=80
# ENV LIMIT=100
# ENV TIMESPAN=1h

WORKDIR /root/
COPY --from=0 /go/bin/ratelimiter .

ENTRYPOINT ["./ratelimiter"]

