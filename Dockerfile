# syntax=docker/dockerfile:1
#
# The todo, as one static binary in a distroless image.
#
#   docker build --build-context platform=../platform -t todo .
#
# go.mod replaces github.com/kitsunium/platform with ../platform until the
# platform is published; the named build context provides that directory.

FROM golang:1.27 AS build
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
