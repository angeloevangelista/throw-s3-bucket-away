FROM --platform=linux/amd64 golang:1.22.4

WORKDIR /app

COPY . .

RUN \
  go mod download && \
  go build -o build main.go

  ENTRYPOINT [ "/app/build" ]
