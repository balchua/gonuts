# Getting started with NATS Streaming

## Install nats streaming statefulset.


### starting the subscriber
sub -s nats://10.152.183.228:4222 -id sample-sub --durable ImDurable -qgroup grp1 Test

### starting the publisher
pub -s nats://10.152.183.228:4222 -id sample-pub Test 