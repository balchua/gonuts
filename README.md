# Learning NATS streaming with Go

## Install nats streaming statefulset.

Go to directory `natss-chart`
`helm install --namespace stan -n stan . `

### starting the subscriber
`sub -s nats://10.152.183.228:4222 -id sample-sub --durable ImDurable -qgroup grp1 Test`

### starting the publisher
`pub -s nats://10.152.183.228:4222 -id sample-pub Test `