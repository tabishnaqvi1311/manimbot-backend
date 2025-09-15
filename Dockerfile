FROM golang:1.24.2-bullseye AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server

# -- #

FROM manimcommunity/manim:stable
WORKDIR /app
COPY --from=builder /app ./

EXPOSE 8000

CMD ["./server"]
