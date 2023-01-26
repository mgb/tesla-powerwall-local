package tesla

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	retry "github.com/avast/retry-go"
	"golang.org/x/net/publicsuffix"

	"github.com/mgb/tesla-powerwall-local/pkg/retrysync"
)

func init() {
	// Blindly accept any SSL cert, required because the server has its own unique certificate.
	// TODO(minegoboom): Look at downloading the cert and storing/allow that.
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

// NewGateway returns a new Gateway proxy
func NewGateway(ipAddress, email, password string, timeoutDuration time.Duration) *Gateway {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}

	return &Gateway{
		ipAddress:       ipAddress,
		email:           email,
		password:        password,
		timeoutDuration: timeoutDuration,

		client: &http.Client{
			Jar: jar,
		},
	}
}

// Gateway contains all the information required to proxy the gateway calls
type Gateway struct {
	ipAddress       string
	email           string
	password        string
	timeoutDuration time.Duration

	login retrysync.Once

	client *http.Client
}

func (t *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx := r.Context()

	var data []byte
	err := retry.Do(
		func() (err error) {
			data, err = t.ProxyCall(ctx, r.URL)
			return err
		},
		retry.Attempts(2),
		retry.Context(ctx),
		retry.RetryIf(func(err error) bool {
			if err == ErrUnauthorized || err == context.Canceled {
				if err := t.Login(ctx); err != nil {
					return false
				}
				return true
			}
			return false
		}),
	)
	if err == ErrUnauthorized {
		t.log("%s", err.Error())
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if ctx.Err() == context.Canceled {
		// If multiple requests are made, the first one will cancel the context causing this to fail.
		// Since multiple gateay calls can be made, this is expected, so no need to log anything.
		http.Error(w, ctx.Err().Error(), http.StatusRequestTimeout)
		return
	}
	if err != nil {
		t.log("%s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		t.log("%s", err.Error())
	}
	w.WriteHeader(http.StatusOK)

	t.log("Proxied called for %s", r.URL.Path)
}

func (t *Gateway) Login(ctx context.Context) error {
	ch := make(chan struct{})

	go func() {
		t.log("Will log in via background process")
		t.login.Do(func() {
			ctx, cancel := context.WithTimeout(context.Background(), t.timeoutDuration)
			defer cancel()

			err := t.runLogin(ctx)
			if err != nil {
				t.log("Failed to login: %s", err.Error())
			}
		})
		close(ch)
	}()

	select {
	case <-ctx.Done():
		t.log("Context failed waiting for login: %s", ctx.Err().Error())
		return ctx.Err()

	case <-ch:
		return nil
	}
}

// Login forces the gatway to login
func (t *Gateway) runLogin(ctx context.Context) error {
	requestBody, err := json.Marshal(struct {
		Username   string `json:"username"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		ForceSmOff bool   `json:"force_sm_off"`
	}{
		Username:   "customer",
		Email:      t.email,
		Password:   t.password,
		ForceSmOff: false,
	})
	if err != nil {
		return err
	}

	u, err := url.Parse(fmt.Sprintf("https://%s/api/login/Basic", t.ipAddress))
	if err != nil {
		return err
	}

	t.log("Attempting to login to: %s", u.String())

	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := checkForError(body).getError(); err != nil {
		return err
	}

	var j struct {
		Token     string    `json:"token"`
		LoginTime time.Time `json:"loginTime"`
	}
	err = json.Unmarshal(body, &j)
	if err != nil {
		return err
	}

	if j.Token == "" {
		return errors.New("token missing, unknown response error")
	}

	t.log("Succesfully logged in with token %q", j.Token)
	return nil
}

// ProxyCall will proxy call the Gateway with the requested URL
func (t *Gateway) ProxyCall(ctx context.Context, path *url.URL) ([]byte, error) {
	u, err := t.convertToTeslaPath(path)
	if err != nil {
		return []byte{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return []byte{}, err
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return data, checkForError(data).getError()
}

func (t *Gateway) log(fmt string, args ...interface{}) {
	log.Printf("[%s] "+fmt, append([]interface{}{t.ipAddress}, args...)...)
}

func (t *Gateway) convertToTeslaPath(path *url.URL) (*url.URL, error) {
	return url.Parse(fmt.Sprintf("https://%s/%s", t.ipAddress, path.Path))
}

func checkForError(body []byte) jsonError {
	j := jsonError{}
	json.Unmarshal(body, &j)
	return j
}
