# syntax=docker/dockerfile:1
#
# The todo, as one static binary in a distroless image.
#
#   docker build -t todo .
#   docker run --rm -p 4000:4000 -v todo-data:/data todo
#
# The product imports the SDK alone, a published version with no replace: the
# build needs this directory and the module proxy, and no other checkout. The
# web app needs no build step: its files are embedded as they are written.
#
# The image runs the role `server`. Its configuration, with
# `docker run -e NAME=value` (the README lists every setting):
#   SERVER_BASE_URL          where members reach the product, for the links of
#                            its mails (default http://localhost:4000)
#   SERVER_SMTP_URL          smtp://user:pass@host:587?tls=starttls|implicit|none;
#                            unset, the mails are only captured: set it in
#                            production
#   SERVER_VERIFY_ADDRESSES  false where there is no mail relay (default true)
#   SERVER_TRUST_PROXY=on    behind one reverse proxy: the client address is
#                            the hop the proxy appended to X-Forwarded-For

FROM golang:1.27.3 AS build
WORKDIR /src/todo
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod GOWORK=off go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    GOWORK=off CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/todo ./cmd/todo \
    && mkdir -p /out/data

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/todo /todo
COPY --from=build --chown=nonroot:nonroot /out/data /data
ENV SERVER_DATA_DIR=/data SERVER_ADDR=:4000
VOLUME /data
EXPOSE 4000
USER nonroot:nonroot
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 CMD ["/todo", "healthcheck"]
ENTRYPOINT ["/todo"]
