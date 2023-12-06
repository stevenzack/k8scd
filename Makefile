run:
	go run .

docker:
	- docker stop k8scd
	- docker rm k8scd
	- docker image prune -f
	docker build -t main .
	docker image prune -f
	docker run -p 9876:9876 -d -e KV=/var/local/kv -v /var/local/kv:/var/local/kv --name k8scd main