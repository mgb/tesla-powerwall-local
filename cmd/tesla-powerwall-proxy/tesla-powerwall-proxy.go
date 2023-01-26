package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mgb/tesla-powerwall-local/pkg/httpsproxy"
	"github.com/mgb/tesla-powerwall-local/pkg/tesla"
	"github.com/spf13/pflag"
)

func main() {
	hosts := pflag.StringArrayP("host", "h", nil, "the hostname(:port) of the Tesla Gateway (can have multiple)")
	username := pflag.StringP("username", "u", "", "email address for login")
	password := pflag.StringP("password", "p", "", "password for login")
	listen := pflag.StringP("listen", "l", "localhost:8043", "http server address")
	loginTimeout := pflag.DurationP("login-timeout", "t", 2*time.Minute, "timeout for logging in")
	force := pflag.BoolP("force", "f", false, "force service online, even if hosts or passwords are wrong")
	pflag.Parse()

	if err := validateParameters(*hosts, *username, *password); err != nil {
		log.Fatal(err)
	}

	var successes int
	var handlers httpsproxy.Handlers
	for _, host := range *hosts {
		g := tesla.NewGateway(host, *username, *password, *loginTimeout)
		err := g.Login(context.Background())
		if err != nil {
			log.Printf("[%s] failed to login to the server (check your username/password): %s", host, err.Error())
		} else {
			successes++
		}

		handlers = append(handlers, g)
	}

	// Require at least one successful login. Others can fail (that path maybe offline for now).
	if !*force && successes == 0 {
		log.Fatalf("unable to login to any of the hosts: %s", strings.Join(*hosts, ", "))
	}

	http.Handle("/api/", handlers)
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `Please see <a href="https://github.com/vloschiavo/powerwall2">README.md</a> for API usage. Sample: <a href="/api/meters/aggregates">/api/meters/aggregates</a> and <a href="/api/system_status/soe">/api/system_status/soe</a>`)
	})
	log.Fatal(http.ListenAndServe(*listen, nil))
}

func validateParameters(hosts []string, username string, password string) error {
	if len(hosts) == 0 {
		return errors.New("host is required")
	}
	for _, h := range hosts {
		if h == "" {
			return errors.New("host cannot be empty")
		}
	}
	if username == "" {
		return errors.New("username is required")
	}
	if password == "" {
		return errors.New("password is required")
	}

	return nil
}
