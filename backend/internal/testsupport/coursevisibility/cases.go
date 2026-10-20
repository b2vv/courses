package coursevisibility

const (
	Student = "student"
	Owner   = "owner"
	Admin   = "admin"
)

type Case struct {
	Name   string
	Status string
	Actor  string
	Want   bool
}

// Cases is the nine (role × status) course-visibility rows: the no-payment subset of the shared
// access truth table. It is hand-maintained here rather than derived from accesstest.TruthTable
// because production code (this is a non-test .go file) must not import the access test fixture —
// platform/architecture guards that boundary. cases_consistency_test.go is the machine cross-check
// that keeps this table from drifting out of agreement with TruthTable (backlog риз.9).
func Cases() []Case {
	return []Case{
		{Name: "published student", Status: "published", Actor: Student, Want: true},
		{Name: "published owner", Status: "published", Actor: Owner, Want: true},
		{Name: "published admin", Status: "published", Actor: Admin, Want: true},
		{Name: "draft student", Status: "draft", Actor: Student, Want: false},
		{Name: "draft owner", Status: "draft", Actor: Owner, Want: true},
		{Name: "draft admin", Status: "draft", Actor: Admin, Want: true},
		{Name: "archived student", Status: "archived", Actor: Student, Want: false},
		{Name: "R34 archived owner", Status: "archived", Actor: Owner, Want: true},
		{Name: "archived admin", Status: "archived", Actor: Admin, Want: true},
	}
}
