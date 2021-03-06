FROM golang:latest
RUN mkdir /app
ADD . /app/
WORKDIR /app/cmd
RUN go build -o banners .
CMD ["./banners", "--config", "/app/configs/config.json"]