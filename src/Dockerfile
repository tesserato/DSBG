FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY README.md sample_content/01_readme.md

RUN go build .

EXPOSE 666