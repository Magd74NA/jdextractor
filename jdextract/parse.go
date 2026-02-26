package jdextract

import (
	"regexp"
	"strings"

	_ "github.com/toon-format/toon-go"
)

// NodeType constants, ordered from most generic to most specific.
// classifyLine returns the most specific type it can determine.
const (
	// Generic — low priority; drop if long.
	NodeUnknown = "unknown"
	NodeBody    = "body"

	// Structural list item.
	NodeBullet = "bullet"

	// Heading family — always keep, narrowed by heading text.
	NodeHeading       = "heading"        // generic: no further signal
	NodeSectionHeader = "section_header" // known section vocabulary
	NodeJobTitle      = "job_title"      // contains seniority indicator
	NodeMetaField     = "meta_field"     // "Key: Value" ATS field

	// Inline signal lines — always keep.
	NodeLocation = "location"
	NodeYearsExp = "years_exp"
	NodeSalary   = "salary"

	// Jina metadata — always keep, highest confidence.
	NodeJinaURL   = "jina_url"
	NodeJinaTitle = "jina_title"

	// Structural noise — always drop.
	NodeJinaMarker      = "jina_marker"
	NodeSetextUnderline = "setext_underline"
	NodeNavLink         = "nav_link"
)

// maxBodyLen is the character threshold above which body/unknown nodes are dropped.
// Long unstructured paragraphs are low information density for the LLM.
const maxBodyLen = 300

var (
	jinaTitleRe = regexp.MustCompile(`^Title:\s+.+`)
	urlSourceRe = regexp.MustCompile(`^URL Source:\s+.+`)

	// headingRe matches ATX headings; group 2 = heading text (may contain inline markdown).
	headingRe = regexp.MustCompile(`^(#{1,6})\s+(.+)`)

	// boldHeadingRe matches standalone bold lines; group 1 = inner text.
	// Greedy \*{2,} at both ends strips all surrounding asterisks regardless of nesting.
	boldHeadingRe = regexp.MustCompile(`^\*{2,}(.+?)\*{2,}$`)

	// bulletRe matches markdown list items with - or * prefix.
	bulletRe = regexp.MustCompile(`^[ \t]*[-*]\s+.+`)

	// setextUnderlineRe matches setext heading underlines (artifact, not content).
	setextUnderlineRe = regexp.MustCompile(`^[-=]{2,}$`)

	// navLinkRe matches lines that consist entirely of markdown links.
	navLinkRe = regexp.MustCompile(`^(\[.+?\]\(.+?\))+$`)

	// metaFieldRe detects "Key: Value" or "Key:Value" in heading text.
	// Colon must have non-empty content after it to distinguish from trailing colons.
	metaFieldRe = regexp.MustCompile(`^[^:]+:\s*.+$`)

	// sectionVocabRe matches known section header vocabulary.
	sectionVocabRe = regexp.MustCompile(`(?i)\b(about|overview|summary|introduction|who we are|the role|about the role|position|responsibilities|what you.ll do|in this role|key responsibilities|you will|requirements|qualifications|what we.re looking for|you bring|you have|skills|benefits|perks|what we offer|compensation|why us|culture|values|team|location|work arrangement)\b`)

	// seniorityRe detects seniority level in heading text.
	seniorityRe = regexp.MustCompile(`(?i)\b(junior|associate|mid|intermediate|senior|staff|principal|lead|head of|director|vp|vice president)\b`)

	// salaryRe matches salary figures and ranges: "$120k", "$80,000—$115,000", "$50/hr".
	salaryRe = regexp.MustCompile(`(?i)\$\s*[\d,]+(?:\.\d+)?(?:\s*[kK])?\s*(?:[-–—]\s*\$?\s*[\d,]+(?:\.\d+)?(?:\s*[kK])?)?(?:\s*(?:per\s+(?:year|annum|hour)|\/(?:yr|hr|hour|year)))?`)

	// yearsExpRe matches experience requirements: "3+ years", "2-5 years".
	yearsExpRe = regexp.MustCompile(`(?i)(\d+)\s*\+?\s*(?:to|[-–])\s*(\d+)\s*years?|(\d+)\+\s*years?`)

	// locationRe matches location strings or work arrangements.
	locationRe = regexp.MustCompile(`(?i)\b(remote|hybrid|on.?site|[A-Z][a-z]+(?:\s[A-Z][a-z]+)?,\s*[A-Za-z]{2,}(?:,\s*[A-Za-z]+)?)\b`)
)

// JobDescriptionNode is one line of the jina.ai markdown response, classified by type.
type JobDescriptionNode struct {
	Content  string
	NodeType string
}

// dropAlways lists node types that are always removed by filterNodes.
var dropAlways = map[string]bool{
	NodeJinaMarker:      true,
	NodeSetextUnderline: true,
	NodeNavLink:         true,
}

// classifyLine returns the most specific NodeType for a single non-empty line.
// Returns "" to signal that the line should be dropped entirely.
func classifyLine(line string) string {
	trimmed := strings.TrimSpace(line)

	// Jina metadata — highest specificity, checked first.
	if jinaTitleRe.MatchString(trimmed) {
		return NodeJinaTitle
	}
	if urlSourceRe.MatchString(trimmed) {
		return NodeJinaURL
	}
	if trimmed == "Markdown Content:" {
		return NodeJinaMarker
	}

	// Structural artifacts — drop-always.
	if setextUnderlineRe.MatchString(trimmed) {
		return NodeSetextUnderline
	}
	if navLinkRe.MatchString(trimmed) {
		return NodeNavLink
	}

	// Signal types — fire before structure so a bullet with salary becomes "salary",
	// preserving the high-value signal over the structural label.
	if salaryRe.MatchString(trimmed) {
		return NodeSalary
	}
	if yearsExpRe.MatchString(trimmed) {
		return NodeYearsExp
	}

	// Heading structure — ATX then bold standalone.
	if m := headingRe.FindStringSubmatch(trimmed); m != nil {
		text := strings.ReplaceAll(m[2], "*", "")
		return classifyHeadingText(text)
	}
	if m := boldHeadingRe.FindStringSubmatch(trimmed); m != nil {
		text := strings.TrimRight(m[1], ":")
		return classifyHeadingText(text)
	}

	// List items.
	if bulletRe.MatchString(trimmed) {
		return NodeBullet
	}

	// Location — bare body lines containing location signals.
	if locationRe.MatchString(trimmed) {
		return NodeLocation
	}

	// Body: drop if too long to preserve LLM context.
	if len(trimmed) > maxBodyLen {
		return ""
	}
	return NodeBody
}

// classifyHeadingText narrows a heading from generic to specific based on its text.
// Order: meta_field → section_header → job_title → heading.
func classifyHeadingText(text string) string {
	text = strings.TrimSpace(text)
	if metaFieldRe.MatchString(text) {
		return NodeMetaField
	}
	if sectionVocabRe.MatchString(text) {
		return NodeSectionHeader
	}
	if seniorityRe.MatchString(text) {
		return NodeJobTitle
	}
	return NodeHeading
}

// buildProtoAST classifies every non-empty line and returns the full unfiltered AST.
// Callers can inspect this before filtering for debugging.
func buildProtoAST(s string) []JobDescriptionNode {
	lines := strings.Split(s, "\n")
	nodes := make([]JobDescriptionNode, 0, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		nt := classifyLine(line)
		if nt == "" {
			continue // long low-info line, drop
		}
		nodes = append(nodes, JobDescriptionNode{
			Content:  line,
			NodeType: nt,
		})
	}

	return nodes
}

// filterNodes removes noise from the AST: always-drop types and
// long body/unknown nodes that would waste LLM context.
func filterNodes(nodes []JobDescriptionNode) []JobDescriptionNode {
	out := make([]JobDescriptionNode, 0, len(nodes))
	for _, n := range nodes {
		if dropAlways[n.NodeType] {
			continue
		}
		out = append(out, n)
	}
	return out
}

// Parse builds a filtered AST from a jina.ai markdown response.
func Parse(s string) []JobDescriptionNode {
	return filterNodes(buildProtoAST(s))
}
