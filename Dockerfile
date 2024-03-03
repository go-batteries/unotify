FROM golang:1.21 as builder

ENV ENVIRONMENT=prod

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /src/server cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /src/worker cmd/worker/main.go

FROM alpine:3.14

ENV ENVIRONMENT=prod

WORKDIR /opt/app

COPY --from=builder /src/server /opt/app/server
COPY --from=builder /src/worker /opt/app/worker
COPY --from=builder /src/bootstrap.sh /opt/app/bootstrap.sh
COPY --from=builder /src/config/app.env /opt/app/config/app.env
COPY --from=builder /src/config/statemachines /opt/app/config/statemachines

# COPY --from=builder /src/openapiv2 /opt/app/openapiv2

RUN chmod +x "/opt/app/server"
RUN chmod +x "/opt/app/worker"
RUN chmod +x "/opt/app/bootstrap.sh"

EXPOSE 9091
EXPOSE 9093

CMD ["/opt/app/bootstrap.sh"]
