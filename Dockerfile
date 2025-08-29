FROM golang:1.25 as base

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

FROM base as builder
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o server -ldflags="-w -s" ./cmd/server

# Build the admin CLI application
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o admin -ldflags="-w -s" ./cmd/admin

# Build golang-migrate
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

FROM alpine:latest

RUN mkdir /app
WORKDIR /app

# Copy the main server application
COPY --from=builder /app/server .

# Copy the admin CLI application
COPY --from=builder /app/admin .

# Copy golang-migrate binary
COPY --from=builder /go/bin/migrate .

# Copy migration files
COPY migrations ./migrations

# Create startup script
COPY start.sh ./start.sh

RUN chmod +x ./server ./admin ./migrate ./start.sh

ENTRYPOINT ["./start.sh"]
