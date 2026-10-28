package http

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"courses/internal/identity/application"
	"courses/internal/identity/domain"
	"courses/platform/httpserver"
)

const administrationTimeLayout = "2006-01-02T15:04:05.000000Z07:00"

// mutableAccountStatus excludes the delete-only account state.
type mutableAccountStatus string

type adminUserStatusRequest struct {
	Status mutableAccountStatus `json:"status"`
}
type adminUserPasswordRequest struct {
	Password string `json:"password"`
}

type adminUserDTO struct {
	ID          string               `json:"id"`
	Email       string               `json:"email"`
	DisplayName string               `json:"displayName"`
	Role        domain.Role          `json:"role"`
	Status      domain.AccountStatus `json:"status"`
	UpdatedAt   string               `json:"updatedAt"`
}

type adminUserPaginationDTO struct {
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
	Total  int64 `json:"total"`
}

type adminUserPageDTO struct {
	Items      []adminUserDTO         `json:"items"`
	Pagination adminUserPaginationDTO `json:"pagination"`
}

func (handler *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	if _, ok := administrator(w, r); !ok {
		return
	}
	input, err := parseAdministrationQuery(r.URL.RawQuery)
	if err != nil {
		writeError(w, r, err, "invalid_request")
		return
	}
	page, err := handler.administration.ListUsers(r.Context(), input)
	if err != nil {
		writeError(w, r, err, "")
		return
	}
	items := make([]adminUserDTO, 0, len(page.Items))
	for _, user := range page.Items {
		items = append(items, mapAdminUser(user))
	}
	httpserver.WriteJSON(w, http.StatusOK, adminUserPageDTO{
		Items: items,
		Pagination: adminUserPaginationDTO{
			Limit: page.Limit, Offset: page.Offset, Total: page.Total,
		},
	})
}

func (handler *Handler) setUserStatus(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	actor, ok := administrator(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if !validAdministrationID(id) || !emptyAdministrationQuery(r.URL.RawQuery) {
		writeError(w, r, domain.ErrInvalidUser, "invalid_request")
		return
	}
	status, err := decodeAdministrationString(w, r, "status")
	request := adminUserStatusRequest{Status: mutableAccountStatus(status)}
	if err != nil || status != string(domain.AccountActive) && status != string(domain.AccountBlocked) {
		writeError(w, r, domain.ErrInvalidUser, "invalid_request")
		return
	}
	if err := handler.administration.SetStatus(r.Context(), actor, id, domain.AccountStatus(request.Status)); err != nil {
		writeError(w, r, err, "")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (handler *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	actor, ok := administrator(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if !validAdministrationID(id) || !emptyAdministrationQuery(r.URL.RawQuery) {
		writeError(w, r, domain.ErrInvalidUser, "invalid_request")
		return
	}
	if err := handler.administration.Delete(r.Context(), actor, id); err != nil {
		writeError(w, r, err, "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) resetUserPassword(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	if _, ok := administrator(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	if !validAdministrationID(id) || !emptyAdministrationQuery(r.URL.RawQuery) {
		writeError(w, r, domain.ErrInvalidUser, "invalid_request")
		return
	}
	password, err := decodeAdministrationString(w, r, "password")
	request := adminUserPasswordRequest{Password: password}
	if err != nil || !domain.ValidPassword(password) {
		writeError(w, r, domain.ErrInvalidUser, "invalid_request")
		return
	}
	if err := handler.administration.ResetPassword(r.Context(), id, request.Password); err != nil {
		writeError(w, r, err, "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) adminDisableTwoFactor(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	if _, ok := administrator(w, r); !ok {
		return
	}
	if handler.twoFactorUnavailable(w, r) {
		return
	}
	id := r.PathValue("id")
	if !validAdministrationID(id) || !emptyAdministrationQuery(r.URL.RawQuery) {
		writeError(w, r, domain.ErrInvalidUser, "invalid_request")
		return
	}
	actor, ok := r.Context().Value(userKey{}).(domain.User)
	if !ok {
		writeError(w, r, domain.ErrAuthenticationRequired, "")
		return
	}
	if err := handler.twoFactor.AdminDisableTwoFactor(r.Context(), actor, id); err != nil {
		writeError(w, r, err, "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func administrator(w http.ResponseWriter, r *http.Request) (domain.Actor, bool) {
	user, ok := r.Context().Value(userKey{}).(domain.User)
	if !ok {
		writeError(w, r, domain.ErrAuthenticationRequired, "")
		return domain.Actor{}, false
	}
	if !user.Role.CanAdministerAccounts() {
		writeError(w, r, domain.ErrForbidden, "forbidden")
		return domain.Actor{}, false
	}
	return domain.Actor{UserID: user.ID}, true
}

func parseAdministrationQuery(raw string) (application.ListUsersInput, error) {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return application.ListUsersInput{}, domain.ErrInvalidUser
	}
	for key, value := range values {
		if len(value) != 1 {
			return application.ListUsersInput{}, domain.ErrInvalidUser
		}
		switch key {
		case "limit", "offset", "q", "status", "role", "sort":
		default:
			return application.ListUsersInput{}, domain.ErrInvalidUser
		}
	}
	input := application.ListUsersInput{Limit: 20}
	if value, exists := values["limit"]; exists {
		input.Limit, err = administrationDecimal(value[0])
		if err != nil || input.Limit < 1 || input.Limit > 50 {
			return application.ListUsersInput{}, domain.ErrInvalidUser
		}
	}
	if value, exists := values["offset"]; exists {
		input.Offset, err = administrationDecimal(value[0])
		if err != nil || input.Offset > 10000 {
			return application.ListUsersInput{}, domain.ErrInvalidUser
		}
	}
	if value, exists := values["q"]; exists {
		input.Search = strings.TrimSpace(value[0])
		if !utf8.ValidString(input.Search) || strings.ContainsRune(input.Search, 0) || utf8.RuneCountInString(input.Search) > 100 {
			return application.ListUsersInput{}, domain.ErrInvalidUser
		}
	}
	if value, exists := values["status"]; exists {
		input.Status = domain.AccountStatus(value[0])
		if input.Status != domain.AccountActive && input.Status != domain.AccountBlocked && input.Status != domain.AccountDeleted {
			return application.ListUsersInput{}, domain.ErrInvalidUser
		}
	}
	if value, exists := values["role"]; exists {
		input.Role = domain.Role(value[0])
		if input.Role != domain.Student && input.Role != domain.Instructor && input.Role != domain.Admin {
			return application.ListUsersInput{}, domain.ErrInvalidUser
		}
	}
	if value, exists := values["sort"]; exists {
		input.Sort = domain.UserSort(value[0])
		if input.Sort != domain.SortByEmail && input.Sort != domain.SortByName && input.Sort != domain.SortByRecent {
			return application.ListUsersInput{}, domain.ErrInvalidUser
		}
	}
	return input, nil
}

func administrationDecimal(value string) (int, error) {
	if value == "" {
		return 0, domain.ErrInvalidUser
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return 0, domain.ErrInvalidUser
		}
	}
	number, err := strconv.Atoi(value)
	if err != nil {
		return 0, domain.ErrInvalidUser
	}
	return number, nil
}

func decodeAdministrationString(w http.ResponseWriter, r *http.Request, field string) (string, error) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return "", domain.ErrInvalidUser
	}
	r.Body = http.MaxBytesReader(w, r.Body, maximumLoginBody)
	body, err := io.ReadAll(r.Body)
	if err != nil || !utf8.Valid(body) {
		return "", domain.ErrInvalidUser
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('{') {
		return "", domain.ErrInvalidUser
	}
	seen := false
	value := ""
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok || seen || key != field || decoder.Decode(&value) != nil {
			return "", domain.ErrInvalidUser
		}
		seen = true
	}
	closing, err := decoder.Token()
	if err != nil || closing != json.Delim('}') || !seen || decoder.Decode(&struct{}{}) != io.EOF {
		return "", domain.ErrInvalidUser
	}
	return value, nil
}

func validAdministrationID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' || strings.ToLower(value) != value || value == "00000000-0000-0000-0000-000000000000" {
		return false
	}
	decoded, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	return err == nil && len(decoded) == 16
}

func emptyAdministrationQuery(raw string) bool {
	query, err := url.ParseQuery(raw)
	return err == nil && len(query) == 0
}

func mapAdminUser(user domain.User) adminUserDTO {
	return adminUserDTO{
		ID: user.ID, Email: user.Email, DisplayName: user.DisplayName, Role: user.Role,
		Status: user.AccountStatus(), UpdatedAt: user.UpdatedAt.UTC().Truncate(time.Microsecond).Format(administrationTimeLayout),
	}
}
