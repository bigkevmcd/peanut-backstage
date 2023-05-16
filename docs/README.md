# Annotations and Labels

## Example

There is an example Deployment in [example/deployment.yaml](../example/deployment.yaml).

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app.kubernetes.io/component: web-server
    app.kubernetes.io/created-by: test-team
    app.kubernetes.io/instance: nginx-staging
    app.kubernetes.io/name: nginx
    app.kubernetes.io/part-of: user-system
    backstage.gitops.pro/lifecycle: staging
    backstage.io/kubernetes-id: user-system
  annotations:
    backstage.gitops.pro/description: This is a test
    backstage.gitops.pro/link-0: https://example.com/user,Example Users,user
    backstage.gitops.pro/link-1: https://example.com/group,Example Groups,group
    backstage.gitops.pro/tags: nginx,data
    testing.com/annotation: test-annotation
spec:
  replicas: 3
  selector:
    matchLabels:
      app.kubernetes.io/name: nginx
  template:
    metadata:
      labels:
        app.kubernetes.io/name: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80
```

# Labels

This is based on the [Kubernetes recommended labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/).

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app.kubernetes.io/component: web-server
    app.kubernetes.io/created-by: test-team
    app.kubernetes.io/instance: nginx-staging
    app.kubernetes.io/name: nginx
    app.kubernetes.io/part-of: user-system
    backstage.gitops.pro/lifecycle: staging
    backstage.io/kubernetes-id: user-system
```

This maps to a Backstage component:

```yaml
apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: nginx
  description: This is a test
```

