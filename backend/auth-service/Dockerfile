FROM golang:alpine as build_container
WORKDIR /app
COPY ./auth-service/go.mod ./auth-service/go.sum ./
RUN go mod download
COPY ./auth-service/ .
RUN go build -o server

FROM alpine
WORKDIR /usr/bin
COPY --from=build_container /app/server .
RUN mkdir cert
COPY ./auth-service/cert/auth-server.crt /cert/auth-server.crt
COPY ./auth-service/cert/auth-server.key /cert/auth-server.key
EXPOSE 8000
ENTRYPOINT ["server"]