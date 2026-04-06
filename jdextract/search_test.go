package jdextract

import (
	"net/http"
	"net/url"
	"testing"
)

func TestMatchesQuery(t *testing.T) {
	tests := []struct {
		q      string
		fields []string
		want   bool
	}{
		{"", []string{"anything"}, true},
		{"  ", []string{"anything"}, true},
		{"acme", []string{"Acme Corp"}, true},
		{"ACME", []string{"acme corp"}, true},
		{"copywriter", []string{"Senior Copywriter"}, true},
		{"xyz", []string{"Acme Corp", "Copywriter"}, false},
		{"corp", []string{"Acme Corp", "Copywriter"}, true},
		{"senior copy", []string{"Senior Copywriter"}, true},
		{"acme", []string{}, false},
	}
	for _, tt := range tests {
		got := matchesQuery(tt.q, tt.fields...)
		if got != tt.want {
			t.Errorf("matchesQuery(%q, %v) = %v, want %v", tt.q, tt.fields, got, tt.want)
		}
	}
}

func makeRequest(params map[string]string) *http.Request {
	v := url.Values{}
	for k, val := range params {
		v.Set(k, val)
	}
	return &http.Request{URL: &url.URL{RawQuery: v.Encode()}}
}

func TestApplyJobFilters(t *testing.T) {
	jobs := []ApplicationMeta{
		{Company: "Acme Corp", Role: "Senior Copywriter", Score: 8, Status: "applied", Date: "2024-03-01"},
		{Company: "Acme Corp", Role: "Intermediate Copywriter", Score: 5, Status: "draft", Date: "2024-02-15"},
		{Company: "Felix Inc", Role: "Content Strategist", Score: 7, Status: "applied", Date: "2024-01-20"},
	}

	t.Run("no filters returns all", func(t *testing.T) {
		got := applyJobFilters(jobs, makeRequest(nil))
		if len(got) != 3 {
			t.Errorf("got %d, want 3", len(got))
		}
	})

	t.Run("q filters by company+role", func(t *testing.T) {
		got := applyJobFilters(jobs, makeRequest(map[string]string{"q": "acme"}))
		if len(got) != 2 {
			t.Errorf("got %d, want 2", len(got))
		}
	})

	t.Run("status filter", func(t *testing.T) {
		got := applyJobFilters(jobs, makeRequest(map[string]string{"status": "applied"}))
		if len(got) != 2 {
			t.Errorf("got %d, want 2", len(got))
		}
	})

	t.Run("score_min filter", func(t *testing.T) {
		got := applyJobFilters(jobs, makeRequest(map[string]string{"score_min": "7"}))
		if len(got) != 2 {
			t.Errorf("got %d, want 2", len(got))
		}
	})

	t.Run("score_max filter", func(t *testing.T) {
		got := applyJobFilters(jobs, makeRequest(map[string]string{"score_max": "6"}))
		if len(got) != 1 {
			t.Errorf("got %d, want 1", len(got))
		}
	})

	t.Run("date_from filter", func(t *testing.T) {
		got := applyJobFilters(jobs, makeRequest(map[string]string{"date_from": "2024-02-01"}))
		if len(got) != 2 {
			t.Errorf("got %d, want 2", len(got))
		}
	})

	t.Run("date_to filter", func(t *testing.T) {
		got := applyJobFilters(jobs, makeRequest(map[string]string{"date_to": "2024-02-28"}))
		if len(got) != 2 {
			t.Errorf("got %d, want 2", len(got))
		}
	})

	t.Run("combined filters", func(t *testing.T) {
		got := applyJobFilters(jobs, makeRequest(map[string]string{"q": "acme", "status": "applied"}))
		if len(got) != 1 {
			t.Errorf("got %d, want 1", len(got))
		}
		if got[0].Role != "Senior Copywriter" {
			t.Errorf("unexpected role: %s", got[0].Role)
		}
	})
}

func TestApplyContactFilters(t *testing.T) {
	contacts := []ContactMeta{
		{
			Name: "Alex Kim", Company: "Acme Corp", Role: "Recruiter",
			Status: "reached-out", Tags: []string{"recruiter", "warm"},
			FollowUpDate: "2024-04-01",
		},
		{
			Name: "Sam Lee", Company: "Felix Inc", Role: "Engineer",
			Status: "new", Tags: []string{"engineer"},
			FollowUpDate: "",
			Conversations: []Conversation{
				{Messages: []Message{{Sender: "me", Content: "discussed compensation package", Date: "2024-03-10"}}},
			},
		},
		{
			Name: "Jordan Park", Company: "Acme Corp", Role: "Manager",
			Status: "connected", Tags: []string{"warm", "manager"},
			FollowUpDate: "2024-05-15",
		},
	}

	t.Run("no filters returns all", func(t *testing.T) {
		got := applyContactFilters(contacts, makeRequest(nil))
		if len(got) != 3 {
			t.Errorf("got %d, want 3", len(got))
		}
	})

	t.Run("q matches name", func(t *testing.T) {
		got := applyContactFilters(contacts, makeRequest(map[string]string{"q": "alex"}))
		if len(got) != 1 || got[0].Name != "Alex Kim" {
			t.Errorf("unexpected result: %v", got)
		}
	})

	t.Run("q matches company", func(t *testing.T) {
		got := applyContactFilters(contacts, makeRequest(map[string]string{"q": "acme"}))
		if len(got) != 2 {
			t.Errorf("got %d, want 2", len(got))
		}
	})

	t.Run("q matches conversation content", func(t *testing.T) {
		got := applyContactFilters(contacts, makeRequest(map[string]string{"q": "compensation"}))
		if len(got) != 1 || got[0].Name != "Sam Lee" {
			t.Errorf("unexpected result: %v", got)
		}
	})

	t.Run("status filter", func(t *testing.T) {
		got := applyContactFilters(contacts, makeRequest(map[string]string{"status": "new"}))
		if len(got) != 1 || got[0].Name != "Sam Lee" {
			t.Errorf("unexpected result: %v", got)
		}
	})

	t.Run("tag filter case-insensitive", func(t *testing.T) {
		got := applyContactFilters(contacts, makeRequest(map[string]string{"tag": "WARM"}))
		if len(got) != 2 {
			t.Errorf("got %d, want 2", len(got))
		}
	})

	t.Run("has_followup=true", func(t *testing.T) {
		got := applyContactFilters(contacts, makeRequest(map[string]string{"has_followup": "true"}))
		if len(got) != 2 {
			t.Errorf("got %d, want 2", len(got))
		}
	})

	t.Run("followup_before", func(t *testing.T) {
		got := applyContactFilters(contacts, makeRequest(map[string]string{"followup_before": "2024-04-30"}))
		if len(got) != 1 || got[0].Name != "Alex Kim" {
			t.Errorf("unexpected result: %v", got)
		}
	})

	t.Run("followup_after", func(t *testing.T) {
		got := applyContactFilters(contacts, makeRequest(map[string]string{"followup_after": "2024-04-15"}))
		if len(got) != 1 || got[0].Name != "Jordan Park" {
			t.Errorf("unexpected result: %v", got)
		}
	})
}
