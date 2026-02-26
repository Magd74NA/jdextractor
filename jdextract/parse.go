package jdextract

import "regexp"

var (
	// headingRe extracts the text of the first # or ## markdown heading.
	headingRe = regexp.MustCompile(`(?m)^#{1,2}\s+(.+)`)

	// separatorRe splits common "Role at Company" / "Role - Company" / "Role | Company" patterns.
	// Covers Greenhouse ("Role at Company"), Lever ("Role - Company"), Workday ("Role | Company").
	separatorRe = regexp.MustCompile(`(?i)^(.+?)\s+(?:at|[-–—|])\s+(.+)$`)
)
