//go:build integration

package mail

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"courses/internal/testsupport/containers"
	"courses/platform/config"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestSenderDeliversMultipartMessageToMailpit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	mailpit, err := containers.Start(ctx, func(ctx context.Context) (testcontainers.Container, error) {
		return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:           "axllent/mailpit:v1.31.2",
				ExposedPorts:    []string{"1025/tcp", "8025/tcp"},
				WaitingFor:      wait.ForHTTP("/api/v1/messages").WithPort("8025/tcp"),
				AlwaysPullImage: false,
			},
			Started: true,
		})
	})
	if err != nil {
		t.Fatalf("Mailpit integration requires Docker: %v", err)
	}
	t.Cleanup(func() { _ = mailpit.Terminate(context.Background()) })

	host, err := mailpit.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	smtpPort, err := mailpit.MappedPort(ctx, "1025/tcp")
	if err != nil {
		t.Fatal(err)
	}
	apiPort, err := mailpit.MappedPort(ctx, "8025/tcp")
	if err != nil {
		t.Fatal(err)
	}
	sender := NewSender(config.SMTPConfig{Host: host, Port: smtpPort.Port(), From: "no-reply@courseforge.local"})
	if err := sender.Send(ctx, Message{
		To:       "student@example.com",
		Subject:  "Confirm your email",
		TextBody: "Plain body",
		HTMLBody: "<p>HTML body</p>",
	}); err != nil {
		t.Fatal(err)
	}

	baseURL := "http://" + host + ":" + apiPort.Port()
	var list struct {
		Messages []struct {
			ID      string
			Subject string
			To      []struct{ Address string }
		} `json:"messages"`
	}
	getJSON(t, baseURL+"/api/v1/messages", &list)
	if len(list.Messages) != 1 || list.Messages[0].ID == "" {
		t.Fatalf("messages = %#v", list.Messages)
	}
	if list.Messages[0].Subject != "Confirm your email" || len(list.Messages[0].To) != 1 || list.Messages[0].To[0].Address != "student@example.com" {
		t.Fatalf("message summary = %#v", list.Messages[0])
	}
	var captured struct {
		Text string
		HTML string
	}
	getJSON(t, baseURL+"/api/v1/message/"+list.Messages[0].ID, &captured)
	if captured.Text != "Plain body" || captured.HTML != "<p>HTML body</p>" {
		t.Fatalf("captured = %#v", captured)
	}
}

func getJSON(t *testing.T, url string, target any) {
	t.Helper()
	response, err := http.Get(url) // #nosec G107 -- URL is testcontainers-owned.
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("GET %s: status=%d body=%q", url, response.StatusCode, body)
	}
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatal(fmt.Errorf("decode %s: %w", url, err))
	}
}
