FROM golang:1.24

WORKDIR /app

COPY ../go.mod ../go.sum ./

RUN go mod download

COPY .. .

RUN go build -o chat-app ./cmd

EXPOSE 8080

CMD [ "./chat-app" ]