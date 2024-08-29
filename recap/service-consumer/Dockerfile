FROM golang:1.23 as builder
WORKDIR /opt/app
COPY . .
RUN CGO_ENABLED=0 go build -o client

FROM scratch
WORKDIR /opt/app
COPY --from=builder /opt/app/client ./client
COPY --from=builder /opt/app/certs ./certs
ENTRYPOINT ["/opt/app/client"]