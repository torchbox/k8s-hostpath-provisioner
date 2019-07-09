FROM golang:stretch as builder
WORKDIR /build
COPY ./go.mod ./go.mod
RUN go mod download
COPY ./hostpath-provisioner.go ./hpp.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o ./hpp

FROM alpine:3.10
COPY --from=builder /build/hpp /
ENTRYPOINT [ "/hpp" ]
