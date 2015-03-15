FROM google/golang

WORKDIR /gopath/src/github.com/spirit-contrib/inlet_http_api
ADD . /gopath/src/github.com/spirit-contrib/inlet_http_api/
RUN go get github.com/spirit-contrib/inlet_http_api

CMD []
ENTRYPOINT []