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

kubectl apply --filename https://raw.githubusercontent.com/vfarcic/devops-toolkit-crossplane/master/crossplane-config/config-app.yaml

kubectl apply --filename https://raw.githubusercontent.com/vfarcic/devops-toolkit-crossplane/master/crossplane-config/config-gitops.yaml

kubectl apply --filename https://raw.githubusercontent.com/vfarcic/devops-toolkit-crossplane/master/crossplane-config/config-k8s.yaml

kubectl apply --filename https://raw.githubusercontent.com/vfarcic/devops-toolkit-crossplane/master/crossplane-config/config-monitoring.yaml

kubectl apply --filename https://raw.githubusercontent.com/vfarcic/devops-toolkit-crossplane/master/crossplane-config/config-sql.yaml

go run main.go interactive
```