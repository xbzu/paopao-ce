# syntax=docker/dockerfile:1.7

FROM node:22-alpine AS web-builder

WORKDIR /src/web
RUN corepack enable && corepack prepare yarn@1.22.22 --activate
COPY web/package.json web/yarn.lock ./
RUN --mount=type=cache,target=/usr/local/share/.cache/yarn \
    yarn install --frozen-lockfile
COPY web/ ./
RUN VITE_HOST= yarn build
RUN if grep -R -E -q '10\.10\.100\.113|paopao-assets\.oss-cn-shanghai\.aliyuncs\.com' dist; then \
      echo 'frontend build contains a private development API address or legacy runtime asset' >&2; \
      exit 1; \
    fi

FROM golang:1.24-alpine AS go-builder

ARG VERSION=container
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

ENV CGO_ENABLED=0 \
    GOPROXY=https://proxy.golang.org,direct

WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . ./
RUN test ! -e .env && test ! -e config.yaml && test ! -e .git
RUN rm -rf ./web/dist
COPY --from=web-builder /src/web/dist ./web/dist
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -tags 'go_json migration' \
      -ldflags "-s -w \
        -X github.com/rocboss/paopao-ce/pkg/version.version=${VERSION} \
        -X github.com/rocboss/paopao-ce/pkg/version.commitID=${COMMIT} \
        -X github.com/rocboss/paopao-ce/pkg/version.buildDate=${BUILD_DATE} \
        -X github.com/rocboss/paopao-ce/pkg/version.buildTags=go_json,migration" \
      -o /out/paopao .

FROM alpine:3.21 AS runtime

ENV TZ=Asia/Shanghai

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 paopao \
    && adduser -S -D -H -u 10001 -G paopao paopao

WORKDIR /app
COPY --from=go-builder --chown=paopao:paopao /out/paopao ./paopao
COPY --chown=paopao:paopao deploy/simple/config.yaml ./config.yaml
RUN mkdir -p /app/custom/data/sqlite3 /app/custom/data/oss \
    && chown -R paopao:paopao /app/custom

USER paopao
VOLUME ["/app/custom"]
EXPOSE 8008

HEALTHCHECK --interval=10s --timeout=3s --start-period=20s --retries=6 \
  CMD wget -q -O /dev/null http://127.0.0.1:8008/healthz || exit 1

ENTRYPOINT ["/app/paopao"]
CMD ["serve"]
