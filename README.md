*mesos-APIproto* is a mockup of the new HTTP API for mesos schedulers

Please refer to the complete [documentation](https://docs.google.com/a/twitter.com/document/d/17EjlrEBEvSBllDC6Xu3BjDoKoGosZpJS0k78JRGx134/edit?usp=sharing "Document") of the API for more information and 
the [JIRA](https://issues.apache.org/jira/browse/MESOS-2288) epic for details on the implementation progress.

## Installation

First get the go dependencies:

```sh
$ go get github.com/jimenez/mesos-APIproto/...
```

Then you can compile and install `mesos-APIproto` with:

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
Once the API is served you can try to subscribe with:

```sh
$ curl -i -X POST localhost:8081/call -H "Content-Type:application/json" --data-binary "@/path/to/go/src/github.com/jimenez/mesos-APIproto/fixtures/subscription.json"
```
and add your recently acquired frameworkId to `fixtures/launchtask.json` then launch a task with:

```sh
$ curl -i -X POST localhost:8081/call -H "Content-Type:application/json" --data-binary "@/path/to/go/src/github.com/jimenez/mesos-APIproto/fixtures/launchtask.json"
```
You can try all the different types of calls, the events sent on the stream you subscribe to will reflect some of the results of your succesful calls.


The prototype can be set to send a limited size of offers, so as to simulate the size of a cluster.

### Functionalities

This prototype is ought to be used to test the consistency of the new HTTP API architecture. 
Master behaviour is off topic.
Although the prototype is meant to mimic the Mesos master<->scheduler new HTTP API behaviour, there are some functionalities that have fixed outcomes. 
Since tasks are not launched in reality, their state is simulated and for practical purposes it will not fail.
Executor behaviour reproduction is completely pretended, messages have to be considered 
succesfully delievered on both ways.
