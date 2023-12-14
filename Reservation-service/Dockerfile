FROM golang:alpine as build_container
WORKDIR /app
COPY ./Reservation-service/go.mod ./Reservation-service/go.sum ./
COPY /utility ../utlity
RUN go mod download
COPY ./Reservation-service/ .
RUN go build -o server

FROM alpine
WORKDIR /usr/bin
COPY --from=build_container /app/server .
EXPOSE 8000
ENTRYPOINT ["server"]