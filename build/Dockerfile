FROM golang:1.25-alpine as go-build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go

FROM oven/bun:alpine as web-build-stage
WORKDIR /app
COPY web/package.json web/bun.lock* ./
RUN bun install --frozen-lockfile
COPY web/ ./
RUN bun run build

FROM scratch
WORKDIR /root/
COPY --from=go-build-stage /app/server .
COPY --from=web-build-stage /app/public ./web/public

EXPOSE 8080

CMD ["./server"]
