package configs

import "strings"

type skips []string

var (
	HTTP                       = false
	BaseUrl                    = ""
	Project                    = ""
	ChangelogHeaderLines       = 2
	MR                         = false
	SkipMsgs             skips = make([]string, 0)
	DefaultHead                = "## ChangeLog\n\n"
	// Tag filtering options
	MaxTagAgeDays             = 365  // Maximum age of tags to consider in days
	MaxTagCount               = 0    // Maximum number of recent tags to consider (0 = no limit)
	OnlyRecentTags            = false // Whether to filter tags by time
)

func ParseSkipMsg(msg string) error {
	msgs := strings.Split(msg, ",")
	for _, m := range msgs {
		SkipMsgs = append(SkipMsgs, m)
	}
	return nil
}

func (s *skips) ShouldSkip(msg string) bool {
	for _, m := range *s {
		if strings.TrimSpace(m) == strings.TrimSpace(msg) {
			return true
		}
	}
	return false
}
