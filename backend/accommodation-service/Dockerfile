FROM golang:alpine as build_container
WORKDIR /app
COPY ./accommodation-service/go.mod ./accommodation-service/go.sum ./
COPY /utility ../utility
RUN go mod download
COPY ./accommodation-service/ .
RUN go build -o server

FROM alpine
WORKDIR /usr/bin
COPY --from=build_container /app/server .
EXPOSE 8000
ENTRYPOINT ["server"]