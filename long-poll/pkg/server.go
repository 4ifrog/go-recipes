package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const sPrefix = "\033[34;1;4mserver\033[0m"

var events = []string{"empty", "full", "new", "removed"}

func writeJSONResponse(w http.ResponseWriter, payload interface{}) error {
	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = w.Write(jsonData)
	if err != nil {
		return err
	}

	return err
}

func getRandomEvent() string {
	index := rand.Intn(3)
	return events[index]
}

// An event will surface anytime between 1 to 10 seconds and is then sent to the client.
// If time for event to surface > 5 seconds, timeout and return a response with no content.

func rootHandler(w http.ResponseWriter, r *http.Request) {
	timeoutDuration := time.Duration(5) * time.Second
	waitDuration := time.Duration(rand.Intn(9)+1) * time.Second
	waitChan := time.Tick(waitDuration) // Wait this long to pick an event
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), timeoutDuration)

	log.Printf("%s received request and waiting %.1fs to emit an event back to the sender", sPrefix, waitDuration.Seconds())
	select {
	case <-r.Context().Done():
		log.Printf("%s request canceled", sPrefix)
		http.Error(w, "request canceled", http.StatusNotModified)
	case <-timeoutCtx.Done():
		log.Printf("%s time out after %.1f", sPrefix, timeoutDuration.Seconds())
		http.Error(w, fmt.Sprintf("time out after %.1f", timeoutDuration.Seconds()), http.StatusNotModified)
	case <-waitChan:
		event := getRandomEvent()
		w.WriteHeader(http.StatusOK)
		payload := map[string]string{"event": event}
		if err := writeJSONResponse(w, payload); err != nil {
			log.Printf("%s server error: %v", sPrefix, err)
		}
		timeoutCancel()
	}
}

func ListenForMessages() {
	port := 8000
	log.Printf("%s starting web server on port %d", sPrefix, port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), http.HandlerFunc(rootHandler))
	log.Fatal(err)
}

