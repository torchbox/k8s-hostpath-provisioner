FROM golang:1.13.5-alpine as builder
RUN apk add --no-cache git
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY ./hostpath-provisioner.go ./hpp.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static" -s -w' -o ./hpp

FROM alpine
COPY --from=builder /build/hpp /
ENTRYPOINT [ "/hpp" ]
