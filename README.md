## Demo

```bash
helm repo add crossplane-stable \
    https://charts.crossplane.io/stable

helm repo update

helm upgrade --install \
    crossplane crossplane-stable/crossplane \
    --namespace crossplane-system \
    --create-namespace \
    --wait

kubectl apply --filename 

kubectl apply --filename 

kubectl apply --filename 

kubectl apply --filename 

kubectl apply --filename 

go run main.go interactive
```