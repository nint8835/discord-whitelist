FROM golang:1.18-stretch AS go-builder

WORKDIR /build

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o discord-whitelist .

FROM gcr.io/distroless/static

WORKDIR /app
COPY --from=go-builder /build/discord-whitelist ./discord-whitelist
ENTRYPOINT ["/app/discord-whitelist"]
