# K8sCD

Simple GitOps tool for Kubernetes

# Requirements

- Install `Git`
- Install `kubectl`

Install k8scd

```
mkdir ~/.k8scd
wget https://github.com/stevenzack/k8scd/releases/download/latest/k8scd_linux_amd64_latest -O ~/.k8scd/k8scd
chmod +x ~/.k8scd/k8scd
cd ~/.k8scd && ./k8scd &
```

And open K8sCD web UI in [http://localhost:9876](http://localhost:9876)
