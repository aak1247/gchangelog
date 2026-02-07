package configs

// PrereleaseSuffixOrder defines prerelease suffix ordering (small -> large).
// Examples: 1.0.0-alpha < 1.0.0-beta < 1.0.0-rc < 1.0.0
var PrereleaseSuffixOrder = []string{
	"alpha",
	"beta",
	"rc",
}

// PostreleaseSuffixOrder defines postrelease suffix ordering (small -> large).
// Examples: 1.0.0 < 1.0.0-hotfix < 1.0.0-hotfix2
var PostreleaseSuffixOrder = []string{
	"hotfix",
}
