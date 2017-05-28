# counter-demo
A very simple Go-Redis app to demo discovery of multiple services.

# What you will learn in this demo
* How `docker build` works.
  * Different ways of building & understanding the outcomes
  * Using the new multi-stage build (which is totally awesome btw)
* Deploy each of them on a docker swarm mode cluster using `docker stack deploy`.
  * Learn how to use various `docker stack` & `docker service` commands to control the services deployed
  * Understand how to pass "env" specific flags to simulate different environment
  * Test that everything works & perform scale up/down of individual services
  * Deploy the app on a DDC cluster running on Docker EE

# About the application
The counter-demo app is probably the simplest dynamic web application ever built. It has a front end, running a webserver all written in `golang` (Go) and whenever it responds to a request, it increments a counter that is tracked in a redis database. Each counter is unique to the host (or container) on which the webserver is running on. 
The stats as to which host / container was accessed how many times based on the counter is displayed in a tabular format. Additionally, there is an indicator that displays the particular env version of the application deployed.

# Building the app
 Below are the steps to build the docker image for each flavor of the application. The only component we will need to build is for the web server / front end application. The back end which is on redis will simply use the official docker image from the store. Details on that are available at https://store.docker.com/images/redis?tab=description.

## Pre-requisites
- You should have docker installed (of course). You can use Docker for Mac (D4M) or Windows or you can use a Linux machine that has docker. If you use Windows, please note you will not be able to run some of the scripts in the test folder.
- You should have a swarm cluster available. This should be the newer swarm mode (not classic swarm), using `docker swarm init`. It does not matter how many nodes are members of the cluster, you can setup a swarm mode cluster using a single node using D4M or D4W.

## Steps to build the application (front end component):
 - The old way, very simple but not efficient in terms of resultant image size:

 - The old way, improved - uses a two step process. The first step is to build just like before, but there is a subsequent step which extracts only the useful bits from the first image and constructs a new minimal image from those bits.

 - The new, by far the best way: Also known as a multi-stage build. What we did in the second way above is all automated using native docker (or Dockerfile) instructions. It is possible to have as many steps as necessary to arrive at the final image.

