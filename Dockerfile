FROM golang:1.22

COPY . /build
WORKDIR /build

RUN go build -o /opt/app /build/cmd/app/main.go

ENTRYPOINT ["/opt/app", "-config", "/configs/config.yaml"]
