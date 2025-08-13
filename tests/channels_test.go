package tests

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/seapvnk/qokl/core"
	"github.com/seapvnk/qokl/server"
)

func setupTestChannel(t *testing.T) *httptest.Server {
	core.InitWS()
	srv := server.New("./")
	ts := httptest.NewServer(srv.Router)
	t.Cleanup(ts.Close)
	return ts
}

func dialWS(t *testing.T, url string) (*websocket.Conn, <-chan string) {
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to connect websocket: %v", err)
	}
	ch := make(chan string, 10)
	go func() {
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				close(ch)
				return
			}
			ch <- string(msg)
		}
	}()
	return ws, ch
}

// waitForNonEmptyMessage waits for a non-empty message up to timeout, logging all received messages.
func waitForNonEmptyMessage(t *testing.T, ch <-chan string, timeout time.Duration) (string, bool) {
	deadline := time.After(timeout)
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				t.Log("channel closed before receiving non-empty message")
				return "", false
			}
			t.Logf("received message: %q", msg)
			if msg != "" {
				return msg, true
			}
		case <-deadline:
			t.Log("timeout waiting for non-empty message")
			return "", false
		}
	}
}

// test if a chat room is working
func TestChatroomSubscribeAndBroadcast(t *testing.T) {
	ts := setupTestChannel(t)
	url := "ws" + strings.TrimPrefix(ts.URL, "http") + "/channels/chatroom"

	ws1, _ := dialWS(t, url)
	defer ws1.Close()
	ws2, ch2 := dialWS(t, url)
	defer ws2.Close()

	time.Sleep(100 * time.Millisecond)

	testMessage := "Hello chatroom!"
	if err := ws1.WriteMessage(websocket.TextMessage, []byte(testMessage)); err != nil {
		t.Fatalf("ws1 failed to send message: %v", err)
	}

	msg, ok := waitForNonEmptyMessage(t, ch2, 3*time.Second)
	if !ok {
		t.Fatal("Timeout waiting for broadcast message on ws2")
	}
	if !strings.HasPrefix(msg, "@") && !strings.Contains(msg, testMessage) {
		t.Errorf("Expected broadcast message containing '@' or %q, got %q", testMessage, msg)
	}
}

// test if private chat its working
func TestPrivateChatToSelf(t *testing.T) {
	ts := setupTestChannel(t)

	connID := "self123"
	url := "ws" + strings.TrimPrefix(ts.URL, "http") + "/channels/private/" + connID + "/chat"

	ws, ch := dialWS(t, url)
	defer ws.Close()

	time.Sleep(150 * time.Millisecond)

	testMsg := "hello to myself"
	err := ws.WriteMessage(websocket.TextMessage, []byte(testMsg))
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	received, ok := waitForNonEmptyMessage(t, ch, 3*time.Second)
	if !ok {
		t.Error("Timeout waiting for broadcast-all message")
	}

	if ! strings.Contains(received, testMsg) {
		t.Errorf("Expected broadcast-all message containing %q, got %q", testMsg, received)
	}
}
