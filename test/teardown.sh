#!/bin/bash
docker stack rm OnBuild QA1 QA2 Prod
echo -n "Waiting.."
for i in {0..9}; do sleep 1 ; echo -n "." ; done
docker ps
docker volume rm OnBuild_data QA1_data QA2_data Prod_data
docker ps
