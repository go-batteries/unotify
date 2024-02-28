FROM golang:1.21 as builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /src/server cmd/server/main.go

FROM alpine:3.14

ENV ENV=prod

WORKDIR /opt/app

COPY --from=builder /src/server /opt/app/server
COPY --from=builder /src/bootstrap.sh /opt/app/bootstrap.sh
COPY --from=builder /src/config/app.env /opt/app/config/app.env
COPY --from=builder /src/openapiv2 /opt/app/openapiv2

RUN chmod +x "/opt/app/server"
RUN chmod +x "/opt/app/bootstrap.sh"

EXPOSE 9091
CMD ["/opt/app/bootstrap.sh"]
