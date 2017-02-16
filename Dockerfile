FROM  daocloud.io/library/golang:1.7

MAINTAINER scnace "scbizu@gmail.com"

ADD . $GOPATH/src/letschat

RUN go install letschat

ENTRYPOINT $GOPATH/bin/letschat

EXPOSE 8090
