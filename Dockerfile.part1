FROM golang:alpine
COPY src/demo.go /go/src
COPY src/tmpl/demo.html /go/tmpl/
RUN apk update && apk add git && \
# need git installed to run 'go get' on github
    go get github.com/garyburd/redigo/redis && \
    go get github.com/gorilla/websocket && \
# statically compile demo.go with all libraries built in
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app /go/src/demo.go

# The lines below are from Dockerfile.part2 & are not really necessary here. 
# They are added merely to make this image (part1) useful to run by itself.
EXPOSE 8080
CMD ["/go/app"]
