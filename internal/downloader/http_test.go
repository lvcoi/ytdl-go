package downloader

import (
	"sync"
	"testing"
	"time"
)

func TestNewClientForType_Android(t *testing.T) {
	opts := Options{Timeout: 30 * time.Second}
	client := newClientForType("android", opts)
	if client == nil {
		t.Fatal("client is nil")
	}
	if client.HTTP() == nil {
		t.Fatal("HTTP() is nil")
	}
}

func TestNewClientForType_Web(t *testing.T) {
	opts := Options{Timeout: 30 * time.Second}
	client := newClientForType("web", opts)
	if client == nil {
		t.Fatal("client is nil")
	}
}

func TestNewClientForType_ConcurrentSafety(t *testing.T) {
	opts := Options{Timeout: 30 * time.Second}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		clientType := "android"
		if i%2 == 0 {
			clientType = "web"
		}
		go func(ct string) {
			defer wg.Done()
			client := newClientForType(ct, opts)
			if client == nil {
				t.Error("client is nil")
			}
		}(clientType)
	}
	wg.Wait()
}
