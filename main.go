package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

//go:embed index.html
var indexHTML []byte

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(indexHTML)
	})

	http.HandleFunc("/crypto-price", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request received for price...")

		w.Header().Set("Content-Type", "text/event-stream")

		priceCh := make(chan int)

		// Start a go routine that will send price updates the on the price channel.
		// These price updates will be sent back to the client.
		go generateCryptoPrice(r.Context(), priceCh)

		for price := range priceCh {
			event, err := formatServerSentEvent("price-update", price)
			if err != nil {
				fmt.Println(err)
				break
			}

			_, err = fmt.Fprint(w, event)
			if err != nil {
				fmt.Println(err)
				break
			}

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}

		fmt.Println("Finished sending price updates...")
	})

	http.ListenAndServe(":4444", nil)
}

// generateCryptoPrice generates price as random integer and sends it the
// provided channel every 1 second.
func generateCryptoPrice(ctx context.Context, priceCh chan<- int) {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	ticker := time.NewTicker(time.Second)

outerloop:
	for {
		select {
		case <-ctx.Done():
			break outerloop
		case <-ticker.C:
			p := r.Intn(100)
			priceCh <- p
		}
	}

	ticker.Stop()

	close(priceCh)

	fmt.Println("generateCryptoPrice: Finished geenrating")
}

// formatServerSentEvent takes name of an event and any kind of data and transforms
// into a server sent event payload structure.
// Data is sent as a json object, { "data": <your_data> }.
//
// Example:
//
//	Input:
//		event="price-update"
//		data=10
//	Output:
//		event: price-update\n
//		data: "{\"data\":10}"\n\n
func formatServerSentEvent(event string, data any) (string, error) {
	m := map[string]any{
		"data": data,
	}

	buff := bytes.NewBuffer([]byte{})

	e := json.NewEncoder(buff)

	err := e.Encode(m)
	if err != nil {
		return "", err
	}

	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("event: %s\n", event))
	sb.WriteString(fmt.Sprintf("data: %v\n\n", buff.String()))

	return sb.String(), nil
}
