package jdextract

import "testing"

// sampleJD is a real jina.ai markdown response for a VML Copywriter posting fetched from LinkedIn.
// Source: https://www.vml.com/careers/job/8234798002-ca-copywriter?gh_jid=8234798002
// Used to surface real-world regex gaps rather than sanitised synthetic input.
const sampleJD = `Title: Copywriter | Careers | VML

URL Source: https://www.vml.com/careers/job/8234798002-ca-copywriter?gh_jid=8234798002

Markdown Content:
Copywriter
----------

#### **Brand:** VML

#### **Capability:** Creative

#### **Location:**Toronto, Canada

#### **Last Updated:**2/25/2026

#### **Requisition ID:**12108

### **About VML**

VML is a leading creative company that combines brand experience, customer experience,
and commerce, creating connected brands to drive growth.

We're seeking an Intermediate Copywriter to craft compelling, on-brand copy that drives
engagement and performance across integrated campaigns, product experiences, and content.

**Key Responsibilities**

*   Translate briefs into messaging strategies, narratives, and copy across digital,
    social, email, web, print, OOH, and video/radio.

*   Write headlines, body copy, taglines, CTAs, scripts, and website copy that reflect
    brand voice and meet business objectives.

*   Partner with art directors to concept integrated ideas; present work with clear
    rationale and incorporate feedback thoughtfully.

*   Edit and proof for grammar, style, and consistency; manage file/version control.

*   Manage multiple workstreams, estimate level of effort, and deliver on time and within scope.

**Qualifications**

*   3-4 years of professional copywriting experience (agency or in-house).

*   Portfolio showcasing concept-driven campaigns, digital-first copy, performance
    creative, and cohesive brand storytelling.

$65,000—$115,000 CAD

We believe the best work happens when we're together. That's why we've adopted a hybrid
approach, with teams in the office an average of four days a week.`

// sampleJD2 is a real jina.ai markdown response for a Felix Senior Copywriter posting on Ashby.
// Source: https://jobs.ashbyhq.com/Felix/0d65c993-c9e7-4957-a454-b6c6186e3f1b
// Key structural difference from sampleJD: Ashby uses **bold paragraphs** for section headers,
// not markdown ATX headings. headingRe is completely blind to this document's structure.
const sampleJD2 = `Title: Senior Copywriter

URL Source: https://jobs.ashbyhq.com/Felix/0d65c993-c9e7-4957-a454-b6c6186e3f1b

Markdown Content:
[Overview](https://jobs.ashbyhq.com/Felix/0d65c993-c9e7-4957-a454-b6c6186e3f1b)[Application](https://jobs.ashbyhq.com/Felix/0d65c993-c9e7-4957-a454-b6c6186e3f1b/application)

**About Felix**

Felix is Canada's first end-to-end platform providing on-demand treatment for everyday health.
Felix has delivered 24 quarters of consecutive growth in both revenue and profit.

**The Role**

We are seeking an experienced senior copywriter with a deep understanding of writing, marketing,
and creativity to join the Felix brand team.

**In this role, you will:**

*   Establish tone of voice, consistent terminology, and tonal considerations for the brand

*   Collaborate with the growth team to develop copy and messaging for static and video digital
    ad creative for always-on paid social marketing

*   Write scripts for both brand & functional level TVCs for mass-reach campaigns

*   Work closely with compliance teams to balance creativity with accuracy and compliance

**We're looking for someone who:**

*   Has 7+ years of writing experience, including 3+ years of in-house experience

*   Is comfortable co-writing and receiving creative briefs

*   Has a portfolio of relevant work experience

**Benefits**

*   Full medical, dental and vision benefits

*   Remote first, work from anywhere in Canada

*   Stock option grant

**Location:**Toronto, Remote (Canada). We have an office in Toronto, but are currently working
remotely and open to candidates from anywhere in Canada.`

// ---------------------------------------------------------------------------
// TestClassifyLine — one case per NodeType, plus negatives
// ---------------------------------------------------------------------------

func TestClassifyLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Jina metadata
		{"jina_title", "Title: Senior Copywriter", NodeJinaTitle},
		{"jina_title with site suffix", "Title: Copywriter | Careers | VML", NodeJinaTitle},
		{"jina_url", "URL Source: https://jobs.ashbyhq.com/Felix/abc", NodeJinaURL},
		{"jina_marker", "Markdown Content:", NodeJinaMarker},

		// Structural noise
		{"setext dashes", "----------", NodeSetextUnderline},
		{"setext equals", "======", NodeSetextUnderline},
		{"nav_link single", "[Overview](https://example.com)", NodeNavLink},
		{"nav_link double", "[Overview](https://example.com)[Application](https://example.com/app)", NodeNavLink},

		// Salary fires before structure
		{"salary bare", "$65,000—$115,000 CAD", NodeSalary},
		{"salary in bullet", "*   Compensation: $80k–$120k per year", NodeSalary},

		// Years exp fires before structure
		{"years_exp plus", "*   Has 7+ years of writing experience, including 3+ years of in-house experience", NodeYearsExp},
		{"years_exp range", "*   3-4 years of professional copywriting experience (agency or in-house).", NodeYearsExp},

		// Headings — meta_field (key:value form)
		{"meta_field brand", "#### **Brand:** VML", NodeMetaField},
		{"meta_field capability", "#### **Capability:** Creative", NodeMetaField},
		{"meta_field location", "#### **Location:**Toronto, Canada", NodeMetaField},
		{"meta_field req id", "#### **Requisition ID:**12108", NodeMetaField},

		// Headings — section_header (known vocabulary)
		{"section_header atx", "### **About VML**", NodeSectionHeader},
		{"section_header bold about", "**About Felix**", NodeSectionHeader},
		{"section_header responsibilities", "**Key Responsibilities**", NodeSectionHeader},
		{"section_header qualifications", "**Qualifications**", NodeSectionHeader},
		{"section_header benefits", "**Benefits**", NodeSectionHeader},
		{"section_header in this role", "**In this role, you will:**", NodeSectionHeader},

		// Headings — job_title (seniority, not a section vocab word)
		{"job_title atx", "## Senior Copywriter", NodeJobTitle},
		{"job_title bold", "**Lead Product Designer**", NodeJobTitle},
		{"job_title intermediate", "**Intermediate Copywriter**", NodeJobTitle},

		// Headings — generic fallback
		{"heading generic", "**The Role**", NodeSectionHeader}, // "the role" is in vocab
		{"heading no vocab", "## Felix", NodeHeading},
		{"heading bare company", "**Acme Corp**", NodeHeading},

		// Bullet (signal checks already exhausted above)
		{"bullet dash", "- Write compelling copy", NodeBullet},
		{"bullet asterisk", "*   Translate briefs into messaging strategies", NodeBullet},
		{"bullet indented", "  - Indented point", NodeBullet},

		// Location on a body line
		{"location remote line", "Remote first, work from anywhere in Canada", NodeLocation},
		{"location hybrid line", "We've adopted a hybrid approach.", NodeLocation},

		// Body
		{"body short", "Felix is a Canadian health platform.", NodeBody},
		{"body continuation indent", "    and cohesive brand storytelling.", NodeBody},

		// Drop: long low-info body (> 300 chars)
		{"drop long body", string(make([]byte, 301)), ""},

		// Negatives — should NOT match a high-specificity type
		{"not salary no dollar", "competitive salary offered", NodeBody},
		{"not nav_link has extra text", "[Overview](https://x.com) and more text", NodeBody},
		{"not setext mixed", "hello-world", NodeBody},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyLine(tt.input)
			if got != tt.want {
				t.Errorf("classifyLine(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestBuildProtoAST — integration over both real JDs
// ---------------------------------------------------------------------------

func TestBuildProtoAST(t *testing.T) {
	countType := func(nodes []JobDescriptionNode, nt string) int {
		n := 0
		for _, node := range nodes {
			if node.NodeType == nt {
				n++
			}
		}
		return n
	}
	hasType := func(nodes []JobDescriptionNode, nt string) bool {
		return countType(nodes, nt) > 0
	}

	t.Run("sampleJD (VML/Greenhouse)", func(t *testing.T) {
		nodes := buildProtoAST(sampleJD)

		if !hasType(nodes, NodeJinaTitle) {
			t.Error("expected jina_title node")
		}
		if !hasType(nodes, NodeJinaURL) {
			t.Error("expected jina_url node")
		}
		if n := countType(nodes, NodeMetaField); n < 5 {
			t.Errorf("got %d meta_field nodes, want at least 5 (Brand, Capability, Location, Last Updated, Requisition ID)", n)
		}
		if !hasType(nodes, NodeSectionHeader) {
			t.Error("expected at least one section_header node")
		}
		if !hasType(nodes, NodeSalary) {
			t.Error("expected salary node for $65,000—$115,000 CAD line")
		}
		if !hasType(nodes, NodeYearsExp) {
			t.Error("expected years_exp node for 3-4 years line")
		}
		if n := countType(nodes, NodeBullet); n < 4 {
			t.Errorf("got %d bullet nodes, want at least 4", n)
		}
	})

	t.Run("sampleJD2 (Felix/Ashby)", func(t *testing.T) {
		nodes := buildProtoAST(sampleJD2)

		if !hasType(nodes, NodeJinaTitle) {
			t.Error("expected jina_title node")
		}
		if !hasType(nodes, NodeJinaURL) {
			t.Error("expected jina_url node")
		}
		// Ashby uses bold headings — section_header must fire on them
		if n := countType(nodes, NodeSectionHeader); n < 3 {
			t.Errorf("got %d section_header nodes, want at least 3 (About Felix, The Role, In this role, Benefits)", n)
		}
		if !hasType(nodes, NodeYearsExp) {
			t.Error("expected years_exp node for 7+ years line")
		}
		if n := countType(nodes, NodeBullet); n < 5 {
			t.Errorf("got %d bullet nodes, want at least 5", n)
		}
	})
}

// ---------------------------------------------------------------------------
// TestParse — filterNodes removes noise types
// ---------------------------------------------------------------------------

func TestParse(t *testing.T) {
	noiseTypes := map[string]bool{
		NodeJinaMarker:      true,
		NodeSetextUnderline: true,
		NodeNavLink:         true,
	}

	for _, tc := range []struct {
		name string
		jd   string
	}{
		{"sampleJD", sampleJD},
		{"sampleJD2", sampleJD2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			nodes := Parse(tc.jd)
			for _, n := range nodes {
				if noiseTypes[n.NodeType] {
					t.Errorf("noise type %q survived filter: %q", n.NodeType, n.Content)
				}
			}
			if len(nodes) == 0 {
				t.Error("Parse returned no nodes")
			}
		})
	}
}
