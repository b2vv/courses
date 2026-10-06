//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"courses/internal/contracts/access"
	"courses/internal/contracts/access/accesstest"
	containers "courses/internal/testsupport/containers"
	transcriptpostgres "courses/internal/transcript/adapters/postgres"
	"courses/internal/transcript/application"
	"courses/internal/transcript/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	ownerAccount    = "30000000-0000-4000-8000-0000000000a1"
	strangerAccount = "30000000-0000-4000-8000-0000000000a2"

	publishedLesson = "40000000-0000-4000-8000-0000000000b1"
	draftLesson     = "40000000-0000-4000-8000-0000000000b2"
	bareLesson      = "40000000-0000-4000-8000-0000000000b3"
	replacedLesson  = "40000000-0000-4000-8000-0000000000b4"
	uploadingLesson = "40000000-0000-4000-8000-0000000000b5"
	twoLangLesson   = "40000000-0000-4000-8000-0000000000b6"
	reuploadLesson  = "40000000-0000-4000-8000-0000000000b7"
)

func fixtures() []string {
	return []string{
		`INSERT INTO users (id,email,display_name,role,password_hash) VALUES
		 ('` + ownerAccount + `','owner@t.local','O','instructor','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNsdA$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is'),
		 ('` + strangerAccount + `','stranger@t.local','S','student','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNsdA$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is')`,
		// A published course must carry published_at (00001_courses.sql:14), and it must fall
		// between created_at and updated_at (:13) — the fixture obeys the real constraints
		// rather than a convenient version of them.
		`INSERT INTO courses (id,slug,title,summary,description,status,instructor_id,published_at) VALUES
		 ('10000000-0000-4000-8000-0000000000c1','pub','P','S','D','published','` + ownerAccount + `',now()),
		 ('10000000-0000-4000-8000-0000000000c2','draft','D','S','D','draft','` + ownerAccount + `',NULL)`,
		`INSERT INTO course_modules (id,course_id,title,position) VALUES
		 ('20000000-0000-4000-8000-0000000000d1','10000000-0000-4000-8000-0000000000c1','M',0),
		 ('20000000-0000-4000-8000-0000000000d2','10000000-0000-4000-8000-0000000000c2','M',0)`,
		`INSERT INTO course_lessons (id,module_id,title,description,position) VALUES
		 ('` + publishedLesson + `','20000000-0000-4000-8000-0000000000d1','L1','',0),
		 ('` + bareLesson + `','20000000-0000-4000-8000-0000000000d1','L2','',1),
		 ('` + replacedLesson + `','20000000-0000-4000-8000-0000000000d1','L3','',2),
		 ('` + uploadingLesson + `','20000000-0000-4000-8000-0000000000d1','L4','',3),
		 ('` + twoLangLesson + `','20000000-0000-4000-8000-0000000000d1','L5','',4),
		 ('` + reuploadLesson + `','20000000-0000-4000-8000-0000000000d1','L7','',5),
		 ('` + draftLesson + `','20000000-0000-4000-8000-0000000000d2','L6','',0)`,

		// A ready asset must carry renditions (video_assets_ready_has_renditions) — the constraint
		// Phase 5 had to repair because array_length of an empty array is NULL and a NULL CHECK
		// passes. The fixture obeys the repaired rule rather than a convenient one.
		// publishedLesson: written against generation 1, which is still the ready one.
		`INSERT INTO video_assets (id,lesson_id,state,generation,original_name,declared_bytes,renditions,expires_at,created_at,updated_at) VALUES
		 ('50000000-0000-4000-8000-0000000000e1','` + publishedLesson + `','ready',1,'a.mp4',10,'{720p}',now()+interval '1 day',now(),now())`,
		// replacedLesson: written against 1, generation 2 is now the ready one.
		`INSERT INTO video_assets (id,lesson_id,state,generation,original_name,declared_bytes,renditions,expires_at,created_at,updated_at) VALUES
		 ('50000000-0000-4000-8000-0000000000e2','` + replacedLesson + `','ready',1,'a.mp4',10,'{720p}',now()+interval '1 day',now(),now()),
		 ('50000000-0000-4000-8000-0000000000e3','` + replacedLesson + `','ready',2,'b.mp4',10,'{720p}',now()+interval '1 day',now(),now())`,
		// reuploadLesson: written against ready generation 1, and a SECOND upload is in flight.
		// The new asset is not ready, so the player still picks generation 1 and the text is not
		// stale. This is the row that separates the filtered selector from the sorting one: a
		// selector that merely orders ready first would return generation 2 here and call the
		// transcript stale from the moment an upload begins.
		`INSERT INTO video_assets (id,lesson_id,state,generation,original_name,declared_bytes,renditions,expires_at,created_at,updated_at) VALUES
		 ('50000000-0000-4000-8000-0000000000e5','` + reuploadLesson + `','ready',1,'a.mp4',10,'{720p}',now()+interval '1 day',now(),now()),
		 ('50000000-0000-4000-8000-0000000000e6','` + reuploadLesson + `','processing',2,'b.mp4',10,'{}',now()+interval '1 day',now(),now())`,
		// uploadingLesson: text written first, the only asset never became ready.
		`INSERT INTO video_assets (id,lesson_id,state,generation,original_name,declared_bytes,renditions,expires_at,created_at,updated_at) VALUES
		 ('50000000-0000-4000-8000-0000000000e4','` + uploadingLesson + `','processing',1,'a.mp4',10,'{}',now()+interval '1 day',now(),now())`,

		`INSERT INTO transcripts (lesson_id,language,authored_against_generation) VALUES
		 ('` + publishedLesson + `','uk',1),
		 ('` + replacedLesson + `','uk',1),
		 ('` + uploadingLesson + `','uk',0),
		 ('` + reuploadLesson + `','uk',1),
		 ('` + draftLesson + `','uk',0),
		 ('` + twoLangLesson + `','uk',0),
		 ('` + twoLangLesson + `','en',0)`,
		`INSERT INTO course_documents (id,course_id,label,url) VALUES
		 ('60000000-0000-4000-8000-0000000000f1','10000000-0000-4000-8000-0000000000c1','Doc','https://example.test/a')`,
		`INSERT INTO transcript_segments (id,lesson_id,language,ordinal,start_ms,end_ms,text,speaker,confidence) VALUES
		 ('70000000-0000-4000-8000-000000000101','` + publishedLesson + `','uk',0,0,1000,'перша','А',0.9),
		 ('70000000-0000-4000-8000-000000000102','` + publishedLesson + `','uk',1,1000,2000,'друга',NULL,NULL),
		 ('70000000-0000-4000-8000-000000000103','` + replacedLesson + `','uk',0,0,500,'стара',NULL,NULL),
		 ('70000000-0000-4000-8000-000000000104','` + uploadingLesson + `','uk',0,0,500,'рання',NULL,NULL),
		 ('70000000-0000-4000-8000-000000000108','` + reuploadLesson + `','uk',0,0,500,'чинна',NULL,NULL),
		 ('70000000-0000-4000-8000-000000000105','` + draftLesson + `','uk',0,0,500,'чернетка',NULL,NULL),
		 ('70000000-0000-4000-8000-000000000106','` + twoLangLesson + `','uk',0,0,500,'укр',NULL,NULL),
		 ('70000000-0000-4000-8000-000000000107','` + twoLangLesson + `','en',0,0,500,'eng',NULL,NULL)`,
		`INSERT INTO segment_documents (segment_id,document_id,position) VALUES
		 ('70000000-0000-4000-8000-000000000101','60000000-0000-4000-8000-0000000000f1',0)`,
	}
}

func TestTranscriptReadAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	container, err := containers.Start(ctx, func(ctx context.Context) (*postgrescontainer.PostgresContainer, error) {
		return postgrescontainer.Run(ctx, "postgres:17.5",
			postgrescontainer.WithDatabase("transcript_test"),
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
	migrations := filepath.Join("..", "..", "..", "..", "migrations")
	provider, err := goose.NewProvider(goose.DialectPostgres, database, os.DirFS(migrations), goose.WithDisableGlobalRegistry(true))
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
	for _, fixture := range fixtures() {
		if _, err := pool.Exec(ctx, fixture); err != nil {
			t.Fatalf("fixture: %v", err)
		}
	}

	repository := transcriptpostgres.NewRepository(pool)

	t.Run("R08 SQL facts cover the access truth table", func(t *testing.T) {
		for index, test := range accesstest.TruthTable {
			t.Run(test.Name, func(t *testing.T) {
				slug := fmt.Sprintf("access-table-%d", index)
				publishedAt := "NULL"
				if test.Status == "published" {
					publishedAt = "now()"
				}
				var courseID, moduleID, lessonID string
				var price, currency any
				if test.Paid {
					price, currency = int64(50000), "UAH"
				}
				if err := pool.QueryRow(ctx, `INSERT INTO courses(id,slug,title,summary,description,status,instructor_id,published_at,price_minor,currency)
					VALUES(gen_random_uuid(),$1,'Access table','Summary','Description',$2,$3,`+publishedAt+`,$4,$5) RETURNING id::text`, slug, test.Status, ownerAccount, price, currency).Scan(&courseID); err != nil {
					t.Fatal(err)
				}
				if err := pool.QueryRow(ctx, `INSERT INTO course_modules(id,course_id,title,position) VALUES(gen_random_uuid(),$1,'Module',0) RETURNING id::text`, courseID).Scan(&moduleID); err != nil {
					t.Fatal(err)
				}
				if err := pool.QueryRow(ctx, `INSERT INTO course_lessons(id,module_id,title,description,position,is_preview) VALUES(gen_random_uuid(),$1,'Lesson','',0,$2) RETURNING id::text`, moduleID, test.Preview).Scan(&lessonID); err != nil {
					t.Fatal(err)
				}

				actorID := strangerAccount
				administrator := false
				if test.Role == accesstest.Owner {
					actorID = ownerAccount
				} else if test.Role == accesstest.Administrator {
					administrator = true
				}
				if test.Granted {
					if _, err := pool.Exec(ctx, `INSERT INTO course_grants(user_id,course_id,source,state,granted_by) VALUES($1,$2,'admin','active',$3)`, actorID, courseID, ownerAccount); err != nil {
						t.Fatal(err)
					}
				}
				result, err := repository.GetLessonTranscript(ctx, actorID, lessonID, "")
				if err != nil {
					t.Fatal(err)
				}
				if got := result.Course.CanBeViewedBy(domain.Actor{UserID: actorID, Administrator: administrator, EmailVerified: true}); got != test.Want {
					t.Fatalf("facts=%+v actor=%+v CanBeViewedBy()=%v, want %v", result.Course, actorID, got, test.Want)
				}
			})
		}
	})
	t.Run("L3b reads every membership state without treating presence as entitlement", func(t *testing.T) {
		const courseID = "10000000-0000-4000-8000-0000000000c1"
		if _, err := pool.Exec(ctx, `INSERT INTO course_grants(user_id,course_id,source,state,granted_by) VALUES($1,$2,'admin','active',$3)`, strangerAccount, courseID, ownerAccount); err != nil {
			t.Fatal(err)
		}
		for _, want := range []access.Membership{access.MembershipActive, access.MembershipBlocked, access.MembershipRevoked} {
			if want != access.MembershipActive {
				if _, err := pool.Exec(ctx, `UPDATE course_grants SET state=$3 WHERE user_id=$1 AND course_id=$2`, strangerAccount, courseID, string(want)); err != nil {
					t.Fatal(err)
				}
			}
			got, err := repository.GetLessonTranscript(ctx, strangerAccount, publishedLesson, "")
			if err != nil || got.Course.Membership != want || got.Course.Entitled != (want == access.MembershipActive) {
				t.Fatalf("course=%+v err=%v", got.Course, err)
			}
		}
	})
	reading := application.NewReading(repository)
	authoring := application.NewAuthoring(repository)
	student := domain.Actor{UserID: strangerAccount, EmailVerified: true}
	courseOwner := domain.Actor{UserID: ownerAccount, EmailVerified: true}

	t.Run("a transcript reads whole, with its documents", func(t *testing.T) {
		got, err := reading.Get(ctx, student, publishedLesson, "")
		if err != nil {
			t.Fatal(err)
		}
		if !got.Found || len(got.Segments) != 2 {
			t.Fatalf("found=%v segments=%d; want true, 2", got.Found, len(got.Segments))
		}
		if got.Segments[0].Speaker != "А" || got.Segments[0].Confidence == nil || *got.Segments[0].Confidence != 0.9 {
			t.Fatalf("optional fields lost: %+v", got.Segments[0])
		}
		// The segment references by id; the library entry travels once at the top level. This is
		// what keeps the worst-case payload at ~13 MiB instead of ~75 MiB.
		if len(got.Segments[0].DocumentIDs) != 1 {
			t.Fatalf("document reference lost: %+v", got.Segments[0].DocumentIDs)
		}
		if len(got.Documents) != 1 || got.Documents[0].URL != "https://example.test/a" {
			t.Fatalf("library entry lost: %+v", got.Documents)
		}
		// The second segment references nothing. A LEFT JOIN must still return it.
		if len(got.Segments[1].DocumentIDs) != 0 {
			t.Fatalf("segment without documents grew some: %+v", got.Segments[1].DocumentIDs)
		}
		if got.Stale {
			t.Fatal("a transcript written against the current ready asset reads as stale")
		}
	})

	t.Run("a lesson without a transcript is a state, not an error", func(t *testing.T) {
		got, err := reading.Get(ctx, student, bareLesson, "")
		if err != nil {
			t.Fatalf("absent transcript returned an error: %v", err)
		}
		if got.Found {
			t.Fatal("found a transcript that was never written")
		}
	})

	t.Run("staleness is derived, and only a ready asset counts", func(t *testing.T) {
		replaced, err := reading.Get(ctx, student, replacedLesson, "")
		if err != nil || !replaced.Stale {
			t.Fatalf("replaced video: stale=%v err=%v; want stale", replaced.Stale, err)
		}
		// Written before any video existed and the upload never became ready: nothing to compare
		// against, so nothing is stale. A stored flag is what used to get this wrong.
		uploading, err := reading.Get(ctx, student, uploadingLesson, "")
		if err != nil || uploading.Stale {
			t.Fatalf("no ready asset: stale=%v err=%v; want not stale", uploading.Stale, err)
		}
		// A second upload is in flight and not ready yet. The player still plays generation 1, so
		// the text is still current. A selector that only ORDERS ready first would answer
		// generation 2 here and make every re-upload instantly stale.
		reupload, err := reading.Get(ctx, student, reuploadLesson, "")
		if err != nil || reupload.Stale {
			t.Fatalf("upload in flight over a ready asset: stale=%v err=%v; want not stale", reupload.Stale, err)
		}
	})

	t.Run("the access rule refuses, and absence cannot be told from refusal", func(t *testing.T) {
		if _, err := reading.Get(ctx, student, draftLesson, ""); err != domain.ErrForbidden {
			t.Fatalf("stranger on a draft course err = %v; want forbidden", err)
		}
		if got, err := reading.Get(ctx, courseOwner, draftLesson, ""); err != nil || !got.Found {
			t.Fatalf("owner on own draft course: found=%v err=%v; want found", got.Found, err)
		}
		// The lesson that has NO transcript on a draft course must answer forbidden too, not
		// "no transcript" — otherwise the absence of text confirms the course exists.
		var draftBare string
		if err := pool.QueryRow(ctx, `INSERT INTO course_lessons (id,module_id,title,description,position)
			VALUES (gen_random_uuid(),'20000000-0000-4000-8000-0000000000d2','L7','',1) RETURNING id::text`).Scan(&draftBare); err != nil {
			t.Fatal(err)
		}
		if _, err := reading.Get(ctx, student, draftBare, ""); err != domain.ErrForbidden {
			t.Fatalf("stranger on a draft lesson without a transcript err = %v; want forbidden", err)
		}
	})

	t.Run("language selection is deterministic", func(t *testing.T) {
		// Two languages exist; no parameter must not be a lottery.
		first, err := reading.Get(ctx, student, twoLangLesson, "")
		if err != nil {
			t.Fatal(err)
		}
		if first.Language != "en" {
			t.Fatalf("language without a parameter = %q; want the first by normalised code", first.Language)
		}
		asked, err := reading.Get(ctx, student, twoLangLesson, "UK")
		if err != nil || asked.Language != "uk" {
			t.Fatalf("language=UK gave %q err=%v; want uk — normalisation must not depend on the caller's casing", asked.Language, err)
		}
		missing, err := reading.Get(ctx, student, twoLangLesson, "de")
		if err != nil {
			t.Fatalf("absent language returned an error: %v", err)
		}
		if missing.Found {
			t.Fatal("a language that was never written was found")
		}
	})

	runWriteTests(t, ctx, pool, authoring, reading)
	runLibraryTests(t, ctx, application.NewLibrary(repository), authoring)
}

var segmentIDs = [...]string{
	"", "71000000-0000-4000-8000-000000000001", "71000000-0000-4000-8000-000000000002",
	"71000000-0000-4000-8000-000000000003", "71000000-0000-4000-8000-000000000004",
	"71000000-0000-4000-8000-000000000005", "71000000-0000-4000-8000-000000000006",
	"71000000-0000-4000-8000-000000000007", "71000000-0000-4000-8000-000000000008",
	"71000000-0000-4000-8000-000000000009",
}

func newSegmentID(n int) string { return segmentIDs[n] }
