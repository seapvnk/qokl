package core

import (
	"errors"
	"log"
	"sync"

	"github.com/glycerine/zygomys/v9/zygo"
	"github.com/olahol/melody"
)

type ConnID string
type Topic string

var (
	connections   = make(map[ConnID]*melody.Session)
	subscriptions = make(map[Topic]map[ConnID]struct{})
	wsMu          sync.RWMutex

	WS *melody.Melody
)

// InitWS initializes the global Melody instance
func InitWS() {
	WS = melody.New()
}

// UseCommunicationModule registers communication-related Lisp functions.
func (vm *VM) UseCommunicationModule() *VM {
	vm.environment.AddFunction("subscribe", fnSubscribe)
	vm.environment.AddFunction("broadcast", fnBroadcast)
	vm.environment.AddFunction("broadcastall", fnBroadcastAll)
	return vm
}

// RegisterConn stores a new Melody session by ConnID.
func RegisterConn(id ConnID, sess *melody.Session) {
	wsMu.Lock()
	defer wsMu.Unlock()
	connections[id] = sess
}

// UnregisterConn removes a Melody session and unsubscribes it from all topics.
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

// fnBroadcast sends a message to all subscribers of a topic.
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
	if !ok || len(subscribers) == 0 {
		log.Printf("[Broadcast] no subscribers for topic %s", topic)
		return zygo.SexpNull, nil
	}

	for connID := range subscribers {
		if sess := connections[connID]; sess != nil && !sess.IsClosed() {
			if err := sess.Write([]byte(msgStr.S)); err != nil {
				log.Printf("[Broadcast] error writing to conn %s: %v", connID, err)
			}
		} else {
			log.Printf("[Broadcast] no active connection for connID %s", connID)
		}
	}

	return zygo.SexpNull, nil
}

// fnBroadcastAll sends a message to all connected sessions.
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

	for connID, sess := range connections {
		if sess != nil && !sess.IsClosed() {
			if err := sess.Write([]byte(message.S)); err != nil {
				log.Printf("[BroadcastAll] failed to write to conn %s: %v", connID, err)
			}
		}
	}

	return zygo.SexpNull, nil
}
