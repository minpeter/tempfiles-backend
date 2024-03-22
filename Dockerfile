# Step 1: Modules caching
FROM --platform=$BUILDPLATFORM golang:1.19.4-alpine as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Builder
FROM --platform=$BUILDPLATFORM golang:1.19.4-alpine AS builder
COPY --from=modules /go/pkg /go/pkg
COPY . /app
ENV CGO_ENABLED=0
WORKDIR /app
ARG TARGETOS TARGETARCH 
ENV CGO_ENABLED=0
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /app/server .

# GOPATH for scratch images is /
FROM scratch
WORKDIR /app
COPY --from=builder /app/server /app/server
EXPOSE 5000
CMD ./server