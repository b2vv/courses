//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"courses/internal/transcript/application"
	"courses/internal/transcript/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

func runM2R10InventoryTests(t *testing.T, ctx context.Context, pool *pgxpool.Pool, authoring *application.Authoring) {
	t.Run("M2-R10 inventory zero one two languages, stale Ready replacement and foreign course", func(t *testing.T) {
		// A private lesson avoids coupling to mutations performed by earlier shared-fixture tests.
		var lesson string
		if err := pool.QueryRow(ctx, `INSERT INTO course_lessons(id,module_id,title,position)
   SELECT gen_random_uuid(),module_id,'Inventory',99 FROM course_lessons WHERE id=$1 RETURNING id::text`, bareLesson).Scan(&lesson); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if _, err := pool.Exec(ctx, `DELETE FROM course_lessons WHERE id=$1`, lesson); err != nil {
				t.Error(err)
			}
		})
		owner := domain.Actor{UserID: ownerAccount}
		check := func(count int, stale bool) {
			t.Helper()
			items, err := authoring.List(ctx, owner, lesson)
			if err != nil || len(items) != count {
				t.Fatalf("inventory=%+v err=%v want count=%d", items, err, count)
			}
			for i, item := range items {
				if item.Stale != stale || item.UpdatedAt == "" || item.IsPrimary != (item.Language == "uk") {
					t.Fatalf("item=%+v", item)
				}
				if i > 0 && items[i-1].Language >= item.Language {
					t.Fatalf("unordered=%+v", items)
				}
			}
		}

		check(0, false)
		var initialAsset, replacementAsset string
		insertAsset := func(state string, generation int, renditions string) string {
			t.Helper()
			var id string
			if err := pool.QueryRow(ctx, `INSERT INTO video_assets(id,lesson_id,state,generation,original_name,declared_bytes,renditions,expires_at,created_at,updated_at)
    VALUES(gen_random_uuid(),$1,$2,$3,'inventory.mp4',10,$4,now()+interval '1 day',now(),now()) RETURNING id::text`, lesson, state, generation, renditions).Scan(&id); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if _, err := pool.Exec(ctx, `DELETE FROM video_assets WHERE id=$1`, id); err != nil {
					t.Error(err)
				}
			})
			return id
		}
		initialAsset = insertAsset("ready", 1, "{720p}")
		if initialAsset == "" {
			t.Fatal("missing initial Ready asset")
		}
		if _, err := authoring.Save(ctx, owner, application.SaveRequest{LessonID: lesson, Language: "uk"}); err != nil {
			t.Fatal(err)
		}
		check(1, false)
		if _, err := authoring.Save(ctx, owner, application.SaveRequest{LessonID: lesson, Language: "en"}); err != nil {
			t.Fatal(err)
		}
		check(2, false)
		replacementAsset = insertAsset("processing", 2, "{}")
		check(2, false) // a newer upload is not a Ready replacement
		if _, err := pool.Exec(ctx, `UPDATE video_assets SET state='ready',renditions='{720p}' WHERE id=$1`, replacementAsset); err != nil {
			t.Fatal(err)
		}
		check(2, true)
		if _, err := authoring.List(ctx, domain.Actor{UserID: strangerAccount}, lesson); err != domain.ErrForbidden {
			t.Fatalf("foreign course err=%v", err)
		}
	})

	t.Run("M2-R10 inventory without any Ready asset scans the zero ready generation as not stale", func(t *testing.T) {
		var lesson string
		if err := pool.QueryRow(ctx, `INSERT INTO course_lessons(id,module_id,title,position)
   SELECT gen_random_uuid(),module_id,'Inventory no Ready',98 FROM course_lessons WHERE id=$1 RETURNING id::text`, bareLesson).Scan(&lesson); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if _, err := pool.Exec(ctx, `DELETE FROM course_lessons WHERE id=$1`, lesson); err != nil {
				t.Error(err)
			}
		})
		owner := domain.Actor{UserID: ownerAccount}
		if _, err := authoring.Save(ctx, owner, application.SaveRequest{LessonID: lesson, Language: "uk"}); err != nil {
			t.Fatal(err)
		}
		items, err := authoring.List(ctx, owner, lesson)
		if err != nil || len(items) != 1 || items[0].Language != "uk" || items[0].Stale || !items[0].IsPrimary {
			t.Fatalf("inventory=%+v err=%v", items, err)
		}
	})
}
