FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/urban-traffic-optimization .
FROM alpine:3.22
RUN adduser -D app && mkdir -p /data && chown -R app:app /data
USER app
COPY --from=build /out/urban-traffic-optimization /usr/local/bin/urban-traffic-optimization
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/urban-traffic-optimization"]
