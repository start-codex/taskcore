FROM node:22-alpine AS frontend

WORKDIR /app/front
COPY front/package.json front/pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY front/ .
# output goes to ../ui/dist (relative to svelte.config.js = /app/ui/dist)
RUN pnpm build

# ---- Go build ----
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Bring in the built frontend so //go:embed all:dist finds the files.
COPY --from=frontend /app/ui/dist ./ui/dist
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

# ---- Final image ----
FROM alpine:3.21

RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /server /app/server

EXPOSE 8080
CMD ["/app/server"]
