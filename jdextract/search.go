package jdextract

import (
	"net/http"
	"strconv"
	"strings"
)

// matchesQuery reports whether q appears (case-insensitive) in any of fields.
// An empty q always returns true.
func matchesQuery(q string, fields ...string) bool {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return true
	}
	for _, f := range fields {
		if strings.Contains(strings.ToLower(f), q) {
			return true
		}
	}
	return false
}

// applyJobFilters filters jobs by the query parameters present in r.
// All parameters are optional and AND-combined.
//
// Supported parameters:
//   - q:          substring match on Company + Role
//   - status:     exact match on Status
//   - score_min:  inclusive lower bound on Score
//   - score_max:  inclusive upper bound on Score
//   - date_from:  YYYY-MM-DD — include jobs on or after this date
//   - date_to:    YYYY-MM-DD — include jobs on or before this date
func applyJobFilters(jobs []ApplicationMeta, r *http.Request) []ApplicationMeta {
	q := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	scoreMinStr := r.URL.Query().Get("score_min")
	scoreMaxStr := r.URL.Query().Get("score_max")
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	scoreMin, hasScoreMin := 0, false
	if scoreMinStr != "" {
		if v, err := strconv.Atoi(scoreMinStr); err == nil {
			scoreMin, hasScoreMin = v, true
		}
	}
	scoreMax, hasScoreMax := 0, false
	if scoreMaxStr != "" {
		if v, err := strconv.Atoi(scoreMaxStr); err == nil {
			scoreMax, hasScoreMax = v, true
		}
	}

	var out []ApplicationMeta
	for _, j := range jobs {
		if !matchesQuery(q, j.Company, j.Role) {
			continue
		}
		if status != "" && j.Status != status {
			continue
		}
		if hasScoreMin && j.Score < scoreMin {
			continue
		}
		if hasScoreMax && j.Score > scoreMax {
			continue
		}
		if dateFrom != "" && j.Date < dateFrom {
			continue
		}
		if dateTo != "" && j.Date > dateTo {
			continue
		}
		out = append(out, j)
	}
	return out
}

// applyContactFilters filters contacts by the query parameters present in r.
// All parameters are optional and AND-combined.
//
// Supported parameters:
//   - q:               substring match on Name, Company, Role, Email, Notes, and all conversation message content
//   - status:          exact match on Status
//   - tag:             contact must include this tag (case-insensitive)
//   - followup_before: YYYY-MM-DD — contacts with follow-up date on or before this date
//   - followup_after:  YYYY-MM-DD — contacts with follow-up date on or after this date
//   - has_followup:    "true" — only contacts where FollowUpDate is set
func applyContactFilters(contacts []ContactMeta, r *http.Request) []ContactMeta {
	q := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	tag := r.URL.Query().Get("tag")
	followupBefore := r.URL.Query().Get("followup_before")
	followupAfter := r.URL.Query().Get("followup_after")
	hasFollowup := r.URL.Query().Get("has_followup") == "true"

	var out []ContactMeta
	for _, c := range contacts {
		if q != "" {
			var msgContent strings.Builder
			for _, conv := range c.Conversations {
				for _, msg := range conv.Messages {
					msgContent.WriteByte(' ')
					msgContent.WriteString(msg.Content)
				}
			}
			if !matchesQuery(q, c.Name, c.Company, c.Role, c.Email, c.Notes, msgContent.String()) {
				continue
			}
		}
		if status != "" && c.Status != status {
			continue
		}
		if tag != "" {
			found := false
			for _, t := range c.Tags {
				if strings.EqualFold(t, tag) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if followupBefore != "" && (c.FollowUpDate == "" || c.FollowUpDate > followupBefore) {
			continue
		}
		if followupAfter != "" && (c.FollowUpDate == "" || c.FollowUpDate < followupAfter) {
			continue
		}
		if hasFollowup && c.FollowUpDate == "" {
			continue
		}
		out = append(out, c)
	}
	return out
}
