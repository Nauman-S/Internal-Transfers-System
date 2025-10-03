FROM docker.io/library/golang:1.25.1-alpine AS builder

WORKDIR /app

RUN apk --no-cache add ca-certificates make

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/tmp/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    make build

FROM docker.io/library/alpine:latest

ARG UID=1000
ARG GID=1000
RUN adduser -D -u $UID -g $GID payments-application
USER payments-application

COPY --from=builder /app/bin/server /app/bin/server

EXPOSE 8080

CMD ["/app/bin/server"]