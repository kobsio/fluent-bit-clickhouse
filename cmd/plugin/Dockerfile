# We can't go past 1.20.2 until this issue is solved: https://github.com/golang/go/issues/62130#issuecomment-1687335898
FROM golang:1.23.4 as build
WORKDIR /root
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build-plugin

FROM fluent/fluent-bit:3.2.4
COPY --from=build /root/out_clickhouse.so /fluent-bit/bin/
EXPOSE 2020
CMD ["/fluent-bit/bin/fluent-bit", "--plugin", "/fluent-bit/bin/out_clickhouse.so", "--config", "/fluent-bit/etc/fluent-bit.conf"]
