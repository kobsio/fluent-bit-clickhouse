# We can't go past 1.20.2 until this issue is solved: https://github.com/golang/go/issues/62130#issuecomment-1687335898
FROM golang:1.21.5-alpine3.17 as build
RUN apk update && apk add git make
WORKDIR /ingester
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build-ingester

FROM alpine:3.21.2
RUN apk update && apk add --no-cache ca-certificates
RUN mkdir /ingester
COPY --from=build /ingester/ingester /ingester
WORKDIR /ingester
USER nobody
ENTRYPOINT  [ "/ingester/ingester" ]
