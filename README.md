# Knative Eventing Sample Source

[![GoDoc](https://godoc.org/knative.dev/sample-source?status.svg)](https://godoc.org/knative.dev/sample-source)
[![Go Report Card](https://goreportcard.com/badge/knative/sample-source)](https://goreportcard.com/report/knative/sample-source)

Knative Eventing `sample-source` defines a simple source that transforms events
from an HTTP server into CloudEvents and demonstrates the canonical style in
which Knative Eventing writes sources.

To learn more about Knative, please visit our
[Knative docs](https://github.com/knative/docs) repository.

If you are interested in contributing, see [CONTRIBUTING.md](./CONTRIBUTING.md)
and [DEVELOPMENT.md](./DEVELOPMENT.md).


## Dependencies
1. K8s v1.15+ environment
2. ko
3. rocketmq-client-go v2.1.0
4. knative v0.17.0

## Run locally



Setup ko to use the minikube docker instance and local registry
```
eval $(minikube docker-env)
export KO_DOCKER_REPO=ko.local
```

Apply the CRD and configuration yaml
```
ko apply -f config
```
Once the sample-source-controller-manager is running in the knative-samples namespace, you can apply the example.yaml to connect our sample-source every 10s directly to a ksvc.
```
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: event-display
  namespace: knative-samples
spec:
  template:
    spec:
      containers:
        - image: gcr.io/knative-releases/knative.dev/eventing-contrib/cmd/event_display
---
apiVersion: samples.knative.dev/v1alpha1
kind: SampleSource
metadata:
  name: sample-source
  namespace: knative-samples
spec:
  interval: "10s"
  sink:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: event-display
```

```
ko apply -f example.yaml
```

