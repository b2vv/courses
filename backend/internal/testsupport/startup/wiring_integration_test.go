//go:build integration

package startup_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	containers "courses/internal/testsupport/containers"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	userID   = "49000000-0000-4000-8000-000000000001"
	courseID = "19000000-0000-4000-8000-000000000001"
	moduleID = "69000000-0000-4000-8000-000000000001"
	lessonID = "79000000-0000-4000-8000-000000000001"
	assetID  = "89000000-0000-4000-8000-000000000001"
)

// Exercises the actual executables, not a test copy of their composition code.
// The runtime image is the same ffmpeg-equipped runner provisioned by integration.sh.
func TestExecutableMultilingualWiring(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	netw, err := network.New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := netw.Remove(context.Background()); err != nil {
			t.Error(err)
		}
	})
	pg, err := containers.Start(ctx, func(ctx context.Context) (*postgrescontainer.PostgresContainer, error) {
		return postgrescontainer.Run(ctx, "postgres:17.5", postgrescontainer.WithDatabase("startup"), postgrescontainer.WithUsername("courses"), postgrescontainer.WithPassword("ephemeral-test-only"), network.WithNetwork([]string{"db"}, netw), testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2)))
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := pg.Terminate(context.Background()); err != nil {
			t.Error(err)
		}
	})
	redis, err := containers.Start(ctx, func(ctx context.Context) (testcontainers.Container, error) {
		return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: testcontainers.ContainerRequest{Image: "redis:7.4-bookworm", Cmd: []string{"redis-server", "--requirepass", "ephemeral-test-only"}, Networks: []string{netw.Name}, NetworkAliases: map[string][]string{netw.Name: {"cache"}}, WaitingFor: wait.ForLog("Ready to accept connections")}, Started: true})
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := redis.Terminate(context.Background()); err != nil {
			t.Error(err)
		}
	})
	connection, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("pgx", connection)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS(filepath.Join(root, "migrations")), goose.WithDisableGlobalRegistry(true))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, connection)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if _, err := pool.Exec(ctx, `CREATE FUNCTION startup_notify() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM pg_notify('startup_track', NEW.id::text); RETURN NEW; END $$; CREATE TRIGGER startup_notify AFTER UPDATE ON video_assets FOR EACH ROW EXECUTE FUNCTION startup_notify()`); err != nil {
		t.Fatal(err)
	}
	rawToken := make([]byte, 32)
	for i := range rawToken {
		rawToken[i] = byte(i + 1)
	}
	digest := sha256.Sum256(rawToken)
	token := base64.RawURLEncoding.EncodeToString(rawToken)
	execSQL := func(query string, args ...interface{}) {
		t.Helper()
		if _, err := pool.Exec(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	execSQL(`INSERT INTO users(id,email,display_name,role,password_hash,email_verified,created_at,updated_at) VALUES($1,'startup@example.test','Startup','instructor','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA',true,now(),now())`, userID)
	execSQL(`INSERT INTO user_sessions(id,user_id,token_digest,created_at,expires_at) VALUES('59000000-0000-4000-8000-000000000001',$1,$2,now(),now()+interval '1 hour')`, userID, digest[:])
	execSQL(`INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,created_at,updated_at) VALUES($1,$2,'startup','Startup','Summary','Description','draft',now(),now())`, courseID, userID)
	execSQL(`INSERT INTO course_modules(id,course_id,title,position,created_at,updated_at) VALUES($1,$2,'Module',0,now(),now())`, moduleID, courseID)
	execSQL(`INSERT INTO course_lessons(id,module_id,title,position,created_at,updated_at) VALUES($1,$2,'Lesson',0,now(),now())`, lessonID, moduleID)
	networkIP := func(container testcontainers.Container) string {
		t.Helper()
		info, err := container.Inspect(ctx)
		if err != nil {
			t.Fatal(err)
		}
		endpoint := info.NetworkSettings.Networks[netw.Name]
		if endpoint == nil || !endpoint.IPAddress.IsValid() {
			t.Fatalf("missing fixture network endpoint: %+v", info.NetworkSettings.Networks)
		}
		return endpoint.IPAddress.String()
	}
	pgIP, redisIP := networkIP(pg), networkIP(redis)
	env := map[string]string{"POSTGRES_HOST": pgIP, "POSTGRES_USER": "courses", "POSTGRES_PASSWORD": "ephemeral-test-only", "POSTGRES_DB": "startup", "REDIS_HOST": redisIP, "REDIS_PASSWORD": "ephemeral-test-only", "APP_ENV": "test", "COOKIE_SECURE": "false", "PUBLIC_ORIGIN": "http://localhost", "SMTP_HOST": "localhost", "MAIL_FROM": "startup@example.test", "IDENTITY_2FA_SECRET_KEY": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", "HTTP_ADDR": ":8080"}
	for _, program := range []string{"api", "media-worker"} {
		t.Run(program, func(t *testing.T) {
			binary := filepath.Join(t.TempDir(), program)
			build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/"+program)
			build.Dir = root
			build.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+runtime.GOARCH, "CGO_ENABLED=0")
			if output, err := build.CombinedOutput(); err != nil {
				t.Fatalf("build executable: %v\n%s", err, output)
			}
			for _, want := range []bool{true, false} {
				t.Run(fmt.Sprint(want), func(t *testing.T) {
					currentAssetID := assetID
					if !want {
						currentAssetID = "89000000-0000-4000-8000-000000000002"
					}
					execSQL(`DELETE FROM video_assets WHERE lesson_id=$1`, lessonID)
					state := "ready"
					if program == "media-worker" {
						state = "queued"
					}
					execSQL(`INSERT INTO video_assets(id,lesson_id,generation,state,declared_bytes,received_bytes,original_name,renditions,expires_at,created_at,updated_at) VALUES($1,$2,1,$3,1024,1024,'fixture.mp4',ARRAY['360p'],now()+interval '1 hour',now()-interval '1 hour',now()-interval '1 hour')`, currentAssetID, lessonID, state)
					runEnv := map[string]string{}
					for k, v := range env {
						runEnv[k] = v
					}
					runEnv["ASR_MULTILINGUAL_DEFAULT"] = fmt.Sprint(want)
					command := "mkdir -p /media; exec /fixture-program"
					if program == "media-worker" {
						// Produce a real two-second audio-bearing source, then let startup recovery enqueue it.
						command = "mkdir -p /media/source/" + currentAssetID + "; ffmpeg -v error -f lavfi -i color=c=black:s=640x360:r=10 -f lavfi -i sine=frequency=440 -t 2 -c:v libx264 -pix_fmt yuv420p -c:a aac -f mp4 /media/source/" + currentAssetID + "/original && exec /fixture-program"
					}
					process, err := containers.Start(ctx, func(ctx context.Context) (testcontainers.Container, error) {
						return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: testcontainers.ContainerRequest{Image: "courses-integration-go:1.27.2-ffmpeg", Entrypoint: []string{"/bin/sh", "-c"}, Cmd: []string{command}, Env: runEnv, Files: []testcontainers.ContainerFile{{HostFilePath: binary, ContainerFilePath: "/fixture-program", FileMode: 0755}}, ExposedPorts: []string{"8080/tcp"}, Networks: []string{netw.Name}, WaitingFor: wait.ForHTTP("/readyz").WithPort("8080/tcp").WithStartupTimeout(30 * time.Second)}, Started: true})
					})
					if process != nil {
						defer func() {
							if err := process.Terminate(context.Background()); err != nil {
								t.Error(err)
							}
						}()
					}
					if err != nil {
						pgLogs, _ := pg.Logs(ctx)
						if pgLogs != nil {
							output, _ := io.ReadAll(pgLogs)
							pgLogs.Close()
							t.Logf("postgres logs: %s", output)
						}
						if process != nil {
							logs, _ := process.Logs(ctx)
							if logs != nil {
								defer logs.Close()
								output, _ := io.ReadAll(logs)
								t.Logf("process logs: %s", output)
							}
						}
						t.Fatal(err)
					}
					if program == "api" {
						host, err := process.Host(ctx)
						if err != nil {
							t.Fatal(err)
						}
						port, err := process.MappedPort(ctx, "8080/tcp")
						if err != nil {
							t.Fatal(err)
						}
						endpoint := "http://" + host + ":" + port.Port() + "/api/v1/admin/videos/" + currentAssetID + "/audio-tracks"
						client := &http.Client{Timeout: 5 * time.Second}
						request := func(method, body string) *http.Response {
							t.Helper()
							req, err := http.NewRequestWithContext(ctx, method, endpoint, strings.NewReader(body))
							if err != nil {
								t.Fatal(err)
							}
							req.AddCookie(&http.Cookie{Name: "courseforge_session", Value: token})
							req.Header.Set("Origin", "http://localhost")
							req.Header.Set("Content-Type", "application/json")
							response, err := client.Do(req)
							if err != nil {
								t.Fatal(err)
							}
							return response
						}
						response := request("GET", "")
						var list struct {
							MultilingualDefault *bool `json:"multilingualDefault"`
						}
						err = json.NewDecoder(response.Body).Decode(&list)
						response.Body.Close()
						if response.StatusCode != 200 || err != nil || list.MultilingualDefault == nil || *list.MultilingualDefault != want {
							t.Fatalf("API default=%v status=%d err=%v want=%v", list.MultilingualDefault, response.StatusCode, err, want)
						}
						response = request("POST", `{"declaredBytes":1024,"originalName":"audio.aac","language":"uk","label":"Ukr"}`)
						output, _ := io.ReadAll(response.Body)
						response.Body.Close()
						if response.StatusCode != 201 {
							t.Fatalf("mint status=%d body=%s", response.StatusCode, output)
						}
						var stored bool
						if err := pool.QueryRow(ctx, `SELECT multilingual FROM audio_tracks WHERE video_asset_id=$1`, currentAssetID).Scan(&stored); err != nil {
							t.Fatal(err)
						}
						if stored != want {
							t.Fatalf("API minted multilingual=%v want=%v", stored, want)
						}
					} else {
						// LISTEN/NOTIFY provides a completion event from the persisted production write.
						// The row may already be ready by the time LISTEN is installed, so check it before waiting.
						listener, err := pool.Acquire(ctx)
						if err != nil {
							t.Fatal(err)
						}
						defer listener.Release()
						if _, err := listener.Exec(ctx, "LISTEN startup_track"); err != nil {
							t.Fatal(err)
						}
						done, stop := context.WithTimeout(ctx, 45*time.Second)
						defer stop()
						for {
							var stored bool
							var state string
							err := pool.QueryRow(done, `SELECT a.state,t.multilingual FROM video_assets a LEFT JOIN audio_tracks t ON t.video_asset_id=a.id WHERE a.id=$1`, currentAssetID).Scan(&state, &stored)
							if err == nil && state == "ready" {
								if stored != want {
									t.Fatalf("worker embedded multilingual=%v want=%v", stored, want)
								}
								break
							}
							if state == "failed" {
								t.Fatalf("worker transcode failed: %v", err)
							}
							if _, err := listener.Conn().WaitForNotification(done); err != nil {
								t.Fatalf("embedded track not published: %v (state=%s)", err, state)
							}
						}
					}
				})
			}
		})
	}
}
