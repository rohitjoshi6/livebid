FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/livebid-api ./cmd/api

FROM alpine:3.20
RUN adduser -D -g '' livebid
USER livebid
COPY --from=build /bin/livebid-api /bin/livebid-api
EXPOSE 8080
ENTRYPOINT ["/bin/livebid-api"]

