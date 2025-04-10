# Go Version
ARG GO_VERSION=1.24

# Build
FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY ./ ./
RUN CGO_ENABLED=0 go build -o /app ./main.go

# Image
FROM gcr.io/distroless/static-debian12 AS production
USER nonroot:nonroot
COPY --from=build --chown=nonroot:nonroot /app /app
ENTRYPOINT ["/app"]