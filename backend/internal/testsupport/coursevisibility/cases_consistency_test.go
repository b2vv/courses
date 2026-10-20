package coursevisibility_test

import (
	"testing"

	"courses/internal/contracts/access/accesstest"
	"courses/internal/testsupport/coursevisibility"
)

// The nine (role × status) rows in coursevisibility.Cases() are the no-payment subset of the shared
// access truth table. The table is hand-maintained (production code must not import the access test
// fixture — platform/architecture guards that), so this test is the machine check that keeps the two
// from drifting: a change to the access rule that updated only one table is the C10-class bug nobody
// sees until it surfaces in the UI (CTO-RESEARCH §6 риз.9). This file is a _test.go, so importing the
// fixture here is allowed — the same way the module access tests already do.
func TestCasesMatchSharedTruthTableVisibilitySubset(t *testing.T) {
	actorFor := map[accesstest.Role]string{
		accesstest.Stranger:      coursevisibility.Student,
		accesstest.Owner:         coursevisibility.Owner,
		accesstest.Administrator: coursevisibility.Admin,
	}

	type key struct{ status, actor string }

	want := make(map[key]bool)
	for _, row := range accesstest.TruthTable {
		if row.Paid || row.Granted || row.Preview {
			continue
		}
		want[key{row.Status, actorFor[row.Role]}] = row.Want
	}

	got := make(map[key]bool)
	for _, c := range coursevisibility.Cases() {
		got[key{c.Status, c.Actor}] = c.Want
	}

	if len(got) != len(want) {
		t.Fatalf("Cases() covers %d (status,actor) pairs; truth-table visibility subset covers %d", len(got), len(want))
	}
	for k, w := range want {
		g, ok := got[k]
		if !ok {
			t.Fatalf("Cases() is missing status=%s actor=%s (truth table wants %t)", k.status, k.actor, w)
		}
		if g != w {
			t.Fatalf("status=%s actor=%s: Cases() want=%t, truth table want=%t", k.status, k.actor, g, w)
		}
	}
}
