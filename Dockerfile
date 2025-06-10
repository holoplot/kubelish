FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build
ARG TARGETOS
ARG TARGETARCH

WORKDIR /go/src/github.com/holoplot/kubelish

# Only invalidate the download layer if the modules changed
COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build ./cmd/kubelish/main.go

FROM alpine:3.22

RUN apk add --update ca-certificates && rm -rf /var/cache/apk/*
COPY --from=build /go/src/github.com/holoplot/kubelish/main kubelish

ENTRYPOINT ["/kubelish"]
