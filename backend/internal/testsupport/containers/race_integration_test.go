//go:build integration

// TEMPORARY(testcontainers-go#3868): delete with this package (backlog B-27).
package containers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// raceSessionID isolates this binary's reaper: testcontainers-go reads TESTCONTAINERS_SESSION_ID once per
// process (internal/config, sync.Once), so it must be set before the first testcontainers call.
var raceSessionID string

func TestMain(m *testing.M) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	raceSessionID = hex.EncodeToString(b)
	if err := os.Setenv("TESTCONTAINERS_SESSION_ID", raceSessionID); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

// The real race, made deterministic: a reaper for this session exists in state "created" (exactly what a
// sibling package binary leaves between ContainerCreate and start), and is started one second later. The
// first postgres start must fail with the library's own message, the target it created must be terminated,
// and a retry must reuse the now-running reaper (reaper.go:311-355, docker.go:1360-1361).
func TestStartRetriesTheRealReaperCreatedRace(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	socket := os.Getenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")
	if socket == "" {
		socket = "/var/run/docker.sock"
	}
	// Mirrors newReaper (reaper.go:388-420): image, exposed port, name, labels, socket bind.
	reaper, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "testcontainers/ryuk:0.14.0",
			ExposedPorts: []string{"8080/tcp"},
			Name:         "reaper_" + raceSessionID,
			Labels: map[string]string{
				"org.testcontainers":           "true",
				"org.testcontainers.lang":      "go",
				"org.testcontainers.sessionId": raceSessionID,
				"org.testcontainers.reaper":    "true",
				"org.testcontainers.ryuk":      "true",
			},
			HostConfigModifier: func(hc *container.HostConfig) {
				hc.AutoRemove = true
				hc.Binds = []string{socket + ":/var/run/docker.sock"}
			},
			WaitingFor: wait.ForLog("Started"),
		},
		Started: false,
	})
	if err != nil {
		t.Fatalf("create reaper in state created: %v", err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(reaper) })

	// The reaper is started only once the first attempt has failed, not on a timer: a fixed one-second
	// sleep lost the race on 2026-10-07 (a loaded Docker Desktop reached the status check after ~2 s,
	// found the reaper running, and the test failed with one clean attempt). Start waits for the
	// "Started" log, so the retry that follows meets a running reaper.
	var attempts []error
	pg, err := Start(ctx, func(ctx context.Context) (*postgrescontainer.PostgresContainer, error) {
		c, err := postgrescontainer.Run(ctx, "postgres:17.5",
			postgrescontainer.WithDatabase("race"), postgrescontainer.WithUsername("race"), postgrescontainer.WithPassword("ephemeral-test-only"),
			testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(3*time.Minute)))
		attempts = append(attempts, err)
		if len(attempts) == 1 && err != nil {
			if startErr := reaper.Start(context.Background()); startErr != nil {
				t.Logf("start the created reaper after the first attempt: %v", startErr)
			}
		}
		return c, err
	})
	if err != nil {
		t.Fatalf("Start after the reaper came up: %v (attempts=%d)", err, len(attempts))
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(pg) })

	if len(attempts) < 2 {
		t.Fatalf("expected the first attempt to hit the race, got %d attempt(s): %v", len(attempts), attempts)
	}
	first := attempts[0]
	if first == nil || !strings.Contains(first.Error(), "wait for reaper") || !strings.Contains(first.Error(), `unexpected container status "created"`) {
		t.Fatalf("first attempt was not the reaper race: %v", first)
	}

	// Only the returned target may survive: every earlier attempt's target must have been terminated.
	dockerClient, err := testcontainers.NewDockerClientWithOpts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer dockerClient.Close()
	listed, err := dockerClient.ContainerList(ctx, client.ContainerListOptions{
		All: true,
		Filters: make(client.Filters).
			Add("label", "org.testcontainers.sessionId="+raceSessionID).
			Add("ancestor", "postgres:17.5"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Items) != 1 || listed.Items[0].ID != pg.GetContainerID() {
		ids := make([]string, 0, len(listed.Items))
		for _, item := range listed.Items {
			ids = append(ids, item.ID[:12])
		}
		t.Fatalf("want only the returned postgres %s of this session, found %v after %d attempts", pg.GetContainerID()[:12], ids, len(attempts))
	}
}
