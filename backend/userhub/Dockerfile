FROM golang:1.14-alpine as builder

COPY go.mod /service/
COPY go.sum /service/

WORKDIR /service

RUN go mod download

COPY statik/*.go ./statik/
COPY token/*.go ./token/
COPY common/*.go ./common/

COPY userhub/ ./userhub/

RUN GOOS=linux GOARCH=amd64 go build -o user-hub-service ./userhub/cmd/

FROM alpine

COPY --from=builder /service/user-hub-service /service/user-hub

WORKDIR /service

ENTRYPOINT ["./user-hub"]
