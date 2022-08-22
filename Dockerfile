ARG GO_VERSION=1.18

# STAGE 1: building the executable
FROM golang:${GO_VERSION}-alpine AS build
RUN apk add --no-cache git
RUN apk --no-cache add ca-certificates

WORKDIR /src
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./

# Build the executable
RUN CGO_ENABLED=0 go build \
	-installsuffix 'static' \
	-o /app .

# STAGE 2: build the container to run
FROM gcr.io/distroless/static AS final

# copy compiled app
COPY --from=build --chown=nonroot:nonroot /app /app

# run binary; use vector form
ENTRYPOINT ["/app"]