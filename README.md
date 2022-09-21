# Ingress proxy for Homin.dev

Live > [HERE](https://homin.dev) <


## CICD

> Currently CICD is manual :(

Integration:

```bash 
export IMAGE_TAG=gcr.io/homin-dev/ingress_proxy:latest 
docker buildx build --platform linux/amd64 -t $IMAGE_TAG .
docker push $IMAGE_TAG
```

Deployment:

> k8s configs are move to [suapapa/k8s-homin.dev](https://github.com/suapapa/k8s-homin.dev)

```bash
k apply -f cm/ingress-links.yaml deploy/deploy-ingress_proxy.yaml
```


## SSL Cert

### Create

Cert create is done by automatically run follwing script when pod is started:

- `create_ssl_cert.sh`

### Renew

Cert will be renewd by cron. See `Dockerfile` for the detail