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

go run main.go

# Open `http://localhost:8080` in a browser
```

## TODO

* Create container image releases
* Create a Helm chart
* Prepopulate fields with default values
* Make required fields
* Support claims
* Edit fields
* Contain it within a window
* Fix the option to choose claims
* Export to YAML
* kubectl apply
* Push to Git
* See all running claims, compositions, and managed resources
* Add to krew
* Sleep