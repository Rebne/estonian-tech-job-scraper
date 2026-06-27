FROM golang:1.26-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/scrapy-project-v2 ./cmd/main.go

FROM golang:1.26-bookworm AS playwright

WORKDIR /tmp/playwright

COPY go.mod go.sum ./

ENV PLAYWRIGHT_BROWSERS_PATH=/ms-playwright

RUN go run github.com/playwright-community/playwright-go/cmd/playwright@v0.5700.1 install --with-deps chromium
RUN mkdir -p /playwright-driver && \
    cp -R /root/.cache/ms-playwright-go/*/node /playwright-driver/node && \
    cp -R /root/.cache/ms-playwright-go/*/package /playwright-driver/package

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends apt-transport-https ca-certificates curl gnupg chromium && \
    curl -sLf --retry 3 --tlsv1.2 --proto "=https" https://packages.doppler.com/public/cli/gpg.DE2A7741A397C129.key | gpg --dearmor -o /usr/share/keyrings/doppler-archive-keyring.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/doppler-archive-keyring.gpg] https://packages.doppler.com/public/cli/deb/debian any-version main" > /etc/apt/sources.list.d/doppler-cli.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends doppler && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/scrapy-project-v2 ./scrapy-project-v2
COPY --from=playwright /ms-playwright /ms-playwright
COPY --from=playwright /playwright-driver /playwright-driver

ENV CHROME_EXECUTABLE_PATH=/usr/bin/chromium \
    PLAYWRIGHT_BROWSERS_PATH=/ms-playwright \
    PLAYWRIGHT_DRIVER_PATH=/playwright-driver
