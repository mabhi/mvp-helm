This project is a POC that enables the end user to install/update and delete applications in a cluster. This is similar to behavior like Helm cli, but instead of using helm command this offers
developer ease of use custom resource to enter requisite information of the repository they like to deploy. So these are for k8s developers who like to go abou their day only with k8s native commands.

## Prerequisites
1. This project is in development mode run in a local kubernetes cluster. For this purpose, I have used `kind` application to generate local cluster.
2. This project also relies extensively on `helmclient` kind to actually do the heavy lifting of installing the packages and dependencies. 
Ref: [github.com/mittwald/go-helm-client](github.com/mittwald/go-helm-client)


## Theory
In order to install package following resource needs to be created in cluster. Use this as a sample:
``` 
--- install-mysql.yaml --- 
apiVersion: mabhi.dev/v1
kind: HelmAction
metadata:
  name: helm-install
actionName: my-mysql
spec:
  chartVersion: "9.9.1"
  chartName: "bitnami/mysql"
  repoUrl: "https://charts.bitnami.com/bitnami"
  repoName: "bitnami"
```

- kubectl apply -f install-mysql.yaml
- kubectl delete hac helm-install

## Steps

- Deploy the custom CRD to inform the cluster about the incoming CRs.
- Create a CR to take user input, the controller will detect it and it will spin up helm deployment.
- Delete the CR to uninstall the release.
