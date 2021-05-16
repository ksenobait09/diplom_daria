FROM golang:1.16.4

WORKDIR /build

COPY . .
RUN go mod download

RUN go build -o main cmd/app/main.go

EXPOSE 1323

CMD ["/build/main", "-root=/daria"]
