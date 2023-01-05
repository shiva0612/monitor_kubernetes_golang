package main

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	CheckOrigin:       func(r *http.Request) bool { return true },
	EnableCompression: true,
}

var (
	WSmap *WSMap
)

type WSMap struct {
	mutex *sync.RWMutex
	m     map[string]*websocket.Conn
}

func NewWSmap() *WSMap {
	return &WSMap{
		m:     make(map[string]*websocket.Conn),
		mutex: &sync.RWMutex{},
	}
}

func (wsm *WSMap) getConn(addr string) *websocket.Conn {
	wsm.mutex.RLock()
	defer wsm.mutex.RUnlock()
	return wsm.m[addr]
}

func (wsm *WSMap) storeConn(addr string, conn *websocket.Conn) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()
	prevConn := wsm.m[addr]
	if prevConn != nil {
		prevConn.Close()
		log.Println("closing the prev ws connection on addr: ", addr)
	}
	log.Println("storing new connection on addr: ", addr)
	wsm.m[addr] = conn
}
func (wsm *WSMap) removeConn(addr string) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()
	delete(wsm.m, addr)
}

func NewWebSocket(rw http.ResponseWriter, req *http.Request) (*websocket.Conn, error) {
	return upgrader.Upgrade(rw, req, nil)
}

func writeDatatoWS(ws *websocket.Conn, fetchFunc func(arguments ...any) (any, error), fetchPeriod int, arguments ...any) {

	addr := strings.Split(ws.RemoteAddr().String(), ":")[0]
	WSmap.storeConn(addr, ws)
	defer func() {
		ws.Close()
		log.Println("go routine fetching the details is closed")
	}()
	ticker := time.NewTicker(time.Duration(fetchPeriod) * time.Second)
	for {
		select {
		case <-ticker.C:
			err := writetows(ws, fetchFunc, arguments...)
			if err != nil {
				return
			}
		}
	}
}

func writetows(ws *websocket.Conn, fetchFunc func(arguments ...any) (any, error), arguments ...any) error {
	data, err := fetchFunc(arguments...)
	if err != nil {
		ws.WriteMessage(1, []byte("data could not be fetched from API,closing connection..."))
		return err
	}
	err = ws.WriteJSON(data)
	if err != nil {
		return err
	}
	return nil
}
