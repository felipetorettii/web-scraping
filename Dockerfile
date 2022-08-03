FROM golang:alpine

WORKDIR /app

RUN apk add git
RUN git config --global url."https://Junkes887:fb27714d199d4833caca9bac3f554b58c5d48bb8@github.com".insteadOf "https://github.com"

COPY go.mod go.sum ./
RUN  go mod download

COPY . .

RUN go build -o main

CMD [ "./main" ]