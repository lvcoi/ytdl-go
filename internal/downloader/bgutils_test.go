package downloader

import (
	"testing"
	"time"
)

func TestBgUtilsLoad(t *testing.T) {
	bg, err := NewBgUtils()
	if err != nil {
		t.Fatalf("failed to create BgUtils: %v", err)
	}
	if bg == nil {
		t.Fatal("BgUtils is nil")
	}
}

func TestGeneratePoToken(t *testing.T) {
	bg, err := NewBgUtils()
	if err != nil {
		t.Fatalf("failed to create BgUtils: %v", err)
	}

	token, err := bg.GeneratePlaceholder("WEB")
	if err != nil {
		t.Fatalf("failed to generate placeholder: %v", err)
	}
	if token == "" {
		t.Fatal("generated token is empty")
	}
	t.Logf("Generated token: %s", token)
}

func TestNewClientWithBgUtils(t *testing.T) {
	opts := Options{Timeout: 30 * time.Second}
	client := newClient(opts)
	if client == nil {
		t.Fatal("client is nil")
	}
	if client.HTTP() == nil {
		t.Fatal("HTTP() is nil")
	}
}
