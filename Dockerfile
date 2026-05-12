# ── Stage 1: build React ─────────────────────────────────────────────────────
FROM node:20-alpine AS ui-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# ── Stage 2: build Go binary ──────────────────────────────────────────────────
FROM golang:1.26-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Embed the React build
COPY --from=ui-builder /app/web/dist ./cmd/server/ui
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o snippr ./cmd/server

# ── Stage 3: minimal runtime ─────────────────────────────────────────────────
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /app/snippr .
EXPOSE 8080
CMD ["./snippr"]
