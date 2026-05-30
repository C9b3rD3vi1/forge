package models

import "regexp"

var (
	mdHeading    = regexp.MustCompile(`(?m)^#{1,6}\s+`)
	mdBoldItalic = regexp.MustCompile(`\*{1,3}(.*?)\*{1,3}`)
	mdBoldItalic2 = regexp.MustCompile(`_{1,3}(.*?)_{1,3}`)
	mdCodeBlock  = regexp.MustCompile("(?s)```[^`]*```")
	mdCodeInline = regexp.MustCompile("`([^`]+)`")
	mdLink       = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	mdImage      = regexp.MustCompile(`!\[([^\]]*)\]\([^)]+\)`)
	mdBlockquote = regexp.MustCompile(`(?m)^>\s+`)
	mdHr         = regexp.MustCompile(`(?m)^[-*_]{3,}\s*$`)
	htmlTags     = regexp.MustCompile(`<[^>]+>`)
	whitespace   = regexp.MustCompile(`\s+`)
)

func StripMarkdown(md string) string {
	s := mdHeading.ReplaceAllString(md, "")
	s = mdCodeBlock.ReplaceAllString(s, " ")
	s = mdBoldItalic.ReplaceAllString(s, "$1")
	s = mdBoldItalic2.ReplaceAllString(s, "$1")
	s = mdCodeInline.ReplaceAllString(s, "$1")
	s = mdLink.ReplaceAllString(s, "$1")
	s = mdImage.ReplaceAllString(s, "$1")
	s = mdBlockquote.ReplaceAllString(s, "")
	s = mdHr.ReplaceAllString(s, "")
	s = htmlTags.ReplaceAllString(s, "")
	s = whitespace.ReplaceAllString(s, " ")
	return s
}
