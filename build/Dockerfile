# FROM golang:latest AS builder

# WORKDIR /go/src/app

# COPY . .

# RUN go mod download

# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# FROM scratch

# COPY --from=builder /go/src/app/app /

# USER nobody:nobody

# CMD ["/app"]


FROM golang:alpine3.20 as go-build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server cmd/server/main.go


FROM scratch
# RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=go-build-stage /app/server .
COPY ./web ./web

EXPOSE 8080

CMD ["./server"]
