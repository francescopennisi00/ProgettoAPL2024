FROM golang:1.21

WORKDIR /app

COPY . .

RUN go mod download

WORKDIR /app/wms-src

RUN go build -o wms_main

CMD ["./wms_main"]