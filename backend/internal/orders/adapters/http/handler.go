package http

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"courses/internal/orders/application"
	"courses/internal/orders/domain"
	"courses/internal/orders/ports"
	"courses/platform/httpserver"
)

const (
	maximumGrantBody    = 8 * 1024
	maximumCallbackBody = 64 * 1024
	dateTimeLayout      = "2006-01-02T15:04:05.000000Z07:00"
)

type Handler struct {
	service       *application.Service
	callback      ports.CallbackVerifier
	publicBaseURL string
}

func NewHandler(service *application.Service, callback ports.CallbackVerifier, publicBaseURL string) *Handler {
	return &Handler{service: service, callback: callback, publicBaseURL: strings.TrimRight(publicBaseURL, "/")}
}

func (handler *Handler) Register(mux *http.ServeMux, authenticate, sameOrigin func(http.Handler) http.Handler) {
	mux.Handle("POST /api/v1/learn/courses/{courseId}/purchase", sameOrigin(authenticate(http.HandlerFunc(handler.startPurchase))))
	mux.Handle("POST /api/v1/learn/courses/{courseId}/enroll", sameOrigin(authenticate(http.HandlerFunc(handler.enroll))))
	mux.Handle("GET /api/v1/learn/courses/{courseId}/membership", authenticate(http.HandlerFunc(handler.membershipState)))
	mux.Handle("GET /api/v1/learn/memberships", authenticate(http.HandlerFunc(handler.myMemberships)))
	mux.Handle("POST /api/v1/learn/orders/{orderId}/checkout", sameOrigin(authenticate(http.HandlerFunc(handler.checkout))))
	mux.Handle("GET /api/v1/learn/orders/{orderId}", authenticate(http.HandlerFunc(handler.purchaseStatus)))
	if handler.callback != nil {
		mux.Handle("POST /api/v1/payments/wayforpay/callback", http.HandlerFunc(handler.callbackHandler))
	}
}

func (handler *Handler) RegisterAdmin(mux *http.ServeMux, authenticate, sameOrigin, administrator func(http.Handler) http.Handler) {
	mux.Handle("POST /api/v1/admin/courses/{courseId}/grants", sameOrigin(authenticate(administrator(http.HandlerFunc(handler.grantCourse)))))
	mux.Handle("POST /api/v1/admin/courses/{courseId}/participants/{userId}/revoke", sameOrigin(authenticate(http.HandlerFunc(handler.revoke))))
	mux.Handle("GET /api/v1/admin/learners", authenticate(administrator(http.HandlerFunc(handler.adminLearners))))
}

func (handler *Handler) RegisterTeach(mux *http.ServeMux, authenticate, sameOrigin func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/teach/courses/{courseId}/roster", authenticate(http.HandlerFunc(handler.roster)))
	mux.Handle("GET /api/v1/teach/learners", authenticate(http.HandlerFunc(handler.teachLearners)))
	mux.Handle("POST /api/v1/teach/courses/{courseId}/participants/{userId}/block", sameOrigin(authenticate(http.HandlerFunc(handler.block))))
	mux.Handle("POST /api/v1/teach/courses/{courseId}/participants/{userId}/unblock", sameOrigin(authenticate(http.HandlerFunc(handler.unblock))))
	mux.Handle("POST /api/v1/teach/comments/{commentId}/block-author", sameOrigin(authenticate(http.HandlerFunc(handler.blockCommentAuthor))))
	mux.Handle("POST /api/v1/teach/comments/{commentId}/unblock-author", sameOrigin(authenticate(http.HandlerFunc(handler.unblockCommentAuthor))))
	mux.Handle("POST /api/v1/teach/reviews/{reviewId}/block-author", sameOrigin(authenticate(http.HandlerFunc(handler.blockReviewAuthor))))
	mux.Handle("POST /api/v1/teach/reviews/{reviewId}/unblock-author", sameOrigin(authenticate(http.HandlerFunc(handler.unblockReviewAuthor))))
}

type moneyDTO struct {
	Minor    int64  `json:"minor"`
	Currency string `json:"currency"`
}

type orderDTO struct {
	ID        string   `json:"id"`
	CourseID  string   `json:"courseId"`
	Amount    moneyDTO `json:"amount"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"createdAt"`
	PaidAt    *string  `json:"paidAt"`
}

type grantRequest struct {
	UserID string `json:"userId"`
	Reason string `json:"reason"`
}

type checkoutRequest struct {
	CourseSlug string `json:"courseSlug"`
}

type grantDTO struct {
	ID        string  `json:"id"`
	UserID    string  `json:"userId"`
	CourseID  string  `json:"courseId"`
	Source    string  `json:"source"`
	State     string  `json:"state"`
	GrantedAt string  `json:"grantedAt"`
	GrantedBy *string `json:"grantedBy"`
}

type membershipDTO struct {
	State  *string `json:"state"`
	Member bool    `json:"member"`
}

type membershipCardDTO struct {
	CourseID        string `json:"courseId"`
	Slug            string `json:"slug"`
	Title           string `json:"title"`
	State           string `json:"state"`
	ProgressPercent int    `json:"progressPercent"`
	EnrolledAt      string `json:"enrolledAt"`
}

type myMembershipsPageDTO struct {
	Items []membershipCardDTO `json:"items"`
}

type rosterEntryDTO struct {
	UserID          string `json:"userId"`
	DisplayName     string `json:"displayName"`
	State           string `json:"state"`
	EnrolledAt      string `json:"enrolledAt"`
	ProgressPercent int    `json:"progressPercent"`
}

type rosterPageDTO struct {
	Entries    []rosterEntryDTO `json:"entries"`
	NextCursor *string          `json:"nextCursor"`
}

type learnerCourseDTO struct {
	CourseID        string `json:"courseId"`
	Slug            string `json:"slug"`
	Title           string `json:"title"`
	State           string `json:"state"`
	ProgressPercent int    `json:"progressPercent"`
	EnrolledAt      string `json:"enrolledAt"`
}

type learnerRowDTO struct {
	UserID         string             `json:"userId"`
	DisplayName    string             `json:"displayName"`
	LastActivityAt *string            `json:"lastActivityAt"`
	Courses        []learnerCourseDTO `json:"courses"`
}

type learnerPageDTO struct {
	Entries    []learnerRowDTO `json:"entries"`
	NextCursor *string         `json:"nextCursor"`
}

func (handler *Handler) startPurchase(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	if !emptyQuery(r.URL.RawQuery) {
		writeError(w, r, domain.ErrOrderInvalid)
		return
	}
	if !validID(r.PathValue("courseId")) {
		writeError(w, r, domain.ErrOrderInvalid)
		return
	}
	keys := r.Header.Values("Idempotency-Key")
	if len(keys) != 1 {
		writeError(w, r, domain.ErrOrderInvalid)
		return
	}
	key := keys[0]
	if len(key) < 1 || len(key) > 128 || strings.TrimSpace(key) != key || !utf8.ValidString(key) {
		writeError(w, r, domain.ErrOrderInvalid)
		return
	}
	order, err := handler.service.StartPurchase(r.Context(), actor, r.PathValue("courseId"), key)
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, mapOrder(order))
}

func (handler *Handler) enroll(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	if !emptyQuery(r.URL.RawQuery) || !validID(r.PathValue("courseId")) {
		writeError(w, r, domain.ErrGrantInvalid)
		return
	}
	grant, err := handler.service.Enroll(r.Context(), actor, r.PathValue("courseId"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, mapGrant(grant))
}

func (handler *Handler) membershipState(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	if !emptyQuery(r.URL.RawQuery) || !validID(r.PathValue("courseId")) {
		writeError(w, r, domain.ErrGrantInvalid)
		return
	}
	state, member, err := handler.service.MembershipState(r.Context(), actor, actor.UserID, r.PathValue("courseId"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	response := membershipDTO{Member: member}
	if member {
		value := string(state)
		response.State = &value
	}
	httpserver.WriteJSON(w, http.StatusOK, response)
}

func (handler *Handler) myMemberships(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	if !emptyQuery(r.URL.RawQuery) {
		writeError(w, r, domain.ErrGrantInvalid)
		return
	}
	cards, err := handler.service.MyMemberships(r.Context(), actor)
	if err != nil {
		writeError(w, r, err)
		return
	}
	items := make([]membershipCardDTO, 0, len(cards))
	for _, card := range cards {
		items = append(items, membershipCardDTO{CourseID: card.CourseID, Slug: card.Slug, Title: card.Title, State: string(card.State), ProgressPercent: card.ProgressPercent, EnrolledAt: card.EnrolledAt.UTC().Format(dateTimeLayout)})
	}
	httpserver.WriteJSON(w, http.StatusOK, myMembershipsPageDTO{Items: items})
}

func (handler *Handler) roster(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	if !validID(r.PathValue("courseId")) {
		writeError(w, r, domain.ErrGrantInvalid)
		return
	}
	query := r.URL.Query()
	if len(query["cursor"]) > 1 || len(query["limit"]) > 1 || len(query) > lenPresent(query, "cursor", "limit") {
		writeError(w, r, domain.ErrInvalidCursor)
		return
	}
	limit := 0
	if values := query["limit"]; len(values) == 1 {
		var err error
		limit, err = strconv.Atoi(values[0])
		if err != nil || limit < 1 || limit > 100 {
			writeError(w, r, domain.ErrInvalidCursor)
			return
		}
	}
	cursor := ""
	if values := query["cursor"]; len(values) == 1 {
		cursor = values[0]
	}
	page, err := handler.service.Roster(r.Context(), actor, r.PathValue("courseId"), cursor, limit)
	if err != nil {
		writeError(w, r, err)
		return
	}
	entries := make([]rosterEntryDTO, 0, len(page.Entries))
	for _, entry := range page.Entries {
		entries = append(entries, rosterEntryDTO{UserID: entry.UserID, DisplayName: entry.DisplayName, State: string(entry.State), EnrolledAt: entry.EnrolledAt.UTC().Format(dateTimeLayout), ProgressPercent: entry.ProgressPercent})
	}
	response := rosterPageDTO{Entries: entries}
	if page.NextCursor != "" {
		response.NextCursor = &page.NextCursor
	}
	httpserver.WriteJSON(w, http.StatusOK, response)
}

func lenPresent(query url.Values, keys ...string) int {
	n := 0
	for _, key := range keys {
		if _, ok := query[key]; ok {
			n++
		}
	}
	return n
}

func (handler *Handler) teachLearners(w http.ResponseWriter, r *http.Request) {
	handler.learners(w, r, false)
}

func (handler *Handler) adminLearners(w http.ResponseWriter, r *http.Request) {
	handler.learners(w, r, true)
}

func (handler *Handler) learners(w http.ResponseWriter, r *http.Request, adminScope bool) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	query := r.URL.Query()
	for key, values := range query {
		switch key {
		case "courseId", "state", "q", "cursor", "limit":
			if len(values) != 1 {
				writeError(w, r, domain.ErrInvalidCursor)
				return
			}
		default:
			writeError(w, r, domain.ErrInvalidCursor)
			return
		}
	}
	courseID := query.Get("courseId")
	if query.Has("courseId") && !validID(courseID) {
		writeError(w, r, domain.ErrInvalidCursor)
		return
	}
	state := query.Get("state")
	if query.Has("state") && state != "active" && state != "blocked" {
		writeError(w, r, domain.ErrInvalidCursor)
		return
	}
	q := query.Get("q")
	if n := utf8.RuneCountInString(q); query.Has("q") && (!utf8.ValidString(q) || n < 1 || n > 80) {
		writeError(w, r, domain.ErrInvalidCursor)
		return
	}
	limit := 0
	if values := query["limit"]; len(values) == 1 {
		parsed, err := strconv.Atoi(values[0])
		if err != nil || parsed < 1 || parsed > 100 {
			writeError(w, r, domain.ErrInvalidCursor)
			return
		}
		limit = parsed
	}
	page, err := handler.service.Learners(r.Context(), actor, adminScope, courseID, state, q, query.Get("cursor"), limit)
	if err != nil {
		writeError(w, r, err)
		return
	}
	entries := make([]learnerRowDTO, 0, len(page.Entries))
	for _, row := range page.Entries {
		courses := make([]learnerCourseDTO, 0, len(row.Courses))
		for _, course := range row.Courses {
			courses = append(courses, learnerCourseDTO{CourseID: course.CourseID, Slug: course.Slug, Title: course.Title, State: string(course.State), ProgressPercent: course.ProgressPercent, EnrolledAt: course.EnrolledAt.UTC().Format(dateTimeLayout)})
		}
		dto := learnerRowDTO{UserID: row.UserID, DisplayName: row.DisplayName, Courses: courses}
		if row.LastActivityAt != nil {
			stamp := row.LastActivityAt.UTC().Format(dateTimeLayout)
			dto.LastActivityAt = &stamp
		}
		entries = append(entries, dto)
	}
	response := learnerPageDTO{Entries: entries}
	if page.NextCursor != "" {
		response.NextCursor = &page.NextCursor
	}
	httpserver.WriteJSON(w, http.StatusOK, response)
}

func (handler *Handler) block(w http.ResponseWriter, r *http.Request) {
	handler.transition(w, r, handler.service.Block)
}

func (handler *Handler) unblock(w http.ResponseWriter, r *http.Request) {
	handler.transition(w, r, handler.service.Unblock)
}

func (handler *Handler) revoke(w http.ResponseWriter, r *http.Request) {
	handler.transition(w, r, handler.service.Revoke)
}

func (handler *Handler) blockCommentAuthor(w http.ResponseWriter, r *http.Request) {
	handler.authorTransition(w, r, "comment", "commentId", handler.service.BlockAuthor)
}

func (handler *Handler) unblockCommentAuthor(w http.ResponseWriter, r *http.Request) {
	handler.authorTransition(w, r, "comment", "commentId", handler.service.UnblockAuthor)
}

func (handler *Handler) blockReviewAuthor(w http.ResponseWriter, r *http.Request) {
	handler.authorTransition(w, r, "review", "reviewId", handler.service.BlockAuthor)
}

func (handler *Handler) unblockReviewAuthor(w http.ResponseWriter, r *http.Request) {
	handler.authorTransition(w, r, "review", "reviewId", handler.service.UnblockAuthor)
}

func (handler *Handler) authorTransition(w http.ResponseWriter, r *http.Request, kind, pathParam string, operation func(context.Context, application.Actor, string, string) (domain.Grant, error)) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	id := r.PathValue(pathParam)
	if !emptyQuery(r.URL.RawQuery) || !validID(id) {
		writeError(w, r, domain.ErrGrantInvalid)
		return
	}
	grant, err := operation(r.Context(), actor, kind, id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, mapGrant(grant))
}

func (handler *Handler) transition(w http.ResponseWriter, r *http.Request, operation func(context.Context, application.Actor, string, string) (domain.Grant, error)) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	if !emptyQuery(r.URL.RawQuery) || !validID(r.PathValue("courseId")) || !validID(r.PathValue("userId")) {
		writeError(w, r, domain.ErrGrantInvalid)
		return
	}
	grant, err := operation(r.Context(), actor, r.PathValue("userId"), r.PathValue("courseId"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, mapGrant(grant))
}

func (handler *Handler) purchaseStatus(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	if !emptyQuery(r.URL.RawQuery) {
		writeError(w, r, domain.ErrOrderInvalid)
		return
	}
	if !validID(r.PathValue("orderId")) {
		writeError(w, r, domain.ErrOrderInvalid)
		return
	}
	order, err := handler.service.PurchaseStatus(r.Context(), actor, r.PathValue("orderId"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, mapOrder(order))
}

func (handler *Handler) checkout(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	if !emptyQuery(r.URL.RawQuery) || !validID(r.PathValue("orderId")) {
		writeError(w, r, domain.ErrOrderInvalid)
		return
	}
	request, err := decodeCheckout(r)
	if err != nil {
		writeError(w, r, domain.ErrOrderInvalid)
		return
	}
	base := handler.publicBaseURL
	returnURL := base + "/orders/" + r.PathValue("orderId")
	if validCourseSlug(request.CourseSlug) {
		returnURL += "?course=" + url.QueryEscape(request.CourseSlug)
	}
	urls := ports.CheckoutURLs{ReturnURL: returnURL, ServiceURL: base + "/api/v1/payments/wayforpay/callback"}
	form, err := handler.service.Checkout(r.Context(), actor, r.PathValue("orderId"), ports.Buyer{ID: actor.UserID}, urls)
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, form)
}

func decodeCheckout(r *http.Request) (checkoutRequest, error) {
	if r.Body == nil {
		return checkoutRequest{}, nil
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maximumGrantBody+1))
	if err != nil || len(body) > maximumGrantBody || strings.TrimSpace(string(body)) == "" {
		return checkoutRequest{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if validateUniqueJSONObject(body) != nil {
		return checkoutRequest{}, domain.ErrOrderInvalid
	}
	var request checkoutRequest
	if err := decoder.Decode(&request); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return checkoutRequest{}, domain.ErrOrderInvalid
	}
	return request, nil
}

func (handler *Handler) callbackHandler(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	r.Body = http.MaxBytesReader(w, r.Body, maximumCallbackBody)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		httpserver.WriteError(w, r, http.StatusUnauthorized, "invalid_callback", "Invalid payment callback")
		return
	}
	now := time.Now()
	callback, err := handler.callback.VerifyCallback(body, now)
	if err != nil {
		httpserver.WriteError(w, r, http.StatusUnauthorized, "invalid_callback", "Invalid payment callback")
		return
	}
	if callback.Event != nil {
		err = handler.service.ApplyPaymentEvent(r.Context(), *callback.Event)
		if errors.Is(err, domain.ErrPaymentEventUnmatched) {
			// WayForPay retries when the service callback does not receive a valid
			// acknowledgement. An unmatched order is transient: accepting here would
			// permanently lose a callback delivered before the order commit.
			writeError(w, r, domain.ErrDependencyUnavailable)
			return
		}
		if err != nil && !errors.Is(err, domain.ErrPaymentEventRejected) {
			writeError(w, r, err)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(handler.callback.Acknowledge(callback.OrderReference, now))
}

func (handler *Handler) grantCourse(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	actor, ok := actorFromRequest(w, r)
	if !ok {
		return
	}
	if !emptyQuery(r.URL.RawQuery) {
		writeError(w, r, domain.ErrGrantInvalid)
		return
	}
	if !validID(r.PathValue("courseId")) {
		writeError(w, r, domain.ErrGrantInvalid)
		return
	}
	request, err := decodeGrant(w, r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if !validID(request.UserID) {
		writeError(w, r, domain.ErrGrantInvalid)
		return
	}
	grant, err := handler.service.GrantCourse(r.Context(), actor, request.UserID, r.PathValue("courseId"), request.Reason)
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, mapGrant(grant))
}

func mapGrant(grant domain.Grant) grantDTO {
	return grantDTO{ID: grant.ID(), UserID: grant.UserID(), CourseID: grant.CourseID(), Source: string(grant.Source()), State: string(grant.State()), GrantedAt: grant.GrantedAt().UTC().Format(dateTimeLayout), GrantedBy: grant.GrantedBy()}
}

func actorFromRequest(w http.ResponseWriter, r *http.Request) (application.Actor, bool) {
	actor, ok := httpserver.ActorFromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, http.StatusUnauthorized, "authentication_required", "Authentication is required")
		return application.Actor{}, false
	}
	return application.Actor{UserID: actor.UserID, Administrator: actor.Role == "admin", EmailVerified: actor.EmailVerified}, true
}

func mapOrder(order domain.Order) orderDTO {
	var paidAt *string
	if value := order.PaidAt(); value != nil {
		formatted := value.UTC().Format(dateTimeLayout)
		paidAt = &formatted
	}
	return orderDTO{ID: order.ID(), CourseID: order.CourseID(), Amount: moneyDTO{Minor: int64(order.Amount().MinorUnits()), Currency: string(order.Amount().Currency())}, Status: string(order.Status()), CreatedAt: order.CreatedAt().UTC().Format(dateTimeLayout), PaidAt: paidAt}
}

func decodeGrant(w http.ResponseWriter, r *http.Request) (grantRequest, error) {
	if strings.TrimSpace(strings.Split(r.Header.Get("Content-Type"), ";")[0]) != "application/json" {
		return grantRequest{}, domain.ErrGrantInvalid
	}
	r.Body = http.MaxBytesReader(w, r.Body, maximumGrantBody)
	body, err := io.ReadAll(r.Body)
	if err != nil || validateUniqueJSONObject(body) != nil {
		return grantRequest{}, domain.ErrGrantInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var request grantRequest
	if err := decoder.Decode(&request); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return grantRequest{}, domain.ErrGrantInvalid
	}
	return request, nil
}

func validateUniqueJSONObject(body []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('{') {
		return domain.ErrGrantInvalid
	}
	if err := validateUniqueObject(decoder); err != nil {
		return err
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return domain.ErrGrantInvalid
	}
	return nil
}

func validateUniqueObject(decoder *json.Decoder) error {
	seen := make(map[string]struct{})
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return domain.ErrGrantInvalid
		}
		if _, exists := seen[key]; exists {
			return domain.ErrGrantInvalid
		}
		seen[key] = struct{}{}
		if err := validateUniqueValue(decoder); err != nil {
			return err
		}
	}
	closing, err := decoder.Token()
	if err != nil || closing != json.Delim('}') {
		return domain.ErrGrantInvalid
	}
	return nil
}

func validateUniqueValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return domain.ErrGrantInvalid
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		return validateUniqueObject(decoder)
	case '[':
		for decoder.More() {
			if err := validateUniqueValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim(']') {
			return domain.ErrGrantInvalid
		}
		return nil
	default:
		return domain.ErrGrantInvalid
	}
}

func emptyQuery(raw string) bool {
	query, err := url.ParseQuery(raw)
	return err == nil && len(query) == 0
}

func validID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' || strings.ToLower(value) != value || value == "00000000-0000-0000-0000-000000000000" {
		return false
	}
	decoded, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	return err == nil && len(decoded) == 16
}

func validCourseSlug(value string) bool {
	if len(value) == 0 || len(value) > 120 || value[0] == '-' || value[len(value)-1] == '-' {
		return false
	}
	previousHyphen := false
	for _, char := range value {
		if char == '-' {
			if previousHyphen {
				return false
			}
			previousHyphen = true
			continue
		}
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') {
			return false
		}
		previousHyphen = false
	}
	return true
}

func noStore(w http.ResponseWriter) { w.Header().Set("Cache-Control", "no-store") }

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrEmailNotVerified):
		httpserver.WriteError(w, r, http.StatusForbidden, "email_not_verified", "Email verification is required")
	case errors.Is(err, domain.ErrOrderInvalid), errors.Is(err, domain.ErrGrantInvalid), errors.Is(err, domain.ErrMoneyInvalid), errors.Is(err, domain.ErrCourseNotPurchasable):
		httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Request is invalid")
	case errors.Is(err, domain.ErrCourseAlreadyOwned):
		httpserver.WriteError(w, r, http.StatusConflict, "course_already_owned", "Course is already owned")
	case errors.Is(err, domain.ErrMembershipConflict):
		httpserver.WriteError(w, r, http.StatusConflict, "membership_conflict", "Membership changed concurrently")
	case errors.Is(err, domain.ErrGrantTransitionInvalid):
		httpserver.WriteError(w, r, http.StatusConflict, "grant_transition_invalid", "Membership cannot make that transition")
	case errors.Is(err, domain.ErrCourseNotEnrollable):
		httpserver.WriteError(w, r, http.StatusConflict, "course_not_enrollable", "Course is not available for enrollment")
	case errors.Is(err, domain.ErrEnrollmentForbidden):
		httpserver.WriteError(w, r, http.StatusForbidden, "enrollment_forbidden", "Enrollment permission is required")
	case errors.Is(err, domain.ErrLearnerCourseForbidden):
		httpserver.WriteError(w, r, http.StatusForbidden, "learner_course_forbidden", "The overview cannot be filtered by another instructor's course")
	case errors.Is(err, domain.ErrInvalidCursor):
		httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Request is invalid")
	case errors.Is(err, domain.ErrOrderNotFound):
		httpserver.WriteError(w, r, http.StatusNotFound, "order_not_found", "Order was not found")
	case errors.Is(err, domain.ErrCommentNotFound):
		httpserver.WriteError(w, r, http.StatusNotFound, "comment_not_found", "Comment was not found")
	case errors.Is(err, domain.ErrReviewNotFound):
		httpserver.WriteError(w, r, http.StatusNotFound, "review_not_found", "Review was not found")
	case errors.Is(err, domain.ErrCheckoutAlreadyOpened):
		httpserver.WriteError(w, r, http.StatusConflict, "checkout_already_opened", "Checkout was already opened")
	case errors.Is(err, domain.ErrGrantForbidden):
		httpserver.WriteError(w, r, http.StatusForbidden, "grant_forbidden", "Administrator permission is required")
	case errors.Is(err, domain.ErrGrantSelfForbidden):
		httpserver.WriteError(w, r, http.StatusForbidden, "grant_self_forbidden", "Administrators cannot grant themselves access")
	case errors.Is(err, domain.ErrAdminNotLearner):
		httpserver.WriteError(w, r, http.StatusForbidden, "admin_not_learner", "Administrators are not learners and cannot enrol, purchase, or be granted a course")
	case errors.Is(err, domain.ErrPaymentEventRejected), errors.Is(err, domain.ErrPaymentEventUnmatched):
		httpserver.WriteError(w, r, http.StatusConflict, "payment_event_rejected", "Payment event was rejected")
	case errors.Is(err, domain.ErrCourseModified):
		httpserver.WriteError(w, r, http.StatusConflict, "course_modified", "Course changed since it was loaded")
	case errors.Is(err, domain.ErrDependencyUnavailable):
		httpserver.WriteError(w, r, http.StatusServiceUnavailable, "dependency_unavailable", "Purchases are temporarily unavailable")
	default:
		httpserver.WriteError(w, r, http.StatusInternalServerError, "internal_error", "Unexpected server error")
	}
}
