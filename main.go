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

func generateCryptoPrice(ctx context.Context, priceCh chan<- int) {
	r := rand.New(rand.NewSource(time.Now().Unix()))

outerloop:
	for {
		select {
		case <-ctx.Done():
			break outerloop
		default:
			p := r.Intn(100)
			priceCh <- p
		}
		time.Sleep(time.Second)
	}

	close(priceCh)

	fmt.Println("generateCryptoPrice: Finished geenrating")
}

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
