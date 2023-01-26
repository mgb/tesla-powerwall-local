package httpsproxy

import (
	"bytes"
	"net/http"
	"sync"
)

type Handlers []http.Handler

func (h Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup
	ch := make(chan bufferedResponseWriter)
	for _, handler := range h {
		handler := handler
		wg.Add(1)
		go func() {
			defer wg.Done()

			var b bufferedResponseWriter
			handler.ServeHTTP(&b, r)
			ch <- b
		}()
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	var badResponse *bufferedResponseWriter
	for b := range ch {
		if b.statusCode != 0 && b.statusCode != http.StatusOK {
			if badResponse != nil {
				// We only care about the first real failure, ignore the rest
				continue
			}
			if b.statusCode == http.StatusRequestTimeout {
				continue
			}

			// shadow b so it doesn't get overwritten
			b := b
			badResponse = &b
			continue
		}

		// First successful call gets returned, other calls will now get canceled.
		w.Write(b.response.Bytes())
		return
	}

	if badResponse == nil {
		w.WriteHeader(http.StatusExpectationFailed)
		return
	}

	w.WriteHeader(badResponse.statusCode)
	w.Write(badResponse.response.Bytes())
}

// bufferedResponseWriter is a helper struct to buffer the response from the handler
type bufferedResponseWriter struct {
	response   bytes.Buffer
	statusCode int
}

func (b *bufferedResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (b *bufferedResponseWriter) WriteHeader(statusCode int) {
	b.statusCode = statusCode
}

func (b *bufferedResponseWriter) Write(p []byte) (int, error) {
	return b.response.Write(p)
}
