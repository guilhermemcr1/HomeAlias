FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod main.go ./
RUN go build -o /homealias-agent .

FROM alpine:3.20
COPY --from=build /homealias-agent /usr/local/bin/homealias-agent
ENTRYPOINT ["homealias-agent"]
