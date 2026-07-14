# Build
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
# COPY go.sum ./            # add after `go mod tidy`
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 go build -o /out/fleet-manager ./cmd/fleet-manager

# Run
FROM gcr.io/distroless/static-debian12
COPY --from=build /out/fleet-manager /fleet-manager
EXPOSE 2112
ENTRYPOINT ["/fleet-manager"]
