package domain

type Actor struct {
	UserID        string
	Administrator bool
}

func (actor Actor) CanEditLesson(ownerID string) error {
	if actor.Administrator || actor.UserID == ownerID {
		return nil
	}
	return ErrCourseForbidden
}
