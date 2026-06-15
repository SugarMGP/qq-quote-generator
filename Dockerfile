# ---- build stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /qq-quote-go .

# ---- runtime stage ----
FROM alpine:3.19

RUN apk add --no-cache \
    chromium \
    nss \
    freetype \
    harfbuzz \
    ttf-freefont \
    font-noto-cjk \
    && rm -rf /var/cache/apk/*

ENV ROD_BROWSER_BIN=/usr/bin/chromium-browser

COPY --from=builder /qq-quote-go /usr/local/bin/qq-quote-go

EXPOSE 5000

ENV POOL_SIZE=4
ENV PORT=5000

ENTRYPOINT ["qq-quote-go"]
