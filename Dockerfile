FROM golang:1.25 as base

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

FROM base as builder
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o main -ldflags="-w -s" .

# Build golang-migrate
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

FROM alpine:latest

RUN mkdir /app
WORKDIR /app

# Copy the main application
COPY --from=builder /app/main .

# Copy golang-migrate binary
COPY --from=builder /go/bin/migrate .

# Copy migration files
COPY migrations ./migrations

# Create startup script
COPY start.sh ./start.sh

RUN chmod +x ./start.sh

ENTRYPOINT ["./start.sh"]