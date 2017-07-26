#!/bin/bash
for i in $(docker ps --filter name=_db. -q); do docker exec $i redis-cli flushall ; done
docker stack rm OnBuild QA1 QA2 Prod
