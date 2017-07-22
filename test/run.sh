#!/bin/bash
printf "%100s\n" "------------------------ Starting Docker build --------------------"
printf "%100s\n" "==================== Building caddy (web) ========================="
docker build -t ${dockerid:-anoop}/caddy:latest -f ../caddy/Dockerfile ../caddy
printf "%100s\n" "==================== Building ONBUILD version of app =============="
docker build -t ${dockerid:-anoop}/counter-demo:onbuild -f ../src/Dockerfile ../src
printf "%100s\n" "========== Building MANUAL Part 1 of 2 version of app ============="
docker build -t ${dockerid:-anoop}/counter-demo:v1 -f ../Dockerfile.part1 ..
printf "%100s\n" "=== Copying binaries from temp container of Part 1 ================"
docker run -d ${dockerid:-anoop}/counter-demo:v1 sleep 5
docker cp $(docker ps -ql):/go/app ..
printf "%100s\n" "============ Building MANUAL Part 2 of 2 version of app ==========="
docker build -t ${dockerid:-anoop}/counter-demo:v2 -f ../Dockerfile.part2 ..
printf "%100s\n" "======================= Cleaning up ==============================="
\rm ../app
printf "%100s\n" "==================== Building MULTI-STAGE version of app =========="
docker build -t ${dockerid:-anoop}/counter-demo:latest -f ../Dockerfile ..
printf "%100s\n" "---------------------------- Finished Docker build -----------------"
printf "%100s\n" ""
printf "%100s\n" ""
printf "%100s\n" "========================== Pushing all images ======================"
docker push ${dockerid:-anoop}/caddy
docker push ${dockerid:-anoop}/counter-demo:onbuild
docker push ${dockerid:-anoop}/counter-demo:v1
docker push ${dockerid:-anoop}/counter-demo:v2
docker push ${dockerid:-anoop}/counter-demo
printf "%100s\n\n" "==================== Finished pushing all images ===================="
printf "%100s\n" "============================================ Deploying STACKS ======="
printf "%100s\n" "==================================================== Onbuild Stack =="
env=onbuild-dev version=onbuild docker stack deploy -c ../docker-compose.yml OnBuild${project}
printf "%100s\n" "========================================= QA Part 1 & Part 2 STACKS=="
env=v1-qa version=v1 docker stack deploy -c ../docker-compose.yml QA1${project}
env=v2-qa version=v2 docker stack deploy -c ../docker-compose.yml QA2${project}
printf "%100s\n" "================================================= PRODUCTION STACK =="
docker stack deploy -c ../docker-compose.yml Prod${project}

printf "%100s\n" ""
printf "%100s\n" "================= Running Tests & generating stats ===================="
printf "%100s\n" "=================================== Services coming up, waiting .... =="
for e in OnBuild${project} QA1${project} QA2${project} Prod${project}
do
  for i in {0..49};
  do
    until $(curl --output /dev/null --silent --head --fail $(docker service inspect --format '{{range .Endpoint.Ports }}localhost:{{ .PublishedPort }}{{ end }}' ${e}_web)); 
    do
      printf '.'; sleep 5
    done
    curl -s -o /dev/null $(docker service inspect --format '{{range .Endpoint.Ports }}localhost:{{ .PublishedPort }}{{ end }}' ${e}_web); printf "%2s" "=_"
  done;
  echo ""
done
for e in OnBuild${project} QA1${project} QA2${project} Prod${project}
do
  echo "$e URL: http://$(docker service inspect --format '{{range .Endpoint.Ports }}localhost:{{ .PublishedPort }}{{ end }}' ${e}_web)";
done
