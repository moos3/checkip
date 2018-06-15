FROM golang:onbuild

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && chmod +x /usr/local/bin/dep

RUN mkdir /app
RUN mkdir -p /go/src/github.com/moos3/checkip
WORKDIR /go/src/github.com/moos3/checkip
ADD . /go/src/github.com/moos3/checkip
RUN dep ensure -vendor-only
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o checkip .

EXPOSE 3000

CMD ["/app/checkip"]