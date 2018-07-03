package rboot

import (
	"fmt"
	"net/http"
	"log"
	"io/ioutil"
	"net"
	"sync"
	"encoding/json"
)

type httpCall struct {
	mux *http.ServeMux

	memoryRead func(key string) []byte
	memorySave func(key string, value []byte)

	listener net.Listener
	outCh    chan Message

	mu    sync.Mutex
	inbox []Message
}

func NewHttpCall(listener net.Listener) *httpCall {
	return &httpCall{
		mux:      http.NewServeMux(),
		listener: listener,
	}
}

func (hc *httpCall) Boot(bot *Robot) {
	hc.memoryRead = bot.MemoRead
	hc.memorySave = bot.MemoSave
	hc.outCh = bot.Outgoing()

	hc.mux.HandleFunc("/pop", hc.httpPop)
	hc.mux.HandleFunc("/send", hc.httpSend)
	hc.mux.HandleFunc("/memoryRead", hc.httpMemoryRead)
	hc.mux.HandleFunc("/memorySave", hc.httpMemorySave)
	srv := &http.Server{Handler: hc.mux}
	srv.SetKeepAlivesEnabled(false)
	go srv.Serve(hc.listener)
}

func (hc *httpCall) httpPop(w http.ResponseWriter, req *http.Request) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	defer req.Body.Close()

	var msg Message
	if len(hc.inbox) > 1 {
		msg, hc.inbox = hc.inbox[0], hc.inbox[1:]
	} else if len(hc.inbox) == 1 {
		msg = hc.inbox[0]
		hc.inbox = []Message{}
	} else if len(hc.inbox) == 0 {
		fmt.Fprint(w, "{}")
		return
	}

	if err := json.NewEncoder(w).Encode(&msg); err != nil {
		log.Fatal(err)
	}
}

func (hc *httpCall) httpSend(w http.ResponseWriter, req *http.Request) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	var msg Message
	if err := json.NewDecoder(req.Body).Decode(&msg); err != nil {
		panic(err)
	}
	defer req.Body.Close()

	go func(m Message) {
		hc.outCh <- m
	}(msg)

	fmt.Fprintln(w, "OK")
}

func (hc *httpCall) httpMemoryRead(w http.ResponseWriter, req *http.Request) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	defer req.Body.Close()

	key := req.URL.Query().Get("key")

	fmt.Fprintf(w, "%s", hc.memoryRead(key))
}

func (hc *httpCall) httpMemorySave(w http.ResponseWriter, req *http.Request) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	defer req.Body.Close()

	key := req.URL.Query().Get("key")

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	hc.memorySave(key, b)
	fmt.Fprintln(w, "OK")
}