FROM golang:alpine AS build
WORKDIR /src
ADD . .
RUN go build ./cmd/gitlab-slack-multiplexer

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=build /src/gitlab-slack-multiplexer /usr/local/bin
ENTRYPOINT ["/usr/local/bin/gitlab-slack-multiplexer"]
