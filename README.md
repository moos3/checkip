# CheckIp
This is a simple checkIp application written in go :) 

### Requirements
You will need to install `dep` and that can be done like so
```
          go get -u -v github.com/golang/dep
          go get -u -v github.com/golang/dep/cmd/dep
```
or on mac os you can do ```brew install dep```

### How to run
There are multiple ways to run this but they all first require you doing one of the following:
```go get -u github.com/moos3/checkIp``` or ```git clone git@github.com/moos3/checkIp.git```

After you have the code, its just a matter of doing the following to build it:

```
cd $GOPATH/src/github.com/moos3/checkIp
dep ensure
go build -o app .
```

That will build it for your local os. then its just a matter of doing `./app` and that will start a web server
listening on port 3000 that you can connect too. It will return your IP address and try to look up your infomation.

### Run as docker contianer 
First you will need to build the container. `docker build -t checkIp:latest .` This will build a simple container
and get everything installed for you. Then you can just run it by `docker run --rm -P 3000:3000 checkIp:latest`  This will 
port map the ports for you and make it accessible in your machine's ip on port 3000. 


### Simple Kubernetes Deployment
If you wish to run this in Kubernetes you should make sure build the docker image and push it to the registry of your choice. See there docs
on how to do that. Once that is completed then just save the following and use `helm install -n checkip-latest ./helm-chart/` and this will bring the service
up on your kubernetes cluster.


### Notes
- If running this in production expose it using a ingress. 
- If running this in production make sure you have enabled a HPA to allow for scaling.