ARG GO_VERSION=1.18

# STAGE 1: building the executable
FROM golang:${GO_VERSION}-alpine AS build
RUN apk add --no-cache git
RUN apk --no-cache add ca-certificates

WORKDIR /src
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./

ENV CLESSY_TOKEN="5763477340:AAFYInOfYsrsHdP64farFJcLyks3_0WuIkw"
ENV CLESSY_WEBHOOK="https://clessy-ng-production.up.railway.app/"
ENV CLESSY_DB_DIR="/db"
ENV CLESSY_EMOJI_PATH="/data/emojis"
ENV CLESSY_UNSPLASH_BG_PATH="/data/pics"
ENV CLESSY_UNSPLASH_FONT="/data/gill.ttf"
ENV CLESSY_MEME_FONT="/data/impact.ttf"
ENV CLESSY_SNAPCHAT_FONT="/data/source.ttf"

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