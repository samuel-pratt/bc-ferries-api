FROM golang:1.26.7-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /server ./cmd/server

FROM build AS test
RUN apk add --no-cache gcc musl-dev chromium && adduser -D -u 10001 app && chown -R app /src /go
ENV CHROME_NO_SANDBOX=true CHROME_SMOKE=true
USER app
CMD ["go", "test", "-race", "-count=1", "./..."]

FROM alpine:3.23
RUN apk add --no-cache ca-certificates tzdata chromium wget && adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /server /app/server
COPY static ./static
ENV CHROME_NO_SANDBOX=true CHROME_PATH=/usr/bin/chromium CHROME_BIN=/usr/bin/chromium
USER app
EXPOSE 8081
ENTRYPOINT ["/app/server"]
