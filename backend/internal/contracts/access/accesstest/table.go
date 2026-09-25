// Package accesstest contains the shared access truth table used by module tests.
package accesstest

// Role identifies the neutral relationship between a subject and a course.
type Role string

const (
	Administrator Role = "administrator"
	Owner         Role = "owner"
	Stranger      Role = "stranger"
)

// Case is one access-rule truth-table row.
type Case struct {
	Name   string
	Role   Role
	Status string
	Want   bool
}

// TruthTable is the nine combinations of subject role and course status.
var TruthTable = [...]Case{
	{Name: "administrator draft", Role: Administrator, Status: "draft", Want: true},
	{Name: "administrator published", Role: Administrator, Status: "published", Want: true},
	{Name: "administrator archived", Role: Administrator, Status: "archived", Want: true},
	{Name: "owner draft", Role: Owner, Status: "draft", Want: true},
	{Name: "owner published", Role: Owner, Status: "published", Want: true},
	{Name: "owner archived", Role: Owner, Status: "archived", Want: true},
	{Name: "stranger draft", Role: Stranger, Status: "draft", Want: false},
	{Name: "stranger published", Role: Stranger, Status: "published", Want: true},
	{Name: "stranger archived", Role: Stranger, Status: "archived", Want: false},
}
