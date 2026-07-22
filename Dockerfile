#############################################
# CIMPLR Policy Check Service — standalone
# Port 8184, no Postgres, POST-only API
#############################################
FROM golang:1.24-bookworm AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o policy-service ./cmd/

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata curl
WORKDIR /app
COPY --from=build /app/policy-service /app/policy-service
EXPOSE 8184
ENV PORT=8184
CMD ["/app/policy-service"]
