# Sample Go NATS Streaming program with KEDA.
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fbalchua%2Fgonuts.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fbalchua%2Fgonuts?ref=badge_shield)


This repo contains sample Go code to publish and subscribe message from [NATS Streaming](https://github.com/nats-io/nats-streaming-server) using [Keda](https://github.com/kedacore/keda) to autoscale the consumers.

The publisher will continuously publish messages to NATS Streaming Server.

The project is divided into several folders.

* `natss-chart` - Contains the helm chart to install Nats Streaming server.
* `pub` - contains the publisher code.
* `sub` - contains the subscriber code.
* [`k8s-manifest/pub`](./k8s-manifest/pub/values.yaml) - contains the Helm chart to deploy the publisher.
* [`k8s-manifest/sub`](./k8s-manifest/sub/values.yaml) - contains the Helm chart to deploy the subscriber.
* [`keda-nats-scaler`](./keda-nats-scaler/stan_scaledobject.yaml) - contains the `ScaledObject` to feed Keda for autoscaling.
  
## Pre-requisites

* [skaffold](https://skaffold.dev/docs/) - Use to build and deploy the application to Kubernetes.
* [helm](https://helm.sh/) - Kubernetes package manager.
* Kubernetes cluster - such as [Microk8s](https://microk8s.io/), with API Aggregation Layer enabled, version preferrably >= 1.15.
* kubectl
* [keda](https://github.com/kedacore/keda)


## Getting Started

### 1. Install nats streaming statefulset.

Go to directory `natss-chart`

`helm install --namespace stan -n stan . `

### 2. Building the sources

Since we are using skaffold to quickly build and deploy the application, we are going to use the `envTemplate` `PUB_TAG` tag.  For more information on using skaffold Taggers [here](https://skaffold.dev/docs/how-tos/taggers/).


**We use kaniko with skaffold to build our container image in-cluster**

Setup kaniko registry access secret

`kubectl -n gonuts create secret generic regcred --from-file $HOME/.docker/config.json`


Modify the docker registry repository in the `skaffold.yaml` file.  Example below:


```yaml
apiVersion: skaffold/v1beta15
kind: Config
profiles:
  - name: pub
    build:
      artifacts:
      - image: <your-repo>/gonuts-pub
        context: pub
        kaniko:
          dockerfile: Dockerfile
          buildContext:
            localDir: {}
          cache:
            repo: <your-repo>/gonuts-pub #for the layers in the Dockerfile
      cluster:
        dockerConfig: 
          secretName: regcred
        namespace: gonuts
    deploy:
      helm:
        releases:
          - name: gonuts-pub
            chartPath: k8s-manifest/pub
            namespace: gonuts
            wait: true
            values: 
              image.repository: <your-repo>/gonuts-pub

  - name: sub
    build:
      artifacts:
      - image: <your-repo>/gonuts-sub
        context: sub
        kaniko:
          dockerfile: Dockerfile
          buildContext:
            localDir: {}
          cache:
            repo: <your-repo>/gonuts-sub          
      cluster:
        dockerConfig: 
          secretName: regcred
        namespace: gonuts
    deploy:
      helm:
        releases:
          - name: gonuts-sub
            chartPath: k8s-manifest/sub
            namespace: gonuts
            values: 
              image.repository: <your-repo>/gonuts-sub

```

#### Build and run the publisher

Go to the project's root directory and execute the command below.

`skaffold run -p pub`

You should see some logs which looks like this.

```shell

$ skaffold run -p pub
Generating tags...
 - 192.168.1.12:32000/gonuts-sub -> 192.168.1.12:32000/gonuts-sub:5ab46ba
Checking cache...
 - 192.168.1.12:32000/gonuts-sub: Not found. Building
Creating docker config secret [regcred]...
Building [192.168.1.12:32000/gonuts-sub]...
Storing build context at /tmp/context-f42376857262386c48fc128506c21176.tar.gz
INFO[0000] Resolved base name golang:1.12.10 to golang:1.12.10 
INFO[0000] Resolved base name gcr.io/distroless/base to gcr.io/distroless/base 
INFO[0000] Resolved base name golang:1.12.10 to golang:1.12.10 
INFO[0000] Resolved base name gcr.io/distroless/base to gcr.io/distroless/base 
INFO[0000] Downloading base image golang:1.12.10        
INFO[0003] Error while retrieving image from cache: getting file info: stat /cache/sha256:e699a540de350a0993ce3a3f8238161e5c26bbf728bffe1dc1f75952a987ea30: no such file or directory 
INFO[0003] Downloading base image golang:1.12.10        

```

**Towards the end, you should see the logs and pub application deployed.**

```shell
Starting deploy...
Helm release gonuts-pub not installed. Installing...
No requirements found in k8s-manifest/pub/charts.
WARN[0113] image [balchu/gonuts-pub:0.0.1@sha256:1b9517a25cd8084f608a30b8307d92f07c88c52f6d9dd41832b576f7ff216b0d] is not used. 
WARN[0113] image [balchu/gonuts-pub] is used instead.   
WARN[0113] See helm sample for how to replace image names with their actual tags: https://github.com/GoogleContainerTools/skaffold/blob/master/examples/helm-deployment/skaffold.yaml 
NAME:   gonuts-pub
LAST DEPLOYED: Tue Oct 15 07:52:58 2019
NAMESPACE: gonuts
STATUS: DEPLOYED

RESOURCES:
==> v1/Deployment
NAME        READY  UP-TO-DATE  AVAILABLE  AGE
gonuts-pub  1/1    1           1          2s

==> v1/Pod(related)
NAME                         READY  STATUS   RESTARTS  AGE
gonuts-pub-75fbd86dd8-5z6b8  1/1    Running  0         2s


NOTES:
1. Get the application URL by running these commands:


Deploy complete in 3.156392449s


$ kubectl -n gonuts get pods
NAME                          READY   STATUS    RESTARTS   AGE
gonuts-pub-75fbd86dd8-5z6b8   1/1     Running   0          51s

```

#### Build and run the subscriber

To install the publisher application, the process is the same, simply run the command below.

`skaffold build -p sub`

You should see something in the console which looks like this.

```shell
$ skaffold run -p sub
Generating tags...
 - balchu/gonuts-sub -> balchu/gonuts-sub:0.0.1
Tags generated in 141.748µs
Checking cache...
 - balchu/gonuts-sub: Not found. Building
Cache check complete in 7.332088677s
Starting build...
Creating docker config secret [regcred]...
Building [balchu/gonuts-sub]...
Storing build context at /tmp/context-bef52d86051b333caeb3cdedb87edd64.tar.gz
INFO[0003] Resolved base name golang:1.12.10 to golang:1.12.10 
INFO[0003] Resolved base name gcr.io/distroless/base to gcr.io/distroless/base 
INFO[0003] Resolved base name golang:1.12.10 to golang:1.12.10 
INFO[0003] Resolved base name gcr.io/distroless/base to gcr.io/distroless/base 
INFO[0003] Downloading base image golang:1.12.10        
INFO[0006] Error while retrieving image from cache: getting file info: stat /cache/sha256:e699a540de350a0993ce3a3f8238161e5c26bbf728bffe1dc1f75952a987ea30: no such file or directory 
INFO[0006] Downloading base image golang:1.12.10        
INFO[0009] Downloading base image gcr.io/distroless/base 
INFO[0010] Error while retrieving image from cache: getting file info: stat /cache/sha256:e37cf3289c1332c5123cbf419a1657c8dad0811f2f8572433b668e13747718f8: no such file or directory 
INFO[0010] Downloading base image gcr.io/distroless/base 
INFO[0011] Built cross stage deps: map[0:[/app]]        
INFO[0011] Downloading base image golang:1.12.10        
INFO[0014] Error while retrieving image from cache: getting file info: stat /cache/sha256:e699a540de350a0993ce3a3f8238161e5c26bbf728bffe1dc1f75952a987ea30: no such file or directory 
INFO[0014] Downloading base image golang:1.12.10        
INFO[0016] Using files from context: [/kaniko/buildcontext/go.mod] 
INFO[0016] Using files from context: [/kaniko/buildcontext/go.sum] 
INFO[0016] Checking for cached layer balchu/gonuts-sub:70bc38fd8a7ca199781341718b0189eeba1bedcb38a2769fa3de737e8f29c2f2... 
INFO[0019] Using caching version of cmd: RUN go mod download 
INFO[0019] Using files from context: [/kaniko/buildcontext/main.go] 
INFO[0019] Checking for cached layer balchu/gonuts-sub:9c3409edd1ca5b909af4d63af1f1e93e9c6c71d1a225f913a54f8d4140bbf7d6... 
INFO[0021] No cached layer found for cmd RUN go build -o /app -v . 
INFO[0021] Unpacking rootfs as cmd RUN go build -o /app -v . requires it. 

```

Afterwards the subscriber application is successfully deployed.

```shell
Build complete in 1m28.406449903s
Starting test...
Test complete in 20.279µs
Tags used in deployment:
 - balchu/gonuts-sub -> balchu/gonuts-sub:0.0.1@sha256:c72b358f759278859e00c9e80d9bc41fe6ffa33f6bd724186f963497bbf4d561
Starting deploy...
Helm release gonuts-sub not installed. Installing...
No requirements found in k8s-manifest/sub/charts.
WARN[0096] image [balchu/gonuts-sub:0.0.1@sha256:c72b358f759278859e00c9e80d9bc41fe6ffa33f6bd724186f963497bbf4d561] is not used. 
WARN[0096] image [balchu/gonuts-sub] is used instead.   
WARN[0096] See helm sample for how to replace image names with their actual tags: https://github.com/GoogleContainerTools/skaffold/blob/master/examples/helm-deployment/skaffold.yaml 
NAME:   gonuts-sub
LAST DEPLOYED: Tue Oct 15 08:09:54 2019
NAMESPACE: gonuts
STATUS: DEPLOYED

RESOURCES:
==> v1/Deployment
NAME        READY  UP-TO-DATE  AVAILABLE  AGE
gonuts-sub  0/1    0           0          0s

==> v1/Pod(related)
NAME                         READY  STATUS   RESTARTS  AGE
gonuts-sub-55495c99cb-hr28q  0/1    Pending  0         0s


NOTES:
1. Get the application URL by running these commands:


Deploy complete in 965.091545ms

$ kubectl -n gonuts get pods
NAME                          READY   STATUS    RESTARTS   AGE
gonuts-pub-75fbd86dd8-5z6b8   1/1     Running   0          17m
gonuts-sub-55495c99cb-hr28q   1/1     Running   0          30s
```

#### Verify consumer is getting messages.

```shell

$ kubectl -n gonuts logs -f gonuts-sub-55495c99cb-hr28q 
Client ID is 1571098195710672840
Connected to nats://stan-nats-ss.stan.svc.cluster.local:4222 clusterID: [local-stan] clientID: [1571098195710672840]
Listening on [Test], clientID=[1571098195710672840], qgroup=[grp1] durable=[ImDurable]
[#1] Received: sequence:1 subject:"Test" data:"Message is : 2019-10-14 23:53:00.253880875 +0000 UTC m=+0.011842249" timestamp:1571097180254174208 
[#2] Received: sequence:2 subject:"Test" data:"Message is : 2019-10-14 23:53:00.769166147 +0000 UTC m=+0.527127679" timestamp:1571097180769799144 
[#3] Received: sequence:3 subject:"Test" data:"Message is : 2019-10-14 23:53:01.278711647 +0000 UTC m=+1.036673208" timestamp:1571097181279410272 
[#4] Received: sequence:4 subject:"Test" data:"Message is : 2019-10-14 23:53:01.788774307 +0000 UTC m=+1.546735690" timestamp:1571097181789001470 
[#5] Received: sequence:5 subject:"Test" data:"Message is : 2019-10-14 23:53:02.295029913 +0000 UTC m=+2.052991385" timestamp:1571097182295624638 
[#6] Received: sequence:6 subject:"Test" data:"Message is : 2019-10-14 23:53:02.804475168 +0000 UTC m=+2.562436662" timestamp:1571097182805155109 
[#7] Received: sequence:7 subject:"Test" data:"Message is : 2019-10-14 23:53:03.313488105 +0000 UTC m=+3.071449602" timestamp:1571097183314223852 
[#8] Received: sequence:8 subject:"Test" data:"Message is : 2019-10-14 23:53:03.824778499 +0000 UTC m=+3.582740077" timestamp:1571097183825354528 
[#9] Received: sequence:9 subject:"Test" data:"Message is : 2019-10-14 23:53:04.333812324 +0000 UTC m=+4.091773798" timestamp:1571097184334662435 
[#10] Received: sequence:10 subject:"Test" data:"Message is : 2019-10-14 23:53:04.843043309 +0000 UTC m=+4.601004807" timestamp:1571097184843811709 
[#11] Received: sequence:11 subject:"Test" data:"Message is : 2019-10-14 23:53:05.352658775 +0000 UTC m=+5.110620353" timestamp:1571097185353256363 
[#12] Received: sequence:12 subject:"Test" data:"Message is : 2019-10-14 23:53:05.861342784 +0000 UTC m=+5.619304279" timestamp:1571097185861954607 
[#13] Received: sequence:13 subject:"Test" data:"Message is : 2019-10-14 23:53:06.371526442 +0000 UTC m=+6.129487956" timestamp:1571097186372047182 
[#14] Received: sequence:14 subject:"Test" data:"Message is : 2019-10-14 23:53:06.878906799 +0000 UTC m=+6.636868299" timestamp:1571097186879544355 
[#15] Received: sequence:15 subject:"Test" data:"Message is : 2019-10-14 23:53:07.385460833 +0000 UTC m=+7.143422383" timestamp:1571097187386228503 
[#16] Received: sequence:16 subject:"Test" data:"Message is : 2019-10-14 23:53:07.892033689 +0000 UTC m=+7.649995236" timestamp:1571097187892522959 
[#17] Received: sequence:17 subject:"Test" data:"Message is : 2019-10-14 23:53:08.397821951 +0000 UTC m=+8.155783439" timestamp:1571097188398585249 
[#18] Received: sequence:18 subject:"Test" data:"Message is : 2019-10-14 23:53:08.904641497 +0000 UTC m=+8.662602992" timestamp:1571097188905298883 
[#19] Received: sequence:19 subject:"Test" data:"Message is : 2019-10-14 23:53:09.411199672 +0000 UTC m=+9.169161147" timestamp:1571097189411784820 
[#20] Received: sequence:20 subject:"Test" data:"Message is : 2019-10-14 23:53:09.91938099 +0000 UTC m=+9.677342489" timestamp:1571097189920038609 
[#21] Received: sequence:21 subject:"Test" data:"Message is : 2019-10-14 23:53:10.426261076 +0000 UTC m=+10.184222622" timestamp:1571097190426920141 
[#22] Received: sequence:22 subject:"Test" data:"Message is : 2019-10-14 23:53:10.932787271 +0000 UTC m=+10.690748800" timestamp:1571097190933297474 
[#23] Received: sequence:23 subject:"Test" data:"Message is : 2019-10-14 23:53:11.438715414 +0000 UTC m=+11.196676999" timestamp:1571097191439532219 
[#24] Received: sequence:24 subject:"Test" data:"Message is : 2019-10-14 23:53:11.945399373 +0000 UTC m=+11.703360969" timestamp:1571097191945971410 
[#25] Received: sequence:25 subject:"Test" data:"Message is : 2019-10-14 23:53:12.458382509 +0000 UTC m=+12.216343995" timestamp:1571097192458873823 
[#26] Received: sequence:26 subject:"Test" data:"Message is : 2019-10-14 23:53:12.968315244 +0000 UTC m=+12.726276708" timestamp:1571097192968837373 
```
#### Install Keda

Follow the instructions from [Keda](https://github.com/kedacore/keda#setup) site.


Verify if keda pod is running.

```shell
$ kubectl -n keda get pods
NAME                    READY   STATUS    RESTARTS   AGE
keda-67df4596b6-4zkgr   1/1     Running   0          6s
```

#### Apply the [`keda-nats-scaler/stan_scaledobject.yaml`](keda-nats-scaler/stan_scaledobject.yaml)

After applying the scaler, you should see the pods scale up.

```shell
$ kubectl -n gonuts get pods
NAME                          READY   STATUS              RESTARTS   AGE
gonuts-pub-75fbd86dd8-5z6b8   1/1     Running             0          30m
gonuts-sub-55495c99cb-76nwr   0/1     ContainerCreating   0          1s
gonuts-sub-55495c99cb-h6rtz   1/1     Running             0          17s
gonuts-sub-55495c99cb-hm76r   0/1     ContainerCreating   0          1s
gonuts-sub-55495c99cb-hr28q   1/1     Running             0          13m
gonuts-sub-55495c99cb-qk8bq   1/1     Running             0          17s
gonuts-sub-55495c99cb-sdblq   0/1     ContainerCreating   0          1s
gonuts-sub-55495c99cb-stbvq   1/1     Running             0          17s
gonuts-sub-55495c99cb-wnm4n   0/1     ContainerCreating   0          1s

```

## Changing the production rate of messages

You can modify the production rate of message by changing the value of `delayInBetweenPublish` in the file `k8s-manifest/pub/value.yaml`

For example,production rate of 1 message every 2 seconds:

```yaml

# Default values for pub.
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.

replicaCount: 1

image:
  repository: localhost:32000/gonuts-pub:SNAPSHOT
  pullPolicy: Always

imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""

delayInBetweenPublish: 2000
subject: "Test"
natsStreamingServerEndpoint: "nats://stan-nats-ss.stan.svc.cluster.local:4222"

```

## Testing scale to zero

To test that Keda will successfully scale the subscriber pods to zero, simply delete the publisher application.

```shell
$ skaffold delete  -p pub
Cleaning up...
release "gonuts-pub" deleted
Cleanup complete in 634.066199ms
```

Wait for a few minutes, depending on the number of pending messages you have in the broker and the publisher pods will be scaled to zero.

```shell
$ kubectl -n gonuts get pods

NAME                          READY   STATUS        RESTARTS   AGE
gonuts-sub-55495c99cb-26kb2   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-4ljw4   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-76nwr   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-bvzkt   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-bwr4n   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-dkdcd   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-fpc7t   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-h6rtz   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-hm76r   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-hr28q   1/1     Terminating   0          26m
gonuts-sub-55495c99cb-jj7fx   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-jmdfv   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-jn9h9   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-klwq7   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-kmttg   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-l67sv   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-lfhck   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-llvlr   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-qf4w7   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-qk8bq   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-rq9sh   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-sdblq   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-stbvq   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-wnm4n   1/1     Terminating   0          12m
gonuts-sub-55495c99cb-xk97t   1/1     Terminating   0          12m

$ kubectl -n gonuts get pods

No resources found in gonuts namespace.
```

### Changing the ScaledObject

You can change the values in file [`keda-nats-scaler/stan_scaledobject.yaml`](keda-nats-scaler/stan_scaledobject.yaml)

Example:

```yaml
apiVersion: keda.k8s.io/v1alpha1
kind: ScaledObject
metadata:
  name: stan-scaledobject
  namespace: gonuts
  labels:
    deploymentName: gonuts-sub
spec:
  pollingInterval: 10   # Optional. Default: 30 seconds
  cooldownPeriod: 30   # Optional. Default: 300 seconds
  minReplicaCount: 0   # Optional. Default: 0
  maxReplicaCount: 30  # Optional. Default: 100  
  scaleTargetRef:
    deploymentName: gonuts-sub
  triggers:
  - type: stan
    metadata:
      natsServerMonitoringEndpoint: "stan-nats-ss.stan.svc.cluster.local:8222"
      queueGroup: "grp1"
      durableName: "ImDurable"
      subject: "Test"
      lagThreshold: "10"
```

Where:

* `natsServerMonitoringEndpoint` : Is the location of the Nats Streaming monitoring endpoint.  In this example it is the FQDN of nats streaming deployed.
* `queuGroup` : The queue group name of the subscribers.
* `durableName` :  Must identify the durability name used by the subscribers.
* `subject` : Sometimes called the channel name.
* `lagThreshold` : This value is used to tell the Horizontal Pod Autoscaler to use as TargetAverageValue.


### Clean up

```
$ skaffold delete -p sub
$ skaffold delete -p pub
$ helm del --purge stan
```

## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fbalchua%2Fgonuts.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fbalchua%2Fgonuts?ref=badge_large)