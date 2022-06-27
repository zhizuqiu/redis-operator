# redis-operator

redis sentinel 模式的 kubernetes operator

## Features

- redis sentinel
- 自动形成主备哨兵模式
- 在线修改密码、配置，无需重启节点
- 确保 slave 复制同一个 master
- 确保 sentinel 监控同一个 master
- 实时同步 pod 的状态
- 支持 host / vpc 网络模式

```
apiVersion: component.zhizuqiu/v1alpha1
kind: Redis
metadata:
  name: redis-sample
spec:
  sentinel:
    image: 'redis:5.0-alpine'
    replicas: 3
    resources:
      requests:
        cpu: 100m
      limits:
        memory: 100Mi
  redis:
    image: 'redis:5.0-alpine'
    replicas: 3
    resources:
      requests:
        cpu: 100m
        memory: 100Mi
      limits:
        cpu: 400m
        memory: 500Mi
  auth:
    secretPath: redis-auth
```

更多[例子](samples/cr) 

## Installation

```
kubectl create ns redis-system
kubectl apply -f samples/crd/*
```

## ports used

- 38111 --metrics-addr
- 38002 --secure-listen-address
- - ```curl -H "Authorization: Bearer TOKEN" https://HOST:38002/metrics  --insecure```

## 相关命令:

```
operator-sdk init --domain=zhizuqiu --repo=github.com/zhizuqiu/redis-operator
operator-sdk create api --group component --version v1alpha1 --kind Redis --resource=true --controller=true
go test -v ./... -short
make docker-build docker-push IMG=docker.io/zhizuqiu/redis-operator:latest
make generate
```

## resources name rules

example:
- INSTANCE_NAME: redis-sample

name rules:
- redis StatefulSets name: `redis-redis-sample-0`,`{redis}-{INSTANCE_NAME}-{index}`
- redis Pod name: `redis-redis-sample-0-0`,`{redis}-{INSTANCE_NAME}-{index}-{0}`,`{STATEFULSETS_NAME}-{0}`
- redis ConfigMap name: `redis-redis-sample`,`{redis}-{INSTANCE_NAME}`
- redis Readiness ConfigMap name: `redis-readiness-redis-sample`,`{redis-readiness}-{INSTANCE_NAME}`
- redis ConfigMap name: `redis-redis-sample`,`{redis}-{INSTANCE_NAME}`

- sentinel StatefulSets name: `sentinel-redis-sample-0`,`{sentinel}-{INSTANCE_NAME}-{index}`
- sentinel Pod name: `sentinel-redis-sample-0`,`{sentinel}-{INSTANCE_NAME}-{index}-{0}`,`{STATEFULSETS_NAME}-{0}`
- sentinel ConfigMap name: `sentinel-redis-sample`,`{sentinel}-{INSTANCE_NAME}`
- sentinel Service name: `sentinel-redis-sample`,`{sentinel}-{INSTANCE_NAME}`

- exporter Deployment name: `exporter-redis-sample`,`{exporter}-{INSTANCE_NAME}`

## path rules

- redis Config Writable Path: `/data/conf/redis.conf`
- redis Config Path: `/redis/redis.conf`

- redis Sentinel Config Writable Path: `/data/conf/sentinel.conf`
- redis Sentinel Config Path: `/redis/sentinel.conf`

```
# 列举某个ns下的所有Pv
kubectl get pv -l app.kubernetes.io/namespace={ns}

# 列举某个实例的所有Pv
kubectl get pv -l app.kubernetes.io/name={instacen_name}
```

## 生成 yaml:

```
# samples/crd/redis-crd.yaml
make create-crd

# samples/crd/redis-rbac.yaml
make create-rbac

# samples/crd/redis-deploy.yaml
make create-deploy

# or
make create-deploy \
IMG=docker.io/zhizuqiu/redis-operator:latest
```
