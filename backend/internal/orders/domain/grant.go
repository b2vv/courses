package domain

import "time"

type GrantSource string

const (
	GrantPurchase    GrantSource = "purchase"
	GrantAdmin       GrantSource = "admin"
	GrantGrandfather GrantSource = "grandfather"
	GrantSelf        GrantSource = "self"
)

type GrantState string

const (
	GrantActive  GrantState = "active"
	GrantBlocked GrantState = "blocked"
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
	knownSource := source == GrantPurchase || source == GrantAdmin || source == GrantGrandfather || source == GrantSelf
	knownState := state == GrantActive || state == GrantBlocked || state == GrantRevoked
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

// This record is authoritative membership; its state is the access authority (enrollment supersedes grant).
func (grant Grant) Block() (Grant, error) {
	if grant.state != GrantActive {
		return Grant{}, ErrGrantTransitionInvalid
	}
	grant.state = GrantBlocked
	return grant, nil
}

func (grant Grant) Unblock() (Grant, error) {
	if grant.state != GrantBlocked {
		return Grant{}, ErrGrantTransitionInvalid
	}
	grant.state = GrantActive
	return grant, nil
}

func (grant Grant) Revoke() (Grant, error) {
	if grant.state != GrantActive && grant.state != GrantBlocked {
		return Grant{}, ErrGrantTransitionInvalid
	}
	grant.state = GrantRevoked
	return grant, nil
}

func SelfEnroll(id, userID, courseID string, grantedAt time.Time) (Grant, error) {
	return NewGrant(id, userID, courseID, GrantSelf, nil, GrantActive, grantedAt, nil)
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
