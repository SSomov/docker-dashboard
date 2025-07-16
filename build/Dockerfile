FROM golang:alpine3.20 as go-build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go


FROM scratch
# RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=go-build-stage /app/server .
COPY ./web ./web

EXPOSE 8080

CMD ["./server"]
