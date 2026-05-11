# Kubernetes Deployment

This directory deploys the FastAPI provider behind an Ingress for
`api.asset.quanttide.com`.

## Build the provider image

```bash
docker build -f manifests/docker/Dockerfile.provider \
  -t crpi-uorshhk4a32pmmio.cn-hangzhou.personal.cr.aliyuncs.com/quanttide/qtcloud-asset-provider:latest \
  .
docker push crpi-uorshhk4a32pmmio.cn-hangzhou.personal.cr.aliyuncs.com/quanttide/qtcloud-asset-provider:latest
```

Change the image repository in `provider.yaml` if the ACR namespace or
repository is different.

## Create the TLS secret

Use the certificate issued for `api.asset.quanttide.com`:

```bash
kubectl create namespace qtcloud-asset
kubectl -n qtcloud-asset create secret tls qtcloud-asset-api-tls \
  --cert=api.asset.quanttide.com.pem \
  --key=api.asset.quanttide.com.key
```

## Apply manifests

```bash
kubectl apply -f manifests/k8s/namespace.yaml
kubectl -n qtcloud-asset create secret docker-registry acr-pull-secret \
  --docker-server=crpi-uorshhk4a32pmmio.cn-hangzhou.personal.cr.aliyuncs.com \
  --docker-username="$ACR_USERNAME" \
  --docker-password="$ACR_PASSWORD"
kubectl apply -f manifests/k8s/provider.yaml
```

Point `api.asset.quanttide.com` to the Ingress controller public endpoint. For
an nginx ingress controller on Alibaba Cloud ACK, this is usually the external
IP or DNS name of the ingress controller `LoadBalancer` service.

## GitHub Actions deployment

The `Deploy Provider to Kubernetes` workflow applies these manifests after the
provider image workflow succeeds. Configure these repository secrets:

- `KUBE_CONFIG`: base64-encoded kubeconfig for the ACK cluster.
- `ACR_USERNAME`: ACR login username.
- `ACR_PASSWORD`: ACR login password.
- `API_TLS_CERT`: PEM certificate for `api.asset.quanttide.com`.
- `API_TLS_KEY`: PEM private key for `api.asset.quanttide.com`.

Create the kubeconfig secret with:

```bash
base64 -w 0 kubeconfig.yaml
```

## Verify

```bash
kubectl -n qtcloud-asset get pods,svc,ingress
curl -i https://api.asset.quanttide.com/health
```
