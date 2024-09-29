FROM golang:1.22.1

WORKDIR /app

COPY go.mod go.sum /
RUN go mod download

COPY cmd/server/ /cmd/server/
COPY internal/ /internal/
COPY pkg/ /pkg/

RUN go build -o /server /cmd/server/*.go

EXPOSE 8888

CMD ["/server"]