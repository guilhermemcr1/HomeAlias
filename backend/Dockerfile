FROM golang:1.23-alpine AS build
WORKDIR /src
ENV GOTOOLCHAIN=local
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server && CGO_ENABLED=0 go build -o /migrate ./cmd/migrate

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /server /server
COPY --from=build /migrate /migrate
COPY migrations /migrations
RUN adduser -D -u 10001 app
USER app
EXPOSE 8080
ENTRYPOINT ["/server"]
