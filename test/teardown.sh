#!/bin/bash
for i in $(docker ps --filter name=_db. -q); do docker exec $i redis-cli flushall ; done
for i in $(docker network inspect -f '{{range $k, $v := .Containers}} {{$k}}{{end}}' OnBuild_frontend) ; \
do docker network disconnect OnBuild_frontend $i ; \
done
for i in $(docker network inspect -f '{{range $k, $v := .Containers}} {{$k}}{{end}}' OnBuild_backend) ; \
do docker network disconnect OnBuild_backend $i ; \
done
docker stack rm OnBuild QA1 QA2 Prod
