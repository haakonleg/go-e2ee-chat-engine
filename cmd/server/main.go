package main

import (
	"log"
	"os"

	"net/http"

	"github.com/haakonleg/go-e2ee-chat-engine/server"
	"golang.org/x/net/websocket"
)

var envVars = map[string]string{
	"PORT":         "",
	"MONGODB_URI":  "",
	"MONGODB_NAME": "",
	"FORCE_TLS":    ""}

func checkEnvVars() {
	for k := range envVars {
		val := os.Getenv(k)
		if val == "" {
			log.Fatalf("Error: environment variable %s is not set", k)
		}
		envVars[k] = val
	}
}

// Wrapper that forces every request to use TLS
func forceTLS(server *server.Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-forwarded-proto") != "https" &&
			r.Header.Get("x-forwarded-proto") != "wss" {
			http.NotFound(w, r)
			return
		}

		wsHandler := websocket.Handler(server.WebsockHandler)
		wsHandler.ServeHTTP(w, r)
	}
}

func main() {
	checkEnvVars()

	serverConfig := server.Config{
		DBName:    envVars["MONGODB_NAME"],
		MongoURL:  envVars["MONGODB_URI"],
		Keepalive: 15}

	server := server.CreateServer(serverConfig)

	log.Printf("Listening on port: %s\n", envVars["PORT"])

	if envVars["FORCE_TLS"] == "yes" {
		http.HandleFunc("/", forceTLS(server))
	} else {
		http.Handle("/", websocket.Handler(server.WebsockHandler))
	}
	err := http.ListenAndServe(":"+envVars["PORT"], nil)
	log.Printf("Error occured in http listener: %s\n", err)
}
