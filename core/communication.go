package core

import (
	"errors"
	"log"
	"sync"

	"github.com/glycerine/zygomys/zygo"
	"github.com/gorilla/websocket"
)

type ConnID string
type Topic string

var (
	connections   = map[ConnID]*websocket.Conn{}
	subscriptions = map[Topic]map[ConnID]struct{}{}
	wsMu          sync.RWMutex
)

// RegisterConn registers a new WebSocket connection.
func RegisterConn(id ConnID, conn *websocket.Conn) {
	wsMu.Lock()
	defer wsMu.Unlock()
	connections[id] = conn
}

// UnregisterConn removes a WebSocket connection and unsubscribes it from all topics.
func UnregisterConn(id ConnID) {
	wsMu.Lock()
	defer wsMu.Unlock()
	delete(connections, id)
	for _, subs := range subscriptions {
		delete(subs, id)
	}
}

// fnSubscribe subscribes a connection to a topic.
// Lisp: (subscribe "topic" "conn_id")
func fnSubscribe(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 2 {
		return zygo.SexpNull, zygo.WrongNargs
	}

	topicStr, ok := args[0].(*zygo.SexpStr)
	if !ok {
		return zygo.SexpNull, errors.New("subscribe: first arg must be string")
	}

	connIDStr, ok := args[1].(*zygo.SexpStr)
	if !ok {
		return zygo.SexpNull, errors.New("subscribe: second arg must be string")
	}

	topic := Topic(topicStr.S)
	connID := ConnID(connIDStr.S)

	wsMu.Lock()
	defer wsMu.Unlock()

	if subscriptions[topic] == nil {
		subscriptions[topic] = make(map[ConnID]struct{})
	}
	subscriptions[topic][connID] = struct{}{}

	return zygo.SexpNull, nil
}

// fnBroadcast broadcasts a message to all subscribers of a topic.
// Lisp: (broadcast "topic" "message")
func fnBroadcast(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 2 {
		return zygo.SexpNull, zygo.WrongNargs
	}

	topicStr, ok := args[0].(*zygo.SexpStr)
	if !ok {
		return zygo.SexpNull, errors.New("broadcast: first arg must be string")
	}

	msgStr, ok := args[1].(*zygo.SexpStr)
	if !ok {
		return zygo.SexpNull, errors.New("broadcast: second arg must be string")
	}

	topic := Topic(topicStr.S)

	wsMu.RLock()
	defer wsMu.RUnlock()

	subscribers, ok := subscriptions[topic]
	if !ok {
		log.Printf("[Broadcast] no subscribers for topic %s", topic)
		return zygo.SexpNull, nil
	}

	for connID := range subscribers {
		conn := connections[connID]
		if conn != nil {
			err := conn.WriteMessage(websocket.TextMessage, []byte(msgStr.S))
			if err != nil {
				log.Printf("[Broadcast] error writing to conn %s: %v", connID, err)
			}
		} else {
			log.Printf("[Broadcast] no connection for connID %s", connID)
		}
	}
	return zygo.SexpNull, nil
}

// fnBroadcastAll broadcasts a message to all connected clients.
// Lisp: (broadcastall "message")
func fnBroadcastAll(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 1 {
		return zygo.SexpNull, zygo.WrongNargs
	}

	message, ok := args[0].(*zygo.SexpStr)
	if !ok {
		return zygo.SexpNull, errors.New("broadcastall: arg must be string")
	}

	wsMu.RLock()
	defer wsMu.RUnlock()

	for connID, conn := range connections {
		if conn != nil {
			err := conn.WriteMessage(websocket.TextMessage, []byte(message.S))
			if err != nil {
				log.Printf("broadcastall: failed to write to conn %s: %v", connID, err)
			}
		}
	}

	return zygo.SexpNull, nil
}
