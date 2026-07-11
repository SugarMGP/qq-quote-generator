FROM rust:1.87-alpine3.21 AS resvg-builder

RUN apk add --no-cache git musl-dev
WORKDIR /src
COPY native/resvg/build.sh native/resvg/build.sh
RUN sh native/resvg/build.sh

FROM golang:1.25-alpine AS go-builder

RUN apk add --no-cache gcc musl-dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=resvg-builder /src/native/resvg/lib/linux-amd64/libresvg.a native/resvg/lib/linux-amd64/libresvg.a
COPY --from=resvg-builder /src/resvg.h resvg.h
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /qq-quote-go .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates fontconfig font-noto-cjk
COPY --from=go-builder /qq-quote-go /usr/local/bin/qq-quote-go

EXPOSE 5000
ENV PORT=5000
ENTRYPOINT ["qq-quote-go"]
