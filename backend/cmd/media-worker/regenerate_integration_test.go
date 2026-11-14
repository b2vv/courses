//go:build integration

package main

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	containers "courses/internal/testsupport/containers"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	regenUser    = "43000000-0000-4000-8000-000000000001"
	regenCourse  = "13000000-0000-4000-8000-000000000001"
	regenModule  = "63000000-0000-4000-8000-000000000001"
	regenLesson  = "73000000-0000-4000-8000-000000000001"
	regenLive    = "85000000-0000-4000-8000-000000000001"
	regenEvicted = "85000000-0000-4000-8000-000000000002"
	regenTrack   = "a5000000-0000-4000-8000-000000000001"
	regenOldTrk  = "a5000000-0000-4000-8000-000000000002"
)

// The master ffmpeg wrote before 00030: the embedded track still says `und`.
const legacyMaster = `#EXTM3U
#EXT-X-VERSION:7
#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="group_aac",NAME="Original",LANGUAGE="und",DEFAULT=YES,AUTOSELECT=YES,CHANNELS="2",URI="audio/index.m3u8"
#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=640x360,CODECS="avc1.64001e,mp4a.40.2",AUDIO="group_aac"
360p/index.m3u8
`

// MS1 M-R7, end to end: populated data at v29 with a real master on disk, migration
// 00030 relabels the legacy track, and the command's own wiring (production
// repository + ffmpeg writer over a media root) rewrites the master — twice, to the
// same bytes — while an evicted asset's master is left alone.
func TestRegenerateMastersAfterLegacyRelabel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	container, err := containers.Start(ctx, func(ctx context.Context) (*postgrescontainer.PostgresContainer, error) {
		return postgrescontainer.Run(ctx, "postgres:17.5",
			postgrescontainer.WithDatabase("regenerate_test"),
			postgrescontainer.WithUsername("courses"),
			postgrescontainer.WithPassword("ephemeral-test-only"),
			testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(3*time.Minute)),
		)
	})
	if err != nil {
		t.Fatalf("PostgreSQL integration requires Docker: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Errorf("terminate test container: %v", err)
		}
	})
	connection, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	database, err := sql.Open("pgx", connection)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	provider, err := goose.NewProvider(goose.DialectPostgres, database, os.DirFS(filepath.Join("..", "..", "migrations")), goose.WithDisableGlobalRegistry(true))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.UpTo(ctx, 29); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, connection)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	for _, statement := range []struct {
		sql       string
		arguments []any
	}{
		{`INSERT INTO users(id,email,display_name,role,password_hash,created_at,updated_at) VALUES($1,'regen@courseforge.local','Regen','instructor','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA',$2,$2)`, []any{regenUser, now}},
		{`INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,created_at,updated_at) VALUES($1,$2,'regen-course','Regen','Summary','Description','draft',$3,$3)`, []any{regenCourse, regenUser, now}},
		{`INSERT INTO course_modules(id,course_id,title,position,created_at,updated_at) VALUES($1,$2,'Module',0,$3,$3)`, []any{regenModule, regenCourse, now}},
		{`INSERT INTO course_lessons(id,module_id,title,position,created_at,updated_at) VALUES($1,$2,'Lesson',0,$3,$3)`, []any{regenLesson, regenModule, now}},
		{`INSERT INTO video_assets(id,lesson_id,generation,state,declared_bytes,received_bytes,original_name,renditions,expires_at,evicted_at,created_at,updated_at)
			VALUES($1,$2,1,'ready',1024,1024,'old.mp4','{360p}',$3,$4,$4,$4),($5,$2,2,'ready',1024,1024,'live.mp4','{360p}',$3,NULL,$4,$4)`,
			[]any{regenEvicted, regenLesson, now.Add(time.Hour), now.Add(-time.Hour), regenLive}},
		{`INSERT INTO audio_tracks(id,video_asset_id,kind,language,label,is_default,state,source_leaf,created_at,updated_at)
			VALUES($1,$2,'embedded','und','Original',true,'ready','original',$4,$4),($3,$5,'embedded','und','Original',true,'ready','original',$4,$4)`,
			[]any{regenTrack, regenLive, regenOldTrk, now, regenEvicted}},
		{`INSERT INTO transcripts(lesson_id,language,is_primary) VALUES($1,'uk',true)`, []any{regenLesson}},
	} {
		if _, err := pool.Exec(ctx, statement.sql, statement.arguments...); err != nil {
			t.Fatalf("fixture %q: %v", statement.sql[:40], err)
		}
	}

	root := t.TempDir()
	masterOf := func(assetID string) string { return filepath.Join(root, "hls", assetID, "master.m3u8") }
	for _, assetID := range []string{regenLive, regenEvicted} {
		if err := os.MkdirAll(filepath.Dir(masterOf(assetID)), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(masterOf(assetID), []byte(legacyMaster), 0o640); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := provider.UpTo(ctx, 30); err != nil {
		t.Fatalf("migration 00030: %v", err)
	}
	var language string
	if err := pool.QueryRow(ctx, `SELECT language FROM audio_tracks WHERE id=$1`, regenTrack).Scan(&language); err != nil || language != "uk" {
		t.Fatalf("live track language after 00030 = %q, %v; want uk", language, err)
	}

	result, err := regenerateMasters(ctx, pool, root)
	if err != nil || result.Regenerated != 1 || len(result.Failed) != 0 || len(result.Skipped) != 0 {
		t.Fatalf("first run = %+v, %v; want exactly the live asset regenerated", result, err)
	}
	first, err := os.ReadFile(masterOf(regenLive))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(first), `LANGUAGE="uk"`) || strings.Contains(string(first), `LANGUAGE="und"`) {
		t.Fatalf("live master after regeneration:\n%s", first)
	}
	if !strings.Contains(string(first), "360p/index.m3u8") {
		t.Fatalf("regeneration lost the video variant:\n%s", first)
	}

	if result, err := regenerateMasters(ctx, pool, root); err != nil || result.Regenerated != 1 {
		t.Fatalf("second run = %+v, %v", result, err)
	}
	second, err := os.ReadFile(masterOf(regenLive))
	if err != nil {
		t.Fatal(err)
	}
	if string(second) != string(first) {
		t.Fatalf("second run changed the master:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
	evicted, err := os.ReadFile(masterOf(regenEvicted))
	if err != nil || string(evicted) != legacyMaster {
		t.Fatalf("evicted master was touched (err=%v):\n%s", err, evicted)
	}
}
