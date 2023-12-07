run:
	go run .
b:
	go build -ldflags="-s -w" -trimpath .
	upx --best --lzma k8scd