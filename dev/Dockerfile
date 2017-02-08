FROM google/golang

WORKDIR /gopath/src
ADD . /gopath/src/

# set go path for 1.6 +
ENV GOPATH /gopath

# get all dependencies
RUN go get github.com/googollee/go-rest
RUN go get github.com/goamz/goamz/aws
RUN go get github.com/goamz/goamz/s3

# build application
RUN go build tac/docservice
RUN go install tac/docservice

# set env variables

# set port to expose
EXPOSE 8081
CMD []
ENTRYPOINT ["/gopath/bin/docservice"]