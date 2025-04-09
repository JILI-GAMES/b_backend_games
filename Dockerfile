FROM golang:alpine as builder

WORKDIR /app

COPY . .

RUN go mod download


# Build the application
RUN go build -o main .

EXPOSE 11400

CMD ["./main"]
