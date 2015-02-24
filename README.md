*mesos-APIproto* is a mockup of the new HTTP API for mesos schedulers

Please refer to the complete documentation of the API for more information:


## Installation

First get the go dependencies:

```sh
$ go get github.com/jimenez/mesos-APIproto/...
```

Then you can compile `volt` with:

```sh
$ go install github.com/jimenez/mesos-APIproto
```

If `$GOPATH/bin` is in your `PATH`, you can invoke `mesos-APIproto` from the CLI.


### Usage

Usage of ./mesos-APIproto:
  -p=8081: Port to listen on
  -s=100: Size of cluster abstracted as number of offers
  -t=0: Failover timeout

```sh
$ ./mesos-APIproto -p <port> -t <failover_timeout> -s <cluster_size>
```

The prototype can be set to send a limited size of offers, so as to simulate the size of a cluster.

### Fonctionalities

This prototype is ought to be used to test the consistency of the new HTTP API architecture. Master behaviour is off topic.
Although the prototype is meant to mimic master new HTTP API behaviour, there are some unctionalities that have fixed outcomes. Since tasks are not launch in reality, their state is simulated and for practical purposes it can never be failed.
Executor behaviour reproduction is completely pretended, messages have to be considered 
succesfully delievered.