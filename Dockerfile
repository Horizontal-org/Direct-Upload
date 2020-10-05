FROM golang:latest as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/direct-upload .


FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/bin/direct-upload /bin/direct-upload
COPY --from=builder /app/docker-config.yaml config.yaml

EXPOSE 8080

VOLUME [ "/data" ]

ENTRYPOINT [ "direct-upload" ]
