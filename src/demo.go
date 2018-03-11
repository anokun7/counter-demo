package main

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/websocket"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// data structure to hold the hit counts per host
type Hit struct {
	Host          string
	Count, Active int
}

// global string to hold container's hostname
var host string

// global const string to hold Database URL
const dbURL = "db:6379"

// To sort the hits slice by hostnames
type ByHost []Hit

func (h ByHost) Len() int           { return len(h) }
func (h ByHost) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h ByHost) Less(i, j int) bool { return h[i].Host < h[j].Host && h[i].Active < h[j].Active }

func handler(w http.ResponseWriter, r *http.Request) {
	// Increment only for requests for URL's "/" because chrome
	// seems to make multiple requests for whatever reason
	if r.URL.Path == "/" {
		//log.Println("Incrementing counter...")

		// connect to redis. The redis db host should be reachable as "db"
		// Using "db" as network alias & default port
		c, err := redis.Dial("tcp", dbURL)
		if err != nil {
			panic(err)
		}
		defer c.Close()

		fcontents, err := ioutil.ReadFile("/run/secrets/redis-pass")
		pw := strings.Split(fmt.Sprintf("%s", fcontents), " ")[1]
		if err != nil {
			log.Printf("Could not read password for redis from file")
		} else {
			_, err = c.Do("AUTH", pw)
			if err != nil {
				log.Printf("Authenticating to db failed, using %s", pw)
			} else {
				log.Printf("Success: Authenticated to db successfully")
			}
		}

		// INCR the value corresponding to the host key
		// Prevent incrementing if container is shutting down
		terminated, _ := redis.Int(c.Do("EXISTS", "~"+host))
		if terminated != 1 {
			c.Do("INCR", host)
		}
	}
	stats(w, "incr")
}

func stats(w http.ResponseWriter, context string) {
	// A pseudo variable to denote env specific variations
	env := os.Getenv("ENVIRONMENT")
	rotate := rand.Intn(180)

	// connect to redis.
	c, err := redis.Dial("tcp", dbURL)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	fcontents, err := ioutil.ReadFile("/run/secrets/redis-pass")
	pw := strings.Split(fmt.Sprintf("%s", fcontents), " ")[1]
	if err != nil {
		log.Printf("Could not read password for redis from file")
	} else {
		_, err = c.Do("AUTH", pw)
		if err != nil {
			log.Printf("Authenticating to db failed, using %s", pw)
		} else {
			log.Printf("Success: Authenticated to db successfully")
		}
	}

	// Get running containers only (all except those that begin with '~'
	keys, _ := redis.Strings(c.Do("KEYS", "[^~]*"))

	// Generate stats for all other hits per hosts
	var hits []Hit
	// The total number of hits for any environment
	total := 0
	for _, key := range keys {
		value, _ := redis.Int(c.Do("GET", key))
		//Detect leaks
		terminated, _ := redis.Int(c.Do("EXISTS", "~"+key))
		if terminated == 1 && key == host {
			log.Printf("%s: Found a leak. Deleting\n", key)
			redis.Int(c.Do("DEL", key))
			leakedValue, _ := redis.Int(c.Do("GET", "~"+key))
			redis.Int(c.Do("SET", key, value+leakedValue))
		} else {
			hits = append(hits, Hit{key, value, 1})
		}
		total = total + value
	}

	// Get terminated containers only (all starting with '~')
	tkeys, _ := redis.Strings(c.Do("KEYS", "~*"))

	for _, key := range tkeys {
		value, _ := redis.Int(c.Do("GET", key))
		hits = append(hits, Hit{strings.Trim(key, "~"), value, 0})
		total = total + value
	}

	// Sort the container hostnames so it looks nicer and consistent
	sort.Sort(ByHost(hits))

	// Using an anonymous struct, only needed to pass to the template
	data := struct {
		CurrentHost, Env string
		Rotate           int
		Hits             []Hit
		Context          string
		Total            int
	}{
		host, env, rotate, hits, context, total,
	}

	// Template stuff, with error handling (critical for troubleshooting)
	t, err := template.ParseFiles("tmpl/demo.html")
	if err != nil {
		log.Fatal("Parsing error: ", err)
		return
	}

	// Voila (template magic)
	exeErr := t.Execute(w, data)
	if exeErr != nil {
		log.Fatal("Execute error: ", exeErr)
	}
}

func viewer(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/stats" {
		// log.Println("Viewing stats....")
		stats(w, "viewer")
	}
}

func init() {
	// The container "name" auto-generated by docker
	host = os.Getenv("HOSTNAME")

	// connect to redis. The redis db host should be reachable as "db"
	// Using "db" as network alias & default port
	c, err := redis.Dial("tcp", dbURL)
	if err != nil {
		i := 0
		for {
			// try every two seconds to connect to db
			time.Sleep(2 * time.Second)
			c, err = redis.Dial("tcp", dbURL)
			if err != nil {
				i += 1
				log.Printf("%s: No Connection to db. Attempt %d\n", host, i)
			} else {
				log.Printf("%s: Connection to db established.\n", host)
				break
			}
		}
	}
	defer c.Close()

	fcontents, err := ioutil.ReadFile("/run/secrets/redis-pass")
	pw := strings.Split(fmt.Sprintf("%s", fcontents), " ")[1]
	if err != nil {
		log.Printf("Could not read password for redis from file")
	} else {
		_, err = c.Do("AUTH", pw)
		if err != nil {
			log.Printf("Authenticating to db failed, using %s", pw)
		} else {
			log.Printf("Success: Authenticated to db successfully")
		}
	}

	log.Printf("Initiating counter for %s\n", host)

	// INCR the value corresponding to the host key
	c.Do("SETNX", host, 0)
}

func main() {
	signalChannel := make(chan os.Signal, 1)
	exitChannel := make(chan bool, 1)

	// go routine to watch for signals, for graceful shutdown
	go shutdown(signalChannel, exitChannel)

	http.HandleFunc("/total", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: Websocket launched\n", host)
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("%s: Error upgrading http to websocket: %s\n", host, err)
			return
		}

		c, err := redis.Dial("tcp", dbURL)
		if err != nil {
			log.Printf("%s: Error connecting to db: %s\n", host, err)
		}
		defer c.Close()

		fcontents, err := ioutil.ReadFile("/run/secrets/redis-pass")
		pw := strings.Split(fmt.Sprintf("%s", fcontents), " ")[1]
		if err != nil {
			log.Printf("Could not read password for redis from file")
		} else {
			_, err = c.Do("AUTH", pw)
			if err != nil {
				log.Printf("Authenticating to db failed, using %s", pw)
			} else {
				log.Printf("Success: Authenticated to db successfully")
			}
		}

		for {
			time.Sleep(2 * time.Second)
			// The total number of hits for any environment
			total := 0
			keys, _ := redis.Strings(c.Do("KEYS", "*"))
			for _, key := range keys {
				value, _ := redis.Int(c.Do("GET", key))
				total = total + value
				wsTotal, err := json.Marshal(total)
				if err != nil {
					log.Printf("%s: Error in json.Marshal(): %s\n", host, err)
					return
				}
				conn.WriteMessage(websocket.TextMessage, wsTotal)
			}
		}
	})
	http.HandleFunc("/stats", viewer)
	http.HandleFunc("/", handler)
	server := &http.Server{
		Addr: ":8080",
	}
	log.Printf("%s: Starting counter-demo application...", host)
	log.Fatal(server.ListenAndServe())
	<-exitChannel
	os.Exit(0)
}

func shutdown(signalChannel chan os.Signal, exitChannel chan bool) {
	signal.Notify(signalChannel, syscall.SIGTERM, syscall.SIGINT)
	for {
		signal := <-signalChannel
		switch signal {
		case syscall.SIGINT:
			log.Printf("%s: Received signal: %s. To shutdown, use 'terminate (SIGTERM)' instead.\n", host, signal)
		case syscall.SIGTERM:
			log.Printf("%s: Received signal: %s. Initiating clean up.\n", host, signal)
			cleanup()
			exitChannel <- true
			return
		default:
			log.Printf("%s: Received %s. No handler defined.\n", host, signal)
		}
	}
}

func cleanup() {
	c, err := redis.Dial("tcp", dbURL)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	fcontents, err := ioutil.ReadFile("/run/secrets/redis-pass")
	pw := strings.Split(fmt.Sprintf("%s", fcontents), " ")[1]
	if err != nil {
		log.Printf("Could not read password for redis from file")
	} else {
		_, err = c.Do("AUTH", pw)
		if err != nil {
			log.Printf("Authenticating to db failed, using %s", pw)
		} else {
			log.Printf("Success: Authenticated to db successfully")
		}
	}

	log.Printf("%s: Cleaning up counters for graceful shutdown.\n", host)

	// RENAME the key corresponding to the host key that was shutdown
	// This will allow it to be differentiated from currently running replicas
	c.Do("RENAME", host, "~"+host)
}
