#!/bin/bash
for i in $(docker ps --filter name=_db. -q); do docker exec $i redis-cli -a $1 flushall ; done

for p in OnBuild QA1 QA2 Prod
do
  for i in $(docker network inspect -f '{{range $k, $v := .Containers}} {{$k}}{{end}}' ${p}_frontend) ; \
  do docker network disconnect ${p}_frontend $i ; \
  done
  for i in $(docker network inspect -f '{{range $k, $v := .Containers}} {{$k}}{{end}}' ${p}_backend) ; \
  do docker network disconnect ${p}_backend $i ; \
  done
done
docker stack rm OnBuild QA1 QA2 Prod
