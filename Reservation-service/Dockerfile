FROM golang:alpine as build_container
WORKDIR /app
COPY ./reservation-service/go.mod ./reservation-service/go.sum ./
COPY /utility ../utlity
RUN go mod download
COPY ./reservation-service/ .
RUN go build -o server

FROM alpine
WORKDIR /usr/bin
COPY --from=build_container /app/server .
EXPOSE 8000
ENTRYPOINT ["server"]