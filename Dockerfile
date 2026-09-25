# syntax=docker/dockerfile:1
#
# The todo, as one static binary in a distroless image.
#
#   docker build --build-context platform=../platform -t todo .
#
# go.mod replaces github.com/kitsunium/platform with ../platform until the
# platform is published; the named build context provides that directory.
# The web app needs no Node stage: web/dist is committed and embedded.
#
# Configuration, with `docker run -e NAME=value`:
#   TODO_BASE_URL       where users reach the app, for the links in its mails
#                       (default http://localhost:4000)
#   KIT_SMTP_URL        smtp://user:password@host:587?tls=starttls|implicit|none;
#                       unset, mails are only captured and a warning says so —
#                       set it in production
#   TODO_ARCHIVE_AFTER  how long a done task stays listed (default 24h)
#   KIT_TRUST_PROXY=on  behind one reverse proxy: the client address is the last
#                       X-Forwarded-For hop, the one the proxy appended (sessions
#                       list it)

FROM golang:1.27.1 AS build
WORKDIR /src/todo
COPY --from=platform go.mod go.sum /src/platform/
COPY --from=platform kit /src/platform/kit
COPY --from=platform model /src/platform/model
COPY --from=platform analyzer /src/platform/analyzer
COPY --from=platform studio/studio.go /src/platform/studio/studio.go
COPY --from=platform studio/dist /src/platform/studio/dist
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/todo . \
    && mkdir -p /out/data

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/todo /todo
COPY --from=build --chown=nonroot:nonroot /out/data /data
ENV KIT_DATA_DIR=/data KIT_ADDR=:4000
VOLUME /data
EXPOSE 4000
USER nonroot:nonroot
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 CMD ["/todo", "healthcheck"]
ENTRYPOINT ["/todo"]
