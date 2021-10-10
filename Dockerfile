# https://chemidy.medium.com/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
FROM golang:alpine AS builder


# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git


WORKDIR $GOPATH/src/github.com/lyrix-music/server


COPY . .
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /go/bin/lyrix-server


FROM scratch
COPY --from=builder /go/bin/lyrix-server /go/bin/lyrix-server
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/go/bin/lyrix-server", "/etc/lyrix/server/config.json"]
