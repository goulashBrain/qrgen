ARG GO_VERSION=1.27
FROM golang:${GO_VERSION} AS build
ARG VERSION=dev
ARG COMMIT=none
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
    -o /qrgen ./cmd/qrgen

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /qrgen /qrgen
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/qrgen"]