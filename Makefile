run:
	go run .

docker:
	docker build -t main .
	docker image prune -f
	docker run -p 9876:9876 -d -e PASSWORD=12345671 --name k8scd main