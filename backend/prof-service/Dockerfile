FROM golang:alpine as build_container
WORKDIR /app
COPY ./prof-service/go.mod ./prof-service/go.sum ./
COPY /utility ../utility
RUN go mod download
COPY ./prof-service/ .
RUN go build -o server

FROM alpine
WORKDIR /usr/bin
COPY --from=build_container /app/server .
RUN mkdir cert
COPY ./prof-service/cert/prof-server.crt /cert/prof-server.crt
COPY ./prof-service/cert/prof-server.key /cert/prof-server.key
# COPY ./prof-service/cert/file.log /cert/file.log
EXPOSE 8000
ENTRYPOINT ["server"]