# counter-demo
A very simple Go-Redis app to demo discovery of multiple services.

# What you will learn in this demo
* How `docker build` works.
  * Different ways of building & understanding the outcomes
  * Using the new multi-stage build (which is totally awesome btw)
* Deploy each of them on a cluster using docker swarm mode.
  * Understand how to pass "env" specific flags to simulate different environment
  * Test that everything works & perform scale up/down of individual services
