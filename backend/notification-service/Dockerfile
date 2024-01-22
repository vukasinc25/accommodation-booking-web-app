FROM golang:alpine as build_container
WORKDIR /app
COPY ./notification-service/go.mod ./notification-service/go.sum ./
COPY /utility ../utility
RUN go mod download
COPY ./notification-service/ .
RUN go build -o server

FROM alpine
WORKDIR /usr/bin
COPY --from=build_container /app/server .
EXPOSE 8000
ENTRYPOINT ["server"]