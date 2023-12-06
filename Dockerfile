FROM golang:1.21.5 AS builder

RUN go version

ARG PROJECT_VERSION

COPY . /app
WORKDIR /app

RUN go mod tidy

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o main .

RUN go test -cover -v ./...

FROM zigzigcheers/upx as minify
WORKDIR /root
COPY --from=builder /app/main .
RUN upx --best --lzma ./main

FROM scratch AS production
WORKDIR /
COPY --from=minify /root/main . 
EXPOSE 9876
ENTRYPOINT [ "./main"]