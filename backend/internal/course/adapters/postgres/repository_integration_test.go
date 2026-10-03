//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"courses/internal/course/adapters/postgres"
	"courses/internal/course/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRepositoryAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	container, err := postgrescontainer.Run(ctx, "postgres:17.5", postgrescontainer.WithDatabase("courses_test"), postgrescontainer.WithUsername("courses"), postgrescontainer.WithPassword("ephemeral-test-only"), testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(3*time.Minute)))
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
	db, err := sql.Open("pgx", connection)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	directory := filepath.Join("..", "..", "..", "..", "migrations")
	provider, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS(directory), goose.WithDisableGlobalRegistry(true))
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
	repository := postgres.NewRepository(pool)
	query, _ := domain.NewQuery(20, 0, "", "recent")
	page, err := repository.ListPublished(ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	if page.Items == nil || len(page.Items) != 0 || page.Total != 0 {
		t.Fatalf("empty page contract: %+v", page)
	}
	preAdministrationIdentitySeed := `INSERT INTO users (id, email, display_name, role, password_hash) VALUES
	('40000000-0000-4000-8000-000000000001', 'student@courseforge.local', 'Студентка Курсу', 'student', '$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNsdA$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is'),
	('40000000-0000-4000-8000-000000000002', 'instructor@courseforge.local', 'Викладач Курсу', 'instructor', '$argon2id$v=19$m=65536,t=3,p=2$aW5zdHJ1Y3Rvci1zZWVkIQ$zjeAhI2zv+SPciv5gZZegL3BFoWNrJGqMg721CkxRrI'),
	('40000000-0000-4000-8000-000000000003', 'admin@courseforge.local', 'Адміністратор Курсу', 'admin', '$argon2id$v=19$m=65536,t=3,p=2$YWRtaW4tZGVtby1zZWVkIQ$obXgLD59DMQsNm2cOgKWcl050QIrNheSnkrYE3HUvT0')
	ON CONFLICT (id) DO NOTHING`

	t.Run("authoring migration backfill failure and rollback", func(t *testing.T) {
		if _, err := provider.DownTo(ctx, 0); err != nil {
			t.Fatal(err)
		}
		if _, err := provider.UpTo(ctx, 2); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, preAdministrationIdentitySeed); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, "UPDATE users SET disabled_at=updated_at WHERE id='40000000-0000-4000-8000-000000000003'"); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO courses (id,slug,title,summary,description,status)
		SELECT ('50000000-0000-4000-8000-'||lpad(n::text,12,'0'))::uuid,'migration-'||n,'Migration fixture','Fixture summary','Fixture description','draft' FROM generate_series(1,5) AS n`); err != nil {
			t.Fatal(err)
		}
		if _, err := provider.Up(ctx); err != nil {
			t.Fatal(err)
		}
		var userCount, deletedCount int
		if err := db.QueryRowContext(ctx, `SELECT count(*), count(deleted_at) FROM users`).Scan(&userCount, &deletedCount); err != nil {
			t.Fatal(err)
		}
		if userCount != 3 || deletedCount != 0 {
			t.Fatalf("identity administration migration changed users: count=%d deleted=%d", userCount, deletedCount)
		}
		var activeEmailIndex, oldEmailConstraint bool
		if err := db.QueryRowContext(ctx, `SELECT
		EXISTS (
			SELECT 1 FROM pg_indexes
			WHERE schemaname='public' AND tablename='users' AND indexname='users_email_active_key'
				AND indexdef LIKE 'CREATE UNIQUE INDEX%'
				AND indexdef LIKE '%WHERE (deleted_at IS NULL)'
		),
		EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid='users'::regclass AND conname='users_email_key')`).Scan(&activeEmailIndex, &oldEmailConstraint); err != nil {
			t.Fatal(err)
		}
		if !activeEmailIndex || oldEmailConstraint {
			t.Fatalf("identity administration email index: active=%t old constraint=%t", activeEmailIndex, oldEmailConstraint)
		}
		var courseCount, ownerCount int
		var ownerID string
		var ownersEnabled bool
		if err := db.QueryRowContext(ctx, `SELECT count(*), count(DISTINCT courses.instructor_id), min(courses.instructor_id::text), bool_and(users.disabled_at IS NULL)
		FROM courses JOIN users ON users.id=courses.instructor_id`).Scan(&courseCount, &ownerCount, &ownerID, &ownersEnabled); err != nil {
			t.Fatal(err)
		}
		if courseCount != 5 || ownerCount != 1 || ownerID != "40000000-0000-4000-8000-000000000002" || !ownersEnabled {
			t.Fatalf("authoring backfill: courses=%d owners=%d owner=%s enabled=%t", courseCount, ownerCount, ownerID, ownersEnabled)
		}
		var nullable string
		if err := db.QueryRowContext(ctx, "SELECT is_nullable FROM information_schema.columns WHERE table_schema='public' AND table_name='courses' AND column_name='instructor_id'").Scan(&nullable); err != nil {
			t.Fatal(err)
		}
		if nullable != "NO" {
			t.Fatalf("instructor_id nullable: %s", nullable)
		}

		if _, err := provider.DownTo(ctx, 2); err != nil {
			t.Fatal(err)
		}
		var lessonsGone, modulesGone, ownerGone bool
		if err := db.QueryRowContext(ctx, `SELECT to_regclass('course_lessons') IS NULL, to_regclass('course_modules') IS NULL,
		NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='courses' AND column_name='instructor_id')`).Scan(&lessonsGone, &modulesGone, &ownerGone); err != nil {
			t.Fatal(err)
		}
		if !lessonsGone || !modulesGone || !ownerGone {
			t.Fatalf("authoring down migration: lessons gone=%t modules gone=%t owner gone=%t", lessonsGone, modulesGone, ownerGone)
		}
		if _, err := pool.Exec(ctx, "DELETE FROM courses"); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, "DELETE FROM users"); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, "INSERT INTO courses (id,slug,title,summary,description,status) VALUES ('50000000-0000-4000-8000-000000000999','ownerless','Ownerless','Fixture summary','Fixture description','draft')"); err != nil {
			t.Fatal(err)
		}
		if _, err := provider.Up(ctx); err == nil {
			t.Fatal("expected authoring migration to reject courses without users")
		} else {
			var pgError *pgconn.PgError
			if !errors.As(err, &pgError) || pgError.Code != "23502" {
				t.Fatalf("expected not-null violation, got %v", err)
			}
		}
		if err := db.QueryRowContext(ctx, "SELECT count(*) FROM courses").Scan(&courseCount); err != nil {
			t.Fatal(err)
		}
		if courseCount != 1 {
			t.Fatalf("failed migration changed course count: %d", courseCount)
		}
		if err := db.QueryRowContext(ctx, `SELECT to_regclass('course_lessons') IS NULL, to_regclass('course_modules') IS NULL,
		NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='courses' AND column_name='instructor_id')`).Scan(&lessonsGone, &modulesGone, &ownerGone); err != nil {
			t.Fatal(err)
		}
		if !lessonsGone || !modulesGone || !ownerGone {
			t.Fatalf("failed migration changed schema: lessons gone=%t modules gone=%t owner gone=%t", lessonsGone, modulesGone, ownerGone)
		}

		if _, err := pool.Exec(ctx, "DELETE FROM courses"); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, preAdministrationIdentitySeed); err != nil {
			t.Fatal(err)
		}
		if _, err := provider.Up(ctx); err != nil {
			t.Fatal(err)
		}
		if _, err := repository.ListPublished(ctx, query); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("media migration preserves populated database", func(t *testing.T) {
		if _, err := provider.DownTo(ctx, 4); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, preAdministrationIdentitySeed); err != nil {
			t.Fatal(err)
		}
		seed := mustRead(t, filepath.Join("..", "..", "..", "..", "sql", "seed.sql"))
		if _, err := pool.Exec(ctx, string(seed)); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO course_modules (id,course_id,title,position) VALUES
		('61000000-0000-4000-8000-000000000001','10000000-0000-4000-8000-000000000004','Migration module',0);
		INSERT INTO course_lessons (id,module_id,title,description,position) VALUES
		('71000000-0000-4000-8000-000000000001','61000000-0000-4000-8000-000000000001','Migration lesson','Must survive migration 00005',0)`); err != nil {
			t.Fatal(err)
		}
		var userCount, courseCount, moduleCount, lessonCount int
		if err := db.QueryRowContext(ctx, `SELECT
		(SELECT count(*) FROM users),
		(SELECT count(*) FROM courses),
		(SELECT count(*) FROM course_modules),
		(SELECT count(*) FROM course_lessons)`).Scan(&userCount, &courseCount, &moduleCount, &lessonCount); err != nil {
			t.Fatal(err)
		}
		if userCount != 3 || courseCount != 5 || moduleCount != 1 || lessonCount != 1 {
			t.Fatalf("media migration fixture: users=%d courses=%d modules=%d lessons=%d", userCount, courseCount, moduleCount, lessonCount)
		}
		tables := []string{"users", "courses", "course_modules", "course_lessons"}
		before := make(map[string]string, len(tables))
		for _, table := range tables {
			var snapshot string
			if err := db.QueryRowContext(ctx, fmt.Sprintf(`SELECT COALESCE(jsonb_agg(to_jsonb(existing_row) ORDER BY id), '[]'::jsonb)::text FROM (SELECT * FROM %s) AS existing_row`, table)).Scan(&snapshot); err != nil {
				t.Fatal(err)
			}
			before[table] = snapshot
		}
		lessonStructureQuery := `SELECT jsonb_build_object(
		'columns', (SELECT jsonb_agg(jsonb_build_array(column_name, data_type, is_nullable, column_default) ORDER BY ordinal_position) FROM information_schema.columns WHERE table_schema='public' AND table_name='course_lessons'),
		'constraints', (SELECT jsonb_agg(jsonb_build_array(conname, pg_get_constraintdef(oid)) ORDER BY conname) FROM pg_constraint WHERE conrelid='course_lessons'::regclass),
		'indexes', (SELECT jsonb_agg(indexdef ORDER BY indexname) FROM pg_indexes WHERE schemaname='public' AND tablename='course_lessons')
		)::text`
		var lessonStructureBefore string
		if err := db.QueryRowContext(ctx, lessonStructureQuery).Scan(&lessonStructureBefore); err != nil {
			t.Fatal(err)
		}

		if _, err := provider.UpTo(ctx, 5); err != nil {
			t.Fatal(err)
		}
		var videoAssetsExists bool
		var mediaIndexCount int
		if err := db.QueryRowContext(ctx, `SELECT
		to_regclass('video_assets') IS NOT NULL,
		(SELECT count(*) FROM pg_indexes WHERE schemaname='public' AND tablename='video_assets' AND indexname IN (
			'video_assets_pending_per_lesson', 'video_assets_generation', 'video_assets_by_lesson'
		))`).Scan(&videoAssetsExists, &mediaIndexCount); err != nil {
			t.Fatal(err)
		}
		if !videoAssetsExists || mediaIndexCount != 3 {
			t.Fatalf("media migration schema: table=%t indexes=%d", videoAssetsExists, mediaIndexCount)
		}
		for _, table := range tables {
			var after string
			if err := db.QueryRowContext(ctx, fmt.Sprintf(`SELECT COALESCE(jsonb_agg(to_jsonb(existing_row) ORDER BY id), '[]'::jsonb)::text FROM (SELECT * FROM %s) AS existing_row`, table)).Scan(&after); err != nil {
				t.Fatal(err)
			}
			if after != before[table] {
				t.Fatalf("media migration changed pre-existing %s rows", table)
			}
		}
		var lessonStructureAfter string
		if err := db.QueryRowContext(ctx, lessonStructureQuery).Scan(&lessonStructureAfter); err != nil {
			t.Fatal(err)
		}
		if lessonStructureAfter != lessonStructureBefore {
			t.Fatal("media migration changed course_lessons structure")
		}
		if _, err := provider.Up(ctx); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("playback migration preserves populated database", func(t *testing.T) {
		if _, err := provider.DownTo(ctx, 5); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO user_sessions (id,user_id,token_digest,created_at,expires_at) VALUES
		('91000000-0000-4000-8000-000000000001','40000000-0000-4000-8000-000000000001',decode(repeat('ab',32),'hex'),'2026-09-07T00:00:00Z','2026-09-08T00:00:00Z');
		INSERT INTO video_assets (id,lesson_id,generation,state,declared_bytes,original_name,expires_at,created_at,updated_at) VALUES
		('81000000-0000-4000-8000-000000000001','71000000-0000-4000-8000-000000000001',1,'queued',1024,'migration.mp4','2026-09-08T00:00:00Z','2026-09-07T00:00:00Z','2026-09-07T00:00:00Z')`); err != nil {
			t.Fatal(err)
		}
		tables := []string{"users", "courses", "course_modules", "course_lessons", "user_sessions"}
		before := make(map[string]string, len(tables))
		for _, table := range tables {
			var snapshot string
			if err := db.QueryRowContext(ctx, fmt.Sprintf(`SELECT COALESCE(jsonb_agg(to_jsonb(existing_row) ORDER BY id), '[]'::jsonb)::text FROM (SELECT * FROM %s) AS existing_row`, table)).Scan(&snapshot); err != nil {
				t.Fatal(err)
			}
			before[table] = snapshot
		}
		videoAssetSnapshot := `SELECT COALESCE(jsonb_agg(to_jsonb(existing_row) ORDER BY id), '[]'::jsonb)::text
		FROM (SELECT id,lesson_id,generation,state,claim_token,declared_bytes,received_bytes,original_name,attempt,
		failure_reason,renditions,expires_at,evicted_at,created_at,updated_at FROM video_assets) AS existing_row`
		var assetsBefore string
		if err := db.QueryRowContext(ctx, videoAssetSnapshot).Scan(&assetsBefore); err != nil {
			t.Fatal(err)
		}

		if _, err := provider.UpTo(ctx, 6); err != nil {
			t.Fatal(err)
		}
		var progressTable, durationNullable bool
		var clientSeqType string
		var cascadeForeignKeys, progressIndexes int
		if err := db.QueryRowContext(ctx, `SELECT
		to_regclass('lesson_progress') IS NOT NULL,
		EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='video_assets' AND column_name='duration_seconds' AND data_type='numeric' AND is_nullable='YES'),
		(SELECT data_type FROM information_schema.columns WHERE table_schema='public' AND table_name='lesson_progress' AND column_name='client_seq'),
		(SELECT count(*) FROM pg_constraint WHERE conrelid='lesson_progress'::regclass AND contype='f' AND confdeltype='c'),
		(SELECT count(*) FROM pg_indexes WHERE schemaname='public' AND tablename='lesson_progress')`).Scan(
			&progressTable, &durationNullable, &clientSeqType, &cascadeForeignKeys, &progressIndexes,
		); err != nil {
			t.Fatal(err)
		}
		if !progressTable || !durationNullable || clientSeqType != "bigint" || cascadeForeignKeys != 2 || progressIndexes != 1 {
			t.Fatalf("playback schema: table=%t duration_nullable=%t client_seq=%s cascades=%d indexes=%d", progressTable, durationNullable, clientSeqType, cascadeForeignKeys, progressIndexes)
		}
		for _, table := range tables {
			var after string
			if err := db.QueryRowContext(ctx, fmt.Sprintf(`SELECT COALESCE(jsonb_agg(to_jsonb(existing_row) ORDER BY id), '[]'::jsonb)::text FROM (SELECT * FROM %s) AS existing_row`, table)).Scan(&after); err != nil {
				t.Fatal(err)
			}
			if after != before[table] {
				t.Fatalf("playback migration changed pre-existing %s rows", table)
			}
		}
		var assetsAfter string
		if err := db.QueryRowContext(ctx, videoAssetSnapshot).Scan(&assetsAfter); err != nil {
			t.Fatal(err)
		}
		if assetsAfter != assetsBefore {
			t.Fatal("playback migration changed pre-existing video_assets rows")
		}
		if _, err := provider.Up(ctx); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("playback migration lock timeout rolls back and retry succeeds", func(t *testing.T) {
		if _, err := provider.DownTo(ctx, 5); err != nil {
			t.Fatal(err)
		}
		blocker, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		released := false
		t.Cleanup(func() {
			if !released {
				_ = blocker.Rollback()
			}
		})
		if _, err := blocker.ExecContext(ctx, `LOCK TABLE video_assets IN ROW EXCLUSIVE MODE`); err != nil {
			t.Fatal(err)
		}
		if _, err := provider.UpTo(ctx, 6); err == nil {
			t.Fatal("expected playback migration lock timeout")
		} else if !strings.Contains(strings.ToLower(err.Error()), "lock timeout") {
			t.Fatalf("playback migration failure=%v, want lock timeout", err)
		}
		var progressTable, durationColumn bool
		if err := db.QueryRowContext(ctx, `SELECT
		to_regclass('lesson_progress') IS NOT NULL,
		EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='video_assets' AND column_name='duration_seconds')`).Scan(&progressTable, &durationColumn); err != nil {
			t.Fatal(err)
		}
		if progressTable || durationColumn {
			t.Fatalf("timed-out migration left partial schema: table=%t duration=%t", progressTable, durationColumn)
		}
		if err := blocker.Rollback(); err != nil {
			t.Fatal(err)
		}
		released = true
		if _, err := provider.UpTo(ctx, 6); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRowContext(ctx, `SELECT
		to_regclass('lesson_progress') IS NOT NULL,
		EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='video_assets' AND column_name='duration_seconds')`).Scan(&progressTable, &durationColumn); err != nil {
			t.Fatal(err)
		}
		if !progressTable || !durationColumn {
			t.Fatalf("retry did not apply playback schema: table=%t duration=%t", progressTable, durationColumn)
		}
		if _, err := provider.Up(ctx); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("failed migration rolls back its complete transaction", func(t *testing.T) {
		if _, err := provider.DownTo(ctx, 6); err != nil {
			t.Fatal(err)
		}
		initial, err := os.ReadFile(filepath.Join(directory, "00001_courses.sql"))
		if err != nil {
			t.Fatal(err)
		}
		files := fstest.MapFS{
			"00001_courses.sql":        &fstest.MapFile{Data: initial},
			"00002_identity.sql":       &fstest.MapFile{Data: mustRead(t, filepath.Join(directory, "00002_identity.sql"))},
			"00003_authoring.sql":      &fstest.MapFile{Data: mustRead(t, filepath.Join(directory, "00003_authoring.sql"))},
			"00004_identity_admin.sql": &fstest.MapFile{Data: mustRead(t, filepath.Join(directory, "00004_identity_admin.sql"))},
			"00005_media.sql":          &fstest.MapFile{Data: mustRead(t, filepath.Join(directory, "00005_media.sql"))},
			"00006_playback.sql":       &fstest.MapFile{Data: mustRead(t, filepath.Join(directory, "00006_playback.sql"))},
			"00007_failure.sql":        &fstest.MapFile{Data: []byte("-- +goose Up\nCREATE TABLE migration_canary (id integer);\nSELECT 1/0;\n-- +goose Down\nDROP TABLE migration_canary;\n")},
		}
		failure, err := goose.NewProvider(goose.DialectPostgres, db, files, goose.WithDisableGlobalRegistry(true))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := failure.Up(ctx); err == nil {
			t.Fatal("expected failed migration")
		}
		var exists bool
		if err := db.QueryRowContext(ctx, "SELECT to_regclass('migration_canary') IS NOT NULL").Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Fatal("failed migration committed partial schema")
		}
		var progressTable, durationColumn bool
		if err := db.QueryRowContext(ctx, `SELECT
		to_regclass('lesson_progress') IS NOT NULL,
		EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='video_assets' AND column_name='duration_seconds')`).Scan(&progressTable, &durationColumn); err != nil {
			t.Fatal(err)
		}
		if !progressTable || !durationColumn {
			t.Fatalf("failed later migration rolled back playback schema: table=%t duration=%t", progressTable, durationColumn)
		}
		if _, err := provider.Up(ctx); err != nil {
			t.Fatal(err)
		}
		if _, err := repository.ListPublished(ctx, query); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("ready assets require real renditions", func(t *testing.T) {
		for index, test := range []struct {
			name          string
			renditionsSQL string
		}{
			{name: "default empty"},
			{name: "explicit empty", renditionsSQL: `'{}'::text[]`},
			{name: "null element", renditionsSQL: `'{NULL}'::text[]`},
			{name: "empty element", renditionsSQL: `'{""}'::text[]`},
			{name: "mixed null element", renditionsSQL: `'{360p,NULL}'::text[]`},
			{name: "mixed empty element", renditionsSQL: `'{360p,""}'::text[]`},
		} {
			t.Run(test.name, func(t *testing.T) {
				id := fmt.Sprintf("80000000-0000-4000-8000-%012d", index+1)
				statement := `INSERT INTO video_assets (id,generation,state,declared_bytes,original_name,expires_at,created_at,updated_at)
				VALUES ($1,$2,'ready',1024,'source.mp4','2026-09-07T00:00:00Z','2026-09-06T00:00:00Z','2026-09-06T00:00:00Z')`
				if test.renditionsSQL != "" {
					statement = fmt.Sprintf(`INSERT INTO video_assets (id,generation,state,declared_bytes,original_name,renditions,expires_at,created_at,updated_at)
					VALUES ($1,$2,'ready',1024,'source.mp4',%s,'2026-09-07T00:00:00Z','2026-09-06T00:00:00Z','2026-09-06T00:00:00Z')`, test.renditionsSQL)
				}
				_, err := pool.Exec(ctx, statement, id, index+1)
				var pgError *pgconn.PgError
				if !errors.As(err, &pgError) || pgError.Code != "23514" || pgError.ConstraintName != "video_assets_ready_has_renditions" {
					t.Fatalf("expected 23514 video_assets_ready_has_renditions, got %v", err)
				}
				var count int
				if err := db.QueryRowContext(ctx, "SELECT count(*) FROM video_assets WHERE id=$1", id).Scan(&count); err != nil {
					t.Fatal(err)
				}
				if count != 0 {
					t.Fatal("refused ready asset was created")
				}
			})
		}

		if _, err := pool.Exec(ctx, `INSERT INTO video_assets (id,generation,state,declared_bytes,original_name,renditions,expires_at,created_at,updated_at)
		VALUES ('80000000-0000-4000-8000-000000000007',7,'ready',1024,'source.mp4','{360p}','2026-09-07T00:00:00Z','2026-09-06T00:00:00Z','2026-09-06T00:00:00Z')`); err != nil {
			t.Fatalf("valid rendition refused: %v", err)
		}
		var accepted int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM video_assets WHERE id='80000000-0000-4000-8000-000000000007' AND renditions='{360p}'`).Scan(&accepted); err != nil {
			t.Fatal(err)
		}
		if accepted != 1 {
			t.Fatal("valid ready asset was not created")
		}

		if _, err := pool.Exec(ctx, `INSERT INTO video_assets (id,generation,state,declared_bytes,original_name,expires_at,created_at,updated_at)
		VALUES ('80000000-0000-4000-8000-000000000008',8,'queued',1024,'source.mp4','2026-09-07T00:00:00Z','2026-09-06T00:00:00Z','2026-09-06T00:00:00Z')`); err != nil {
			t.Fatal(err)
		}
		_, err := pool.Exec(ctx, "UPDATE video_assets SET state='ready' WHERE id='80000000-0000-4000-8000-000000000008'")
		var pgError *pgconn.PgError
		if !errors.As(err, &pgError) || pgError.Code != "23514" || pgError.ConstraintName != "video_assets_ready_has_renditions" {
			t.Fatalf("expected queued-to-ready 23514 video_assets_ready_has_renditions, got %v", err)
		}
		var state string
		if err := db.QueryRowContext(ctx, "SELECT state FROM video_assets WHERE id='80000000-0000-4000-8000-000000000008'").Scan(&state); err != nil {
			t.Fatal(err)
		}
		if state != "queued" {
			t.Fatalf("failed ready update changed state to %s", state)
		}
	})
	t.Run("concurrent repeated seed and hidden courses", func(t *testing.T) {
		identitySeed, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "sql", "seed-identities.sql"))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, string(identitySeed)); err != nil {
			t.Fatal(err)
		}
		seed, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "sql", "seed.sql"))
		if err != nil {
			t.Fatal(err)
		}
		results := make(chan error, 2)
		for range 2 {
			go func() { _, err := pool.Exec(ctx, string(seed)); results <- err }()
		}
		for range 2 {
			if err := <-results; err != nil {
				t.Fatal(err)
			}
		}
		page, err := repository.ListPublished(ctx, query)
		if err != nil {
			t.Fatal(err)
		}
		if page.Total != 3 || len(page.Items) != 3 || page.Items[0].Slug != "go-backend-foundations" {
			t.Fatalf("seed/visibility/order: %+v", page)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO course_modules (id,course_id,title,position) VALUES
		('60000000-0000-4000-8000-000000000002','10000000-0000-4000-8000-000000000001','PostgreSQL',1),
		('60000000-0000-4000-8000-000000000001','10000000-0000-4000-8000-000000000001','HTTP API',0),
		('60000000-0000-4000-8000-000000000003','10000000-0000-4000-8000-000000000004','Draft secret',0),
		('60000000-0000-4000-8000-000000000004','10000000-0000-4000-8000-000000000005','Archived secret',0);
		INSERT INTO course_lessons (id,module_id,title,description,position) VALUES
		('70000000-0000-4000-8000-000000000002','60000000-0000-4000-8000-000000000001','Contracts','Must stay private',1),
		('70000000-0000-4000-8000-000000000001','60000000-0000-4000-8000-000000000001','Handlers','Must stay private',0),
		('70000000-0000-4000-8000-000000000003','60000000-0000-4000-8000-000000000002','Transactions','Must stay private',0),
		('70000000-0000-4000-8000-000000000004','60000000-0000-4000-8000-000000000003','Draft lesson','Must stay private',0),
		('70000000-0000-4000-8000-000000000005','60000000-0000-4000-8000-000000000004','Archived lesson','Must stay private',0)`); err != nil {
			t.Fatal(err)
		}
		for _, slug := range []string{"draft-course", "archived-course", "missing-course"} {
			if _, err := repository.GetPublished(ctx, slug); !errors.Is(err, domain.ErrCourseNotFound) {
				t.Fatalf("%s: %v", slug, err)
			}
		}
		course, err := repository.GetPublished(ctx, "go-backend-foundations")
		if err != nil {
			t.Fatal(err)
		}
		if course.Title != "Go: фундамент надійного бекенду" || course.Status != domain.Published || course.PublishedAt == nil {
			t.Fatalf("wrong mapped course: %+v", course)
		}
		if len(course.Program) != 2 || course.Program[0].Title != "HTTP API" || course.Program[1].Title != "PostgreSQL" ||
			len(course.Program[0].Lessons) != 2 || course.Program[0].Lessons[0].Title != "Handlers" ||
			course.Program[0].Lessons[1].Title != "Contracts" || len(course.Program[1].Lessons) != 1 ||
			course.Program[1].Lessons[0].Title != "Transactions" {
			t.Fatalf("wrong public program order: %+v", course.Program)
		}
	})
	t.Run("Rt07 Rt09 catalog aggregates count only active reviews without lesson multiplication", func(t *testing.T) {
		// This subtest writes to the shared seeded course; it restores it so later subtests and the
		// catalog's published set see the seed unchanged (AGENTS: whoever adds a counted thing cleans up).
		t.Cleanup(func() {
			for _, statement := range []string{
				`DELETE FROM discussion_reviews WHERE id IN ('80000000-0000-4000-8000-000000000001','80000000-0000-4000-8000-000000000002','80000000-0000-4000-8000-000000000003','80000000-0000-4000-8000-000000000004')`,
				`DELETE FROM users WHERE id='40000000-0000-4000-8000-000000000004'`,
				`UPDATE courses SET status='published', published_at='2026-09-03 10:00:00+00' WHERE id='10000000-0000-4000-8000-000000000001'`,
			} {
				if _, err := pool.Exec(context.Background(), statement); err != nil {
					t.Errorf("cleanup %q: %v", statement, err)
				}
			}
		})
		if _, err := pool.Exec(ctx, `INSERT INTO users (id,email,display_name,role,password_hash) VALUES ('40000000-0000-4000-8000-000000000004','rating-reviewer@courseforge.local','Rating reviewer','student','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA')`); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO discussion_reviews (id,course_id,author_id,stars,body,completed_lessons,total_lessons,created_at) VALUES
			('80000000-0000-4000-8000-000000000001','10000000-0000-4000-8000-000000000001','40000000-0000-4000-8000-000000000001',4,'This review body is deliberately long enough.',0,3,now()),
			('80000000-0000-4000-8000-000000000002','10000000-0000-4000-8000-000000000001','40000000-0000-4000-8000-000000000002',5,'Another review body is deliberately long enough.',0,3,now()),
			('80000000-0000-4000-8000-000000000003','10000000-0000-4000-8000-000000000001','40000000-0000-4000-8000-000000000003',5,'Third review body is deliberately long enough.',0,3,now())`); err != nil {
			t.Fatal(err)
		}
		got, err := repository.GetPublished(ctx, "go-backend-foundations")
		if err != nil {
			t.Fatal(err)
		}
		if got.LessonCount != 3 || got.Rating == nil || got.Rating.Count != 3 || got.Rating.Average != 4.7 {
			t.Fatalf("Rt07 detail=%+v", got.Course)
		}
		// A review that is already removed must not count (Rt09): insert it removed.
		if _, err := pool.Exec(ctx, `INSERT INTO discussion_reviews (id,course_id,author_id,stars,body,completed_lessons,total_lessons,created_at,removed_at) VALUES
			('80000000-0000-4000-8000-000000000004','10000000-0000-4000-8000-000000000001','40000000-0000-4000-8000-000000000004',1,'Removed review body is deliberately long enough.',0,3,now(),now())`); err != nil {
			t.Fatal(err)
		}
		got, err = repository.GetPublished(ctx, "go-backend-foundations")
		if err != nil || got.Rating == nil || got.Rating.Count != 3 {
			t.Fatalf("Rt09 removed extra rating=%+v err=%v", got.Rating, err)
		}
		if _, err := pool.Exec(ctx, `UPDATE discussion_reviews SET removed_at=now() WHERE id='80000000-0000-4000-8000-000000000003'`); err != nil {
			t.Fatal(err)
		}
		got, err = repository.GetPublished(ctx, "go-backend-foundations")
		if err != nil {
			t.Fatal(err)
		}
		if got.Rating != nil {
			t.Fatalf("Rt09 removed review should leave rating null: %+v", got.Rating)
		}
		if _, err := pool.Exec(ctx, `UPDATE courses SET status='archived', published_at=NULL WHERE id='10000000-0000-4000-8000-000000000001'`); err != nil {
			t.Fatal(err)
		}
		if _, err := repository.GetPublished(ctx, "go-backend-foundations"); !errors.Is(err, domain.ErrCourseNotFound) {
			t.Fatalf("archived catalog course: %v", err)
		}
		if _, err := pool.Exec(ctx, `UPDATE courses SET status='published', published_at='2026-09-03 10:00:00+00' WHERE id='10000000-0000-4000-8000-000000000001'`); err != nil {
			t.Fatal(err)
		}
		got, err = repository.GetPublished(ctx, "go-backend-foundations")
		if err != nil || got.Rating != nil {
			t.Fatalf("republished rating=%+v err=%v", got.Rating, err)
		}
	})
	t.Run("database constraints reject invalid entities", func(t *testing.T) {
		for _, statement := range []string{
			"INSERT INTO courses (id,slug,title,summary,description,status) VALUES ('90000000-0000-4000-8000-000000000001','missing-owner','Missing owner','Fixture summary','Fixture description','draft')",
			"UPDATE courses SET slug='invalid--slug' WHERE slug='draft-course'",
			"UPDATE courses SET status='unknown' WHERE slug='draft-course'",
			"UPDATE courses SET status='published' WHERE slug='draft-course'",
			"UPDATE courses SET title='  ' WHERE slug='draft-course'",
			"UPDATE courses SET slug='go-backend-foundations' WHERE slug='draft-course'",
			"UPDATE courses SET updated_at='2020-01-01T00:00:00Z' WHERE slug='draft-course'",
		} {
			if _, err := pool.Exec(ctx, statement); err == nil {
				t.Fatalf("constraint did not reject %s", statement)
			}
		}
	})
	t.Run("one max beyond end and stable ties", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO courses (id,instructor_id,slug,title,summary,description,status,created_at,updated_at,published_at)
		SELECT ('20000000-0000-4000-8000-'||lpad(n::text,12,'0'))::uuid,'40000000-0000-4000-8000-000000000002','bulk-'||n,'Tied title','Bulk fixture','Fixture description','published','2026-09-01T00:00:00Z','2026-09-01T00:00:00Z','2026-09-01T00:00:00Z' FROM generate_series(1,51) AS n`)
		if err != nil {
			t.Fatal(err)
		}
		for _, sort := range []string{"recent", "title"} {
			for _, test := range []struct{ limit, offset, want int }{{1, 0, 1}, {50, 0, 50}, {50, 50, 1}, {20, 10000, 0}} {
				q, _ := domain.NewQuery(test.limit, test.offset, "Tied title", sort)
				page, err := repository.ListPublished(ctx, q)
				if err != nil {
					t.Fatal(err)
				}
				if len(page.Items) != test.want || page.Total != 51 {
					t.Fatalf("%s limit=%d offset=%d: %+v", sort, test.limit, test.offset, page)
				}
				for i, course := range page.Items {
					if course.Slug != fmt.Sprintf("bulk-%d", test.offset+i+1) {
						t.Fatalf("unstable tie: %s", course.Slug)
					}
				}
			}
		}
	})
	t.Run("literal wildcard search", func(t *testing.T) {
		for index, title := range []string{"100% reliable", "snake_case", `path\to!`} {
			_, err := pool.Exec(ctx, `INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,created_at,updated_at,published_at) VALUES($1,'40000000-0000-4000-8000-000000000002',$2,$3,'Literal fixture','Fixture description','published','2026-09-01T00:00:00Z','2026-09-01T00:00:00Z','2026-09-01T00:00:00Z')`, fmt.Sprintf("30000000-0000-4000-8000-%012d", index+1), fmt.Sprintf("literal-%d", index), title)
			if err != nil {
				t.Fatal(err)
			}
		}
		for _, search := range []string{"%", "_", "!", `\`, "RELIABLE"} {
			q, _ := domain.NewQuery(20, 0, search, "recent")
			page, err := repository.ListPublished(ctx, q)
			if err != nil {
				t.Fatal(err)
			}
			if page.Total != 1 || len(page.Items) != 1 {
				t.Fatalf("%q matched nonliteral rows: %+v", search, page)
			}
		}
	})
	t.Run("query cancellation releases the connection", func(t *testing.T) {
		lock, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer lock.Rollback(context.Background())
		if _, err := lock.Exec(ctx, "LOCK TABLE courses IN ACCESS EXCLUSIVE MODE"); err != nil {
			t.Fatal(err)
		}
		request, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		defer cancel()
		started := time.Now()
		if _, err := repository.ListPublished(request, query); !errors.Is(err, domain.ErrDependencyUnavailable) {
			t.Fatalf("got %v", err)
		}
		if time.Since(started) > time.Second {
			t.Fatal("query ignored cancellation")
		}
		if err := lock.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
		if _, err := repository.ListPublished(ctx, query); err != nil {
			t.Fatalf("connection failed after cancellation: %v", err)
		}
	})
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}
