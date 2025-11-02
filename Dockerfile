# syntax=docker/dockerfile:1.6

FROM golang:1.22 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
    go build -trimpath -ldflags="-s -w" -o /out/job-runner ./cmd/job-runner

FROM gcr.io/distroless/static
COPY --from=build /out/job-runner /job-runner
EXPOSE 8080
ENTRYPOINT ["/job-runner"]
