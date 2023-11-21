FROM golang:alpine as build_container
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o server

FROM alpine
WORKDIR /usr/bin
COPY --from=build_container /app/server .
RUN mkdir cert
COPY ./cert ./cert/
EXPOSE 8000
ENTRYPOINT ["server"]