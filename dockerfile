FROM golang:1.23

WORKDIR app/

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o /binary

EXPOSE 10000

CMD ["/binary"]
