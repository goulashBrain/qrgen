ARG GO_VERSION=1.27
FROM golang:${GO_VERSION} AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /qrgen ./cmd/qrgen

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /qrgen /qrgen
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/qrgen"]
