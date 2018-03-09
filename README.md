# open-service-broker-sacloud

[![Go Report Card](https://goreportcard.com/badge/github.com/sacloud/open-service-broker-sacloud)](https://goreportcard.com/report/github.com/sacloud/open-service-broker-sacloud)
[![Build Status](https://travis-ci.org/sacloud/open-service-broker-sacloud.svg?branch=master)](https://travis-ci.org/sacloud/open-service-broker-sacloud)

[`open-service-broker-sacloud`](https://github.com/sacloud/open-service-broker-sacloud) is a implementation of 
[`Open Service Broker API`](https://www.openservicebrokerapi.org) for [SAKURA Cloud](https://cloud.sakura.ad.jp).  

*CLOUD FOUNDRY and OPEN SERVICE BROKER are trademarks of the CloudFoundry.org Foundation in the United States and other countries.*  

## Supported Services

- [MariaDB](docs/services/mariadb.md)
- [PostgreSQL](docs/services/postgres.md)

## Installation and Usage

### Prerequisites

- A kubernetes cluster deployed in Sakura Cloud(with switched networks)
- A working Helm installation
- [Service Catalog](https://github.com/kubernetes-incubator/service-catalog)
- [Helm](https://github.com/kubernetes/helm)
- [Optional] [`svcat`: Service Catalog CLI](https://github.com/kubernetes-incubator/service-catalog/tree/master/cmd/svcat)

### Install

Use Helm to install Open Service Broker for SAKURA Cloud onto your Kubernetes cluster.
Refer to the Open Service Broker for SAKURA Cloud Helm chart for details on how to complete the installation.

> [Open Service Broker for SAKURA Cloud Helm chart](https://github.com/sacloud/helm-charts/tree/master/open-service-broker-sacloud)

### Provisioning

With the Kubernetes Service Catalog and Open Service Broker for SAKURA Cloud both installed on your Kubernetes cluster,
try creating a ServiceInstance resource to see service provisioning in action.

For example, the following will provision MariaDB on SAKURA Cloud:

```bash
# Put your SAKURA Cloud resource settings to service instance definition
$ vi examples/mariadb-service.yaml
```

```console
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: my-mariadb-instance
  namespace: default
spec:
  clusterServiceClassExternalName: sacloud-mariadb
  clusterServicePlanExternalName: db-10g
  parameters:
    switchID: <put-your-switch-id>
    ipaddress: "<put-your-database-ipaddress>"
    maskLen: <put-your-database-nw-mask-len>
    defaultRoute: "<put-your-database-def-route-ipaddress>"
```

```bash
# Provision MariaDB using service catalog
$ kubectl create -f examples/mariadb-service.yaml
```

After the ServiceInstance resource is submitted, you can view its status:

```bash
# using kubectl
$ kubectl get serviceinstance my-mariadb-instance -o yaml 

# using svcat(Service Catalog CLI)
$ svcat describe instance my-mariadb-instance
```

You'll see output that includes a status indicating that asynchronous provisioning is ongoing. Eventually,
that status will change to indicate that asynchronous provisioning is complete.

### Binding

Upon provision success, bind to the instance:

```bash
$ kubectl create -d examples/mariadb-binding.yaml
```

To check the status of the binding:

```bash
# using kubectl
$ kubectl get servicebinding my-mariadb-binding -o yaml

# using svcat(Service Catalog CLI)
$ svcat describe binding my-mariadb-binding
```

You'll see some output indicating that the binding was successful.
Once it is, a secret named my-mariadb-secret will be written that contains the database connection details in it.

You can observe that this secret exists and has been populated:

```bash
kubectl get secret my-mariadb-secret -o yaml
```

This secret can be used just as any other.

### Unbinding

To unbind:

```bash
$ kubectl delete -f examples/mariadb-binding.yaml
```

### Deprovisioning

To deprovision:

```bash
$ kubectl delete -f examples/mariadb-service.yaml
```

## Debug Service Broker 

To show Service Broker server log, do following:

```bash
# Confirm Service Broker server pod name
$ kubectl get pod --namespace=osbs

# Show logs
$ kubectl logs -f --namespace=osbs <service-broker-pod-name>
```

## License

 `open-service-broker-sacloud` Copyright (C) 2018 Kazumichi Yamamoto.

  This project is published under [Apache 2.0 License](LICENSE.txt).
  
## Author

  * Kazumichi Yamamoto ([@yamamoto-febc](https://github.com/yamamoto-febc))