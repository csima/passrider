FROM alpine
RUN mkdir /go
RUN mkdir /go/bin
ENV GOPATH /go
ENV GOBIN /go/bin
ADD . /go/src/github.com/badkode/passrider
RUN apk add --update go git bash make gcc g++ && \
    go get gopkg.in/redis.v3 && \
    go get github.com/davecgh/go-spew/spew && \
    go get github.com/gorilla/mux && \
    go get github.com/gorilla/sessions && \
    go get github.com/codegangsta/negroni && \
    go get github.com/Jeffail/gabs && \
    go get github.com/twinj/uuid && \
    go get github.com/jinzhu/gorm && \
    go get github.com/jinzhu/gorm/dialects/sqlite && \
    go get github.com/PuerkitoBio/goquery && \
    go get github.com/jasonlvhit/gocron && \
    make -C /go/src/github.com/badkode/passrider install && \
    #go install github.com/badkode/passrider && \
    apk del git go make gcc g++ && \
    rm -rf /go/pkg && \
    rm -rf /go/src && \
    rm -rf /var/cache/apk/*

ENTRYPOINT ["/go/bin/passrider"]