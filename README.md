# Sample Go NATS Streaming program with KEDA.
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fbalchua%2Fgonuts.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fbalchua%2Fgonuts?ref=badge_shield)


This repo contains sample Go code to publish and subscribe message from [NATS Streaming](https://github.com/nats-io/nats-streaming-server) using [Keda](https://github.com/kedacore/keda) to autoscale the consumers.

_*This project has been upgraded to use keda v2*_

The publisher will continuously publish messages to NATS Streaming Server.

The project is divided into several folders.

* `natss-chart` - Contains the helm chart to install Nats Streaming server.
* `pub` - contains the publisher code.
* `sub` - contains the subscriber code.
* [`k8s-manifest/pub`](./k8s-manifest/pub/values.yaml) - contains the Helm chart to deploy the publisher.
* [`k8s-manifest/sub`](./k8s-manifest/sub/values.yaml) - contains the Helm chart to deploy the subscriber.
* [`keda-nats-scaler`](./keda-nats-scaler/stan_scaledobject.yaml) - contains the `ScaledObject` to feed Keda for autoscaling.
  
## Pre-requisites

* [skaffold](https://skaffold.dev/docs/) - Use to build and deploy the application to Kubernetes.  Use version v1.16.0
* [helm](https://helm.sh/) - Use v3. Kubernetes package manager.
* Kubernetes cluster - such as [Microk8s](https://microk8s.io/), with API Aggregation Layer enabled, version preferrably >= 1.15.
* kubectl
* [keda](https://github.com/kedacore/keda) - Use v2


## Getting Started

### 1. Install nats streaming statefulset.

Make sure you are using helm v3.
From the root directory.

```
$ skaffold deploy -p stan
```

Verify that nats streaming is running

```
$ kubectl -n stan get pods 
NAME             READY   STATUS    RESTARTS   AGE
stan-nats-ss-0   1/1     Running   0          82s
```


### 2. Building the sources

Since we are using skaffold to quickly build and deploy the application, we are going to use the `envTemplate` `PUB_TAG` tag.  For more information on using skaffold Taggers [here](https://skaffold.dev/docs/how-tos/taggers/).


**We use kaniko with skaffold to build our container image in-cluster**

Setup kaniko registry access secret

`kubectl -n gonuts create secret generic regcred --from-file $HOME/.docker/config.json`


Modify the docker registry repository in the `skaffold.yaml` file.  Example below:


```yaml
apiVersion: skaffold/v2beta9
kind: Config
profiles:
- name: pub
  build:
    artifacts:
    - image: 192.168.1.12:32000/gonuts-pub
      context: pub
      kaniko:
        dockerfile: Dockerfile
        cache:
          repo: 192.168.1.12:32000/gonuts-pub
    insecureRegistries:
    - 192.168.1.12:32000
    cluster:
      namespace: gonuts
      dockerConfig:
        secretName: regcred
  deploy:
    helm:
      releases:
      - name: gonuts-pub
        chartPath: k8s-manifest/pub
        artifactOverrides:
          image.repository: 192.168.1.12:32000/gonuts-pub
        namespace: gonuts
        wait: true
- name: sub
  build:
    artifacts:
    - image: 192.168.1.12:32000/gonuts-sub
      context: sub
      kaniko:
        dockerfile: Dockerfile
        cache:
          repo: 192.168.1.12:32000/gonuts-sub
    insecureRegistries:
    - 192.168.1.12:32000
    cluster:
      namespace: gonuts
      dockerConfig:
        secretName: regcred
  deploy:
    helm:
      releases:
      - name: gonuts-sub
        chartPath: k8s-manifest/sub
        artifactOverrides:
          image.repository: 192.168.1.12:32000/gonuts-sub
        namespace: gonuts
        wait: true

```

#### Build and run the publisher

Go to the project's root directory and execute the command below.

`skaffold run -p pub`

You should see some logs which looks like this.

```shell

$ skaffold run -p pub
Generating tags...
 - 192.168.1.12:32000/gonuts-pub -> 192.168.1.12:32000/gonuts-pub:c0ce3d6-dirty
Checking cache...
 - 192.168.1.12:32000/gonuts-pub: Not found. Building
Creating docker config secret [regcred]...
Building [192.168.1.12:32000/gonuts-pub]...
E1031 01:52:47.968630       1 aws_credentials.go:77] while getting AWS credentials NoCredentialProviders: no valid providers in chain. Deprecated.
	For verbose messaging see aws.Config.CredentialsChainVerboseErrors
INFO[0006] Resolved base name golang:1.15.3 to build    
INFO[0006] Retrieving image manifest golang:1.15.3      
INFO[0006] Retrieving image golang:1.15.3               
INFO[0009] Retrieving image manifest golang:1.15.3      
INFO[0009] Retrieving image golang:1.15.3               
INFO[0012] Retrieving image manifest gcr.io/distroless/base:debug 
. . . .
github.com/nats-io/nats.go/encoders/builtin
golang.org/x/crypto/ed25519
github.com/nats-io/nuid
net
github.com/gogo/protobuf/proto
github.com/nats-io/nkeys
crypto/x509
crypto/tls
github.com/gogo/protobuf/protoc-gen-gogo/descriptor
github.com/gogo/protobuf/gogoproto
github.com/nats-io/stan.go/pb
github.com/nats-io/nats.go/util
github.com/nats-io/nats.go
github.com/nats-io/stan.go
github.com/balchua/gonuts/pub
INFO[0060] Taking snapshot of full filesystem...        
INFO[0063] Pushing layer 192.168.1.12:32000/gonuts-pub:f60f6e0b1e43b1111d97637bb55947f5046d147feedce1f3187f4de7f3d11b67 to cache now 
INFO[0065] Saving file app for later use                
INFO[0065] Deleting filesystem...                       
INFO[0066] Retrieving image manifest gcr.io/distroless/base:debug 
INFO[0066] Retrieving image gcr.io/distroless/base:debug 
INFO[0067] Retrieving image manifest gcr.io/distroless/base:debug 
INFO[0067] Retrieving image gcr.io/distroless/base:debug 
INFO[0069] Executing 0 build triggers                   
INFO[0069] Unpacking rootfs as cmd COPY --from=build /app /app requires it. 
INFO[0071] COPY --from=build /app /app                  
INFO[0071] Taking snapshot of files...                  
INFO[0071] ENV GOTRACEBACK=all                          
INFO[0071] No files changed in this command, skipping snapshotting. 
INFO[0071] ENTRYPOINT ["/app", "-s", "nats://stan-nats-ss.stan.svc.cluster.local:4222", "Test"] 
. . . . 
     

```

**Towards the end, you should see the logs and pub application deployed.**

```shell
Starting deploy...
Helm release gonuts-pub not installed. Installing...
NAME: gonuts-pub
LAST DEPLOYED: Sat Oct 31 10:17:22 2020
NAMESPACE: gonuts
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
1. Get the application URL by running these commands:
Waiting for deployments to stabilize...
 - gonuts:deployment/gonuts-pub is ready.
Deployments stabilized in 1.413569237s  


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
 - 192.168.1.12:32000/gonuts-sub -> 192.168.1.12:32000/gonuts-sub:c0ce3d6-dirty
Checking cache...
 - 192.168.1.12:32000/gonuts-sub: Not found. Building
Creating docker config secret [regcred]...
Building [192.168.1.12:32000/gonuts-sub]...
E1031 02:19:59.968604       1 aws_credentials.go:77] while getting AWS credentials NoCredentialProviders: no valid providers in chain. Deprecated.
	For verbose messaging see aws.Config.CredentialsChainVerboseErrors
INFO[0006] Resolved base name golang:1.15.3 to build    
INFO[0006] Retrieving image manifest golang:1.15.3      
INFO[0006] Retrieving image golang:1.15.3               
INFO[0013] Retrieving image manifest golang:1.15.3      
INFO[0013] Retrieving image golang:1.15.3               
INFO[0017] Retrieving image manifest gcr.io/distroless/base 
INFO[0017] Retrieving image gcr.io/distroless/base      
INFO[0018] Retrieving image manifest gcr.io/distroless/base 
INFO[0018] Retrieving image gcr.io/distroless/base      
INFO[0021] Built cross stage deps: map[0:[/app]]        
INFO[0021] Retrieving image manifest golang:1.15.3      
INFO[0021] Retrieving image golang:1.15.3               
INFO[0024] Retrieving image manifest golang:1.15.3      
INFO[0024] Retrieving image golang:1.15.3               
INFO[0027] Executing 0 build triggers                   
INFO[0027] Using files from context: [/kaniko/buildcontext/go.mod] 
INFO[0027] Using files from context: [/kaniko/buildcontext/go.sum] 
INFO[0027] Checking for cached layer 192.168.1.12:32000/gonuts-sub:c272aee5d3a35296e11ece19cf269d1372497a75a04a99b69345abc2fcc7b0c3... 
INFO[0027] No cached layer found for cmd RUN go mod download 
INFO[0027] Unpacking rootfs as cmd ADD ./go.mod /src/github.com/balchua/gonuts/ requires it. 
INFO[0049] WORKDIR /src/github.com/balchua/gonuts       
INFO[0049] cmd: workdir                                 
INFO[0049] Changed working directory to /src/github.com/balchua/gonuts 
INFO[0049] Creating directory /src/github.com/balchua/gonuts 
INFO[0049] Taking snapshot of files...                  
INFO[0049] Using files from context: [/kaniko/buildcontext/go.mod] 
INFO[0049] ADD ./go.mod /src/github.com/balchua/gonuts/ 
INFO[0049] Taking snapshot of files...                  
INFO[0049] Using files from context: [/kaniko/buildcontext/go.sum] 
INFO[0049] ADD ./go.sum /src/github.com/balchua/gonuts/ 
INFO[0049] Taking snapshot of files...                  


```

Afterwards the subscriber application is successfully deployed.

```shell
Helm release gonuts-sub not installed. Installing...
NAME: gonuts-sub
LAST DEPLOYED: Sat Oct 31 10:21:08 2020
NAMESPACE: gonuts
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
1. Get the application URL by running these commands:
Waiting for deployments to stabilize...
 - gonuts:deployment/gonuts-sub is ready.
Deployments stabilized in 1.28934836s
You can also run [skaffold run --tail] to get the logs


$ kubectl -n gonuts get pods
NAME                          READY   STATUS    RESTARTS   AGE
gonuts-pub-75fbd86dd8-5z6b8   1/1     Running   0          17m
gonuts-sub-55495c99cb-hr28q   1/1     Running   0          30s
```

#### Verify consumer is getting messages.

```shell

$ kubectl -n gonuts logs -f gonuts-sub-55495c99cb-hr28q 
Client ID is 1604110869763282645
Connected to nats://stan-nats-ss.stan.svc.cluster.local:4222 clusterID: [local-stan] clientID: [1604110869763282645]
Listening on [Test], clientID=[1604110869763282645], qgroup=[grp1] durable=[ImDurable]
[#1] Received: sequence:475 subject:"Test" data:"Message is : 2020-10-31 02:21:09.950956691 +0000 UTC m=+226.009217578" timestamp:1604110869951160516 
[#2] Received: sequence:476 subject:"Test" data:"Message is : 2020-10-31 02:21:10.460821108 +0000 UTC m=+226.519082124" timestamp:1604110870461246187 
[#3] Received: sequence:477 subject:"Test" data:"Message is : 2020-10-31 02:21:10.969098067 +0000 UTC m=+227.027359057" timestamp:1604110870969558237 
[#4] Received: sequence:478 subject:"Test" data:"Message is : 2020-10-31 02:21:11.475637567 +0000 UTC m=+227.533898578" timestamp:1604110871476110536 
[#5] Received: sequence:479 subject:"Test" data:"Message is : 2020-10-31 02:21:11.982656597 +0000 UTC m=+228.040917627" timestamp:1604110871983130403 
[#6] Received: sequence:480 subject:"Test" data:"Message is : 2020-10-31 02:21:12.490413572 +0000 UTC m=+228.548674564" timestamp:1604110872490996118 
[#7] Received: sequence:481 subject:"Test" data:"Message is : 2020-10-31 02:21:12.999504956 +0000 UTC m=+229.057765958" timestamp:1604110873000274816 
[#8] Received: sequence:482 subject:"Test" data:"Message is : 2020-10-31 02:21:13.508621089 +0000 UTC m=+229.566882000" timestamp:1604110873508890863 
[#9] Received: sequence:483 subject:"Test" data:"Message is : 2020-10-31 02:21:14.014100946 +0000 UTC m=+230.072361830" timestamp:1604110874014319383 
[#10] Received: sequence:484 subject:"Test" data:"Message is : 2020-10-31 02:21:14.519517361 +0000 UTC m=+230.577778496" timestamp:1604110874520109861 
[#11] Received: sequence:485 subject:"Test" data:"Message is : 2020-10-31 02:21:15.025881262 +0000 UTC m=+231.084142288" timestamp:1604110875026322690 
[#12] Received: sequence:486 subject:"Test" data:"Message is : 2020-10-31 02:21:15.53233631 +0000 UTC m=+231.590597396" timestamp:1604110875532910570 
[#13] Received: sequence:487 subject:"Test" data:"Message is : 2020-10-31 02:21:16.042193082 +0000 UTC m=+232.100454069" timestamp:1604110876042678078 
[#14] Received: sequence:488 subject:"Test" data:"Message is : 2020-10-31 02:21:16.548750654 +0000 UTC m=+232.607011699" timestamp:1604110876549303234 
[#15] Received: sequence:489 subject:"Test" data:"Message is : 2020-10-31 02:21:17.055079978 +0000 UTC m=+233.113340949" timestamp:1604110877055623442 
[#16] Received: sequence:490 subject:"Test" data:"Message is : 2020-10-31 02:21:17.560853054 +0000 UTC m=+233.619114038" timestamp:1604110877561431724 
[#17] Received: sequence:491 subject:"Test" data:"Message is : 2020-10-31 02:21:18.067055717 +0000 UTC m=+234.125316734" timestamp:1604110878067530215 
[#18] Received: sequence:492 subject:"Test" data:"Message is : 2020-10-31 02:21:18.572877241 +0000 UTC m=+234.631138219" timestamp:1604110878573454006 
[#19] Received: sequence:493 subject:"Test" data:"Message is : 2020-10-31 02:21:19.079877334 +0000 UTC m=+235.138138306" timestamp:1604110879080454252 
[#20] Received: sequence:494 subject:"Test" data:"Message is : 2020-10-31 02:21:19.587194439 +0000 UTC m=+235.645455367" timestamp:1604110879587422308 
[#21] Received: sequence:495 subject:"Test" data:"Message is : 2020-10-31 02:21:20.093883361 +0000 UTC m=+236.152144375" timestamp:1604110880094427736 
[#22] Received: sequence:496 subject:"Test" data:"Message is : 2020-10-31 02:21:20.599528333 +0000 UTC m=+236.657789400" timestamp:1604110880600084393 
[#23] Received: sequence:497 subject:"Test" data:"Message is : 2020-10-31 02:21:21.109621159 +0000 UTC m=+237.167882147" timestamp:1604110881110171335 
[#24] Received: sequence:498 subject:"Test" data:"Message is : 2020-10-31 02:21:21.616687622 +0000 UTC m=+237.674948609" timestamp:1604110881617122514 

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

```
$ kubectl apply -f keda-nats-scaler/stan_scaledobject.yaml
```

Check that the `ScaledObject ` is properly installed.

```
$ kubectl -n gonuts get scaledobject
NAME                SCALETARGETKIND      SCALETARGETNAME   TRIGGERS   AUTHENTICATION   READY   ACTIVE   AGE
stan-scaledobject   apps/v1.Deployment   gonuts-sub        stan                        True    True     86s
```

After applying the scaler, you should see the pods scale up.

```shell
$ kubectl -n gonuts get pods
NAME                          READY   STATUS    RESTARTS   AGE
gonuts-pub-cd6d75b8f-brsjj    1/1     Running   0          13m
gonuts-sub-5fbcb7765f-bgddz   1/1     Running   0          9m41s
gonuts-sub-5fbcb7765f-npd6n   1/1     Running   0          40s
gonuts-sub-5fbcb7765f-lsvnx   1/1     Running   0          40s
gonuts-sub-5fbcb7765f-b69kl   1/1     Running   0          40s
gonuts-sub-5fbcb7765f-ml5z8   1/1     Running   0          24s
gonuts-sub-5fbcb7765f-z8qll   1/1     Running   0          24s
gonuts-sub-5fbcb7765f-fjbr2   1/1     Running   0          24s
gonuts-sub-5fbcb7765f-kg9vs   1/1     Running   0          24s
gonuts-sub-5fbcb7765f-wcltj   1/1     Running   0          9s
gonuts-sub-5fbcb7765f-fwp85   1/1     Running   0          9s
gonuts-sub-5fbcb7765f-gjwfk   1/1     Running   0          9s
gonuts-sub-5fbcb7765f-gkfnb   1/1     Running   0          9s
gonuts-sub-5fbcb7765f-xrb5p   1/1     Running   0          9s
gonuts-sub-5fbcb7765f-vk9mq   1/1     Running   0          9s
gonuts-sub-5fbcb7765f-q49f8   1/1     Running   0          9s
gonuts-sub-5fbcb7765f-r8c8f   1/1     Running   0          9s

```

Check the Horizontal Pod Autoscaler status

```
$ kubectl -n gonuts get hpa
NAME                         REFERENCE               TARGETS           MINPODS   MAXPODS   REPLICAS   AGE
keda-hpa-stan-scaledobject   Deployment/gonuts-sub   28367m/10 (avg)   1         30        30         94s
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