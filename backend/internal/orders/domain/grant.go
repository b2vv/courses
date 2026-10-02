package domain

import "time"

type GrantSource string

const (
	GrantPurchase    GrantSource = "purchase"
	GrantAdmin       GrantSource = "admin"
	GrantGrandfather GrantSource = "grandfather"
)

type GrantState string

const (
	GrantActive  GrantState = "active"
	GrantRevoked GrantState = "revoked"
)

type Grant struct {
	id        string
	userID    string
	courseID  string
	source    GrantSource
	orderID   *string
	state     GrantState
	grantedAt time.Time
	grantedBy *string
}

func NewGrant(
	id, userID, courseID string,
	source GrantSource,
	orderID *string,
	state GrantState,
	grantedAt time.Time,
	grantedBy *string,
) (Grant, error) {
	knownSource := source == GrantPurchase || source == GrantAdmin || source == GrantGrandfather
	knownState := state == GrantActive || state == GrantRevoked
	hasOrder := orderID != nil
	if id == "" || userID == "" || courseID == "" || grantedAt.IsZero() || !knownSource || !knownState ||
		(source == GrantPurchase) != hasOrder || hasOrder && *orderID == "" {
		return Grant{}, ErrGrantInvalid
	}
	return Grant{
		id: id, userID: userID, courseID: courseID, source: source, orderID: copyString(orderID),
		state: state, grantedAt: grantedAt, grantedBy: copyString(grantedBy),
	}, nil
}

func (grant Grant) ID() string { return grant.id }

func (grant Grant) UserID() string { return grant.userID }

func (grant Grant) CourseID() string { return grant.courseID }

func (grant Grant) Source() GrantSource { return grant.source }

func (grant Grant) OrderID() *string { return copyString(grant.orderID) }

func (grant Grant) State() GrantState { return grant.state }

func (grant Grant) GrantedAt() time.Time { return grant.grantedAt }

func (grant Grant) GrantedBy() *string { return copyString(grant.grantedBy) }

func CanStartOrder(userID, courseID string, grants []Grant) error {
	for _, grant := range grants {
		if grant.userID == userID && grant.courseID == courseID && grant.state == GrantActive {
			return ErrCourseAlreadyOwned
		}
	}
	return nil
}

func copyString(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
