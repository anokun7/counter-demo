package main

import (
  "github.com/garyburd/redigo/redis"
  "log"
  "os"
  "net/http"
  "html/template"
)

type Hit struct {
  Host string
  Count int
}

type Data struct {
  CurrentHost, Env string
  Hits []Hit
}

var data Data

func handler(w http.ResponseWriter, r *http.Request) {
  var hits []Hit
  host := os.Getenv("HOSTNAME")
  env := os.Getenv("ENVIRONMENT")
  c, err := redis.Dial("tcp", "db:6379")
  if err != nil {
    panic(err)
  }
  defer c.Close()
  c.Do("INCR", host)
  keys, _ := redis.Strings(c.Do("KEYS", "*"))
  for _, key := range keys {
    value, _ := redis.Int(c.Do("GET", key))
    hit := Hit{key, value}
    hits = append(hits, hit)
  }
  data := Data{host, env, hits}
  t, err := template.ParseFiles("tmpl/demo.html")
  if err != nil {
    log.Fatal("Parsing error: ", err)
    return
  }
  exeErr := t.Execute(w, data)
  if exeErr != nil {
    log.Fatal("Execute error: ", exeErr)
  }
}

func main() {
  http.HandleFunc("/", handler)
  http.ListenAndServe(":8080", nil)
}
