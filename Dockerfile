FROM golang:1.20-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o vidi-server ./main.go

EXPOSE 8086

CMD ["/app/vidi-server"]
