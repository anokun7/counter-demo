#!/bin/bash
echo "-------------------- Starting Docker build --------------------"
echo "==================== Building ONBUILD version  ===================="
docker build -t anoop/counter-demo:onbuild -f ../src/Dockerfile ../src
echo "==================== Building MANUAL Part 1 of 2 version  ===================="
docker build -t anoop/counter-demo:v1 -f ../Dockerfile.part1 ..
echo "==================== Copying binaries from temp container of Part 1 ===================="
docker run -d anoop/counter-demo:v1 sleep 5
docker cp $(docker ps -ql):/go/app ..
echo "==================== Building MANUAL Part 2 of 2 version  ===================="
docker build -t anoop/counter-demo:v2 -f ../Dockerfile.part2 ..
echo "==================== Cleaning up ===================="
\rm ../app
echo "==================== Building MULTI-STAGE version  ===================="
docker build -t anoop/counter-demo:latest -f ../Dockerfile ..
echo "-------------------- Finished Docker build --------------------"
echo ""
echo ""
echo "==================== Pushing all images ===================="
docker push ${dockerid:-anoop}/counter-demo:onbuild
docker push ${dockerid:-anoop}/counter-demo:v1
docker push ${dockerid:-anoop}/counter-demo:v2
docker push ${dockerid:-anoop}/counter-demo
echo "==================== Finished pushing all images ===================="
echo "==================== Deploying STACKS ===================="
env=onbuild-dev version=onbuild docker stack deploy -c ../docker-compose.yml OnBuild
env=v1-qa version=v1 docker stack deploy -c ../docker-compose.yml QA1
env=v2-qa version=v2 docker stack deploy -c ../docker-compose.yml QA2
docker stack deploy -c ../docker-compose.yml Prod

echo ""
echo "==================== Running Tests & generating stats ===================="
for e in OnBuild QA1 QA2 Prod;
do
  for i in {0..99};
  do
    until $(curl --output /dev/null --silent --head --fail $(docker service inspect --format '{{range .Endpoint.Ports }}localhost:{{ .PublishedPort }}{{ end }}' ${e}_web)); 
    do
      printf '.'
      sleep 5
    done
    curl -s -o /dev/null $(docker service inspect --format '{{range .Endpoint.Ports }}localhost:{{ .PublishedPort }}{{ end }}' ${e}_web); echo -n "=_";
  done;
  echo "";
done
