FROM golang:latest as build

WORKDIR /src

COPY gateway ./gateway
COPY go.mod go.sum main.go ./


RUN go get -d -v ./...
RUN go build -ldflags '-s -w' -o /gw main.go
RUN chmod +x /gw

# Now copy it into our base image.
FROM gcr.io/distroless/base
ARG ARCH=x86_64
COPY --from=build /gw /
COPY .cerbos/Linux_${ARCH}/cerbos /
COPY conf.default.yml /conf.yml

# Uncomment for testing with a disk storage
# define REMOTE_CERBOS_URL pointing to HTTP API (see CerbosServerFunctionAPI in sam.yml)
# then run `make test`
#COPY test/testdata/store /store
#COPY test/testdata/conf.yml test/testdata/verify_key.jwk /
ENTRYPOINT ["/gw"]
