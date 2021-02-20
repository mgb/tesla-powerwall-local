package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/mgb/tesla-powerwall-local/pkg/tesla"
	"github.com/spf13/pflag"
)

func main() {
	host := pflag.StringP("host", "h", "", "the hostname(:port) of the Tesla Gateway")
	username := pflag.StringP("username", "u", "", "email address for login")
	password := pflag.StringP("password", "p", "", "password for login")
	listen := pflag.StringP("listen", "l", "localhost:8043", "http server address")
	pflag.Parse()

	if *host == "" || *username == "" || *password == "" {
		log.Fatal("host, username, and password flags are required")
	}

	g := tesla.NewGateway(*host, *username, *password)
	err := g.Login(context.Background())
	if err != nil {
		log.Fatalf("failed to login to the server (check your username/password): %s", err.Error())
	}

	http.Handle("/api/", g)
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `Please see <a href="https://github.com/vloschiavo/powerwall2">README.md</a> for API usage. Sample: <a href="/api/meters/aggregates">/api/meters/aggregates</a> and <a href="/api/system_status/soe">/api/system_status/soe</a>`)
	})
	log.Fatal(http.ListenAndServe(*listen, nil))
}
