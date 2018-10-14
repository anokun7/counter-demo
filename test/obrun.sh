#!/bin/bash
printf "%100s\n" "------------------------ Starting Docker build --------------------"
printf "%100s\n" "============================ Building web ========================="
docker build -t ${dockerid:-anoop}/webserver:latest -f ../webserver/Dockerfile ../webserver
printf "%100s\n" "==================== Building ONBUILD version of app =============="
docker build -t ${dockerid:-anoop}/counter-demo:onbuild -f ../src/Dockerfile ../src
printf "%100s\n" "========================== Pushing all images ======================"
docker push ${dockerid:-anoop}/webserver
docker push ${dockerid:-anoop}/counter-demo:onbuild
printf "%100s\n\n" "==================== Finished pushing all images ===================="
printf "%100s\n" "============================================ Deploying STACKS ======="
printf "%100s\n" "==================================================== Onbuild Stack =="
env=onbuild-dev version=onbuild docker stack deploy -c ../docker-compose.yml OnBuild${project}
for e in OnBuild${project}
do
  echo "$e URL: http://$(docker service inspect --format '{{range .Endpoint.Ports }}localhost:{{ .PublishedPort }}{{ end }}/counter/' ${e}_web)";
done
printf "%100s\n" ""
printf "%100s\n" "================= Running Tests & generating stats ===================="
printf "%100s\n" "=================================== Services coming up, waiting .... =="
for e in OnBuild${project}
do
  for i in {0..49};
  do
    until $(curl --output /dev/null --silent --head --fail $(docker service inspect --format '{{range .Endpoint.Ports }}localhost:{{ .PublishedPort }}{{ end }}/counter/' ${e}_web)); 
    do
      printf '.'; sleep 5
    done
    curl -s -o /dev/null $(docker service inspect --format '{{range .Endpoint.Ports }}localhost:{{ .PublishedPort }}{{ end }}/counter/' ${e}_web); printf "%2s" "=_"
  done;
  ab -l -d -q -r -S -c 50 -n 1000 $(docker service inspect --format '{{range .Endpoint.Ports }}http://localhost:{{ .PublishedPort }}{{ end }}/counter/' ${e}_web) 2>&1 > /dev/null
  echo ""
done
