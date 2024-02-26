FROM golang:1.21

WORKDIR /app

COPY . .

RUN go mod download

WORKDIR /app/um-src

RUN go build -o um_main

CMD ["./um_main"]
