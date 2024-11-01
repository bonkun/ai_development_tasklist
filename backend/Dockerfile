FROM golang:1.21

WORKDIR /app

RUN go install github.com/cosmtrek/air@v1.49.0

COPY go.mod ./
RUN go mod download
RUN go mod tidy
COPY . .

WORKDIR /app/cmd

CMD ["air", "-c", "../.air.toml"]