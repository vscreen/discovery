package discovery

import (
	"context"
	"testing"
	"time"
)

func TestPublish(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	s := Service{
		Name: "lukas",
		Type: "_vscreen._tcp",
		Port: 8000,
		Data: make(map[string]string),
	}
	if err := Publish(ctx, &s); err != nil {
		t.Fatal(err)
	}
}
