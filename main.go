package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-redis/redis"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})

	key := randSeq(8)
	log.Printf("key set as %s", key)

	pong, err := client.Ping().Result()
	log.Println(pong, err)

	handlers := CounterHandlers{
		client: client,
		key:    key,
	}

	http.HandleFunc("/", hello)
	http.HandleFunc("/increment", handlers.Increment)
	http.HandleFunc("/decrement", handlers.Decrement)
	http.HandleFunc("/count", handlers.Count)

	log.Println("Starting http server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

type CounterHandlers struct {
	client *redis.Client
	key    string
}

func (h CounterHandlers) Increment(w http.ResponseWriter, r *http.Request) {
	val, err := h.client.Incr(h.key).Result()
	if err != nil {
		log.Printf("error incrementing %v", err)
		http.Error(w, "error incrementing", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Incremented count to %d", val)
}

func (h CounterHandlers) Decrement(w http.ResponseWriter, r *http.Request) {
	val, err := h.client.Decr(h.key).Result()
	if err != nil {
		log.Printf("error deccrementing %v", err)
		http.Error(w, "error deccrementing", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Decremented count to %d", val)
}

func (h CounterHandlers) Count(w http.ResponseWriter, r *http.Request) {
	val, err := h.client.Get(h.key).Result()
	if err != nil {
		log.Printf("error retreiving value %v", err)
		http.Error(w, "error retrieving value", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Current count is %s", val)
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `
	<!doctype html>
	<html>
	<head>
		<meta charset="utf-8">
		<title>Welcome</title>
	</head>
	<body>
		<h1>Welcome to the Redis Example</h1>
		<p>You can increment the stored count at <a href="/increment">/increment</a></p>
		<p>You can decrement the stored count at <a href="/decrement">/decrement</a></p>
		<p>You can retrieve the stored count at <a href="/count">/count</a></p>
	</body>
	</html>
`)
}
