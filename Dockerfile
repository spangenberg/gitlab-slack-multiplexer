FROM golang:alpine AS build
RUN apk --no-cache add bash git make
WORKDIR /go/src/github.com/spangenberg/gitlab-slack-multiplexer
COPY . .
RUN make

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=build /go/src/github.com/spangenberg/gitlab-slack-multiplexer/bin/gitlab-slack-multiplexer /usr/local/bin
ENTRYPOINT ["/usr/local/bin/gitlab-slack-multiplexer"]
