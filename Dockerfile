FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o vidi-server .

# The container binds all interfaces so a published port works. Set
# VIDI_API_TOKEN at runtime; startup refuses an unprotected non-loopback bind.
ENV VIDI_HOST=0.0.0.0 \
    VIDI_PORT=8084
EXPOSE 8084

CMD ["/app/vidi-server"]
