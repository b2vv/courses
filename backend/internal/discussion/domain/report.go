package domain

type Report struct {
	ID         string
	CommentID  string
	ReporterID string
}

func NewReport(input Report, comment Comment) (Report, error) {
	if input.ReporterID == comment.AuthorID {
		return Report{}, ErrReportOwnComment
	}
	return input, nil
}
