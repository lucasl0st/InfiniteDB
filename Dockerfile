FROM golang:1.20-alpine as builder

WORKDIR /src

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /infinitedb-server

FROM scratch

LABEL org.opencontainers.image.source="https://github.com/lucasl0st/InfiniteDB"
LABEL org.opencontainers.image.description="A Scalable Database"
LABEL org.opencontainers.image.licenses=GPL-3.0

ENV GIN_MODE=release
ENV PORT=8080

COPY --from=builder /infinitedb-server /usr/bin/infinitedb-server
COPY --from=tarampampam/curl:7.88.1 /bin/curl /bin/curl

EXPOSE $PORT

HEALTHCHECK --interval=5s --start-period=60s CMD [ "curl", "--fail", "http://localhost:8080/health"]

CMD [ "/usr/bin/infinitedb-server" ]