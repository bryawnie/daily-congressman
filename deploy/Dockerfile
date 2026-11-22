# One image per binary: docker build --build-arg BINARY=jobs -f deploy/Dockerfile .
FROM golang:1.27-alpine AS build
ARG BINARY=jobs
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/${BINARY}

FROM alpine:3.22
RUN apk add --no-cache ca-certificates \
    && adduser -D -u 10001 app \
    && mkdir /data && chown app /data
USER app
ENV DB_PATH=/data/congressmen.db
VOLUME /data
COPY --from=build /out/app /usr/local/bin/app
ENTRYPOINT ["app"]
