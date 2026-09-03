package domain

import "time"

type Session struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func NewSession(session Session) (Session, error) {
	if !validID(session.ID) || !validID(session.UserID) || session.CreatedAt.IsZero() || !session.ExpiresAt.After(session.CreatedAt) {
		return Session{}, ErrInvalidSession
	}
	return session, nil
}

func (session Session) IsActiveAt(now time.Time) bool {
	return !now.Before(session.CreatedAt) && now.Before(session.ExpiresAt)
}
