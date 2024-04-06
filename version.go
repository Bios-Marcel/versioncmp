package versioncmp

import (
	"slices"
	"strconv"
	"strings"
)

type VersionCompareRules struct {
	CompareNightly bool
	CompareMeta    bool
}

// Compare will attempt to compare two versions and return the greater
// one. For valid usecases, check out the unit tests.
//
// Note that the first version is the one assumed to be older. This means, that
// incase we can't do a semantic comparison, but a forced comparison is desired,
// we will interpert "a != b" as "b is greter".
func Compare(originalA, originalB string, rules VersionCompareRules) string {
	if originalA == originalB {
		return ""
	}

	versionA := parse(originalA)
	versionB := parse(originalB)

	if !rules.CompareNightly && versionA.Nightly && versionB.Nightly {
		return ""
	}

	a := originalA
	b := originalB
	// Prevent index out of bounds and simplify follow-up code.
	if len(versionA.Values) > len(versionB.Values) {
		versionA, versionB = versionB, versionA
		a, b = b, a
	}

	// This compares semantic verioning and dates. Note that only dates of the
	// format yyyy-mm-dd work.
	var comparison int
	for index, groupA := range versionA.Values {
		groupB := versionB.Values[index]

		if len(groupA) > len(groupB) {
			a, b = b, a
			groupA, groupB = groupB, groupA
		}

		// Special treatment for sinners that use dd.mm.yyyy. If they use mm.dd.yyyy
		// or yyyy.dd.mm i genuinely don't care though.
		// FIXME Potentially there can be a date in additional to semver?
		// Do we need sepaerate groups of numbers?
		if len(groupA) == 3 && len(groupB) == 3 {
			// Last part might be a year. For now, the dates are between 1960
			// and 2030. It would be possible to use the current year, but for
			// now we'll hardcode this.
			if groupA[2] < 1960 || groupB[2] < 1960 ||
				groupA[2] > 2030 || groupB[2] > 2030 {
				goto NO_REVERSE_DATE
			}
			// No month
			if groupA[1] > 12 || groupB[1] > 12 {
				goto NO_REVERSE_DATE
			}
			// No day of month. FIXME Technically this means we can have dates
			// such as 31.02.2024, which don't exist.
			if groupA[0] > 31 || groupB[0] > 31 {
				goto NO_REVERSE_DATE
			}

			// All conditions are met, so we reverse the date into the expected
			// format, which is YYYY-MM-DD
			slices.Reverse(groupA)
			slices.Reverse(groupB)

		NO_REVERSE_DATE:
		}

		for index, valueA := range groupA {
			if groupB[index] < valueA {
				// a is bigger
				comparison = 1
				break
			}
			if groupB[index] > valueA {
				// a is smaller
				comparison = -1
				break
			}
		}

		// If we find any group that is bigger, we return right away, as the
		// first group counts.
		if comparison == 1 {
			return a
		}
		if comparison == -1 {
			return b
		}

		// For example `1.2.3` vs `1.2`. Same so far, but the longer group wins,
		// as the version MUST be higher, due to higher specificty, as the
		// specifity can't be reduced on an update without changing the higher
		// version components.
		if len(groupA) > len(groupB) {
			return a
		}
	}

	stabilityA := stabilityValue[versionA.Stability]
	stabilityB := stabilityValue[versionB.Stability]
	if stabilityA > stabilityB {
		return a
	}
	if stabilityA < stabilityB {
		return b
	}

	if rules.CompareMeta && !slices.Equal(versionA.Meta, versionB.Meta) {
		return originalB
	}

	return ""
}

var stabilityValue = map[string]int{
	"dev":    0,
	"alpha":  1,
	"beta":   2,
	"rc":     3,
	"pre":    4,
	"stable": 5,
}

// FIXME Instead contains checks and then clean out all non-numerics?
var stabilityMapping = map[string]string{
	"dev":     "dev",
	"devel":   "dev",
	"develop": "dev",

	"pre":        "pre",
	"prerel":     "pre",
	"prerelease": "pre",

	"alpha": "alpha",

	"beta": "beta",

	"rc":               "rc",
	"candidate":        "rc",
	"releasecandidate": "rc",
}

func parse(value string) version {
	version := version{
		Stability: "stable",
	}

	groups := split(value)
	for _, group := range groups {
		var values []uint64
		for _, part := range group {
			part = strings.ToLower(part)

			if strings.Contains(part, "nightly") {
				// Add as meta value? Use meta value for comparison if nightly
				// comparison has been set to true anyway?
				version.Nightly = true
				continue
			}

			for stabilityVerbose, mapped := range stabilityMapping {
				trimmed := strings.TrimSuffix(strings.TrimPrefix(part, stabilityVerbose), stabilityVerbose)
				if trimmed != part {
					version.Stability = mapped
					part = trimmed
					break
				}
			}

			parsed, err := strconv.ParseUint(part, 10, 32)
			// If not on an int, this is okay!
			if err == nil {
				values = append(values, parsed)
			} else {
				version.Meta = append(version.Meta, part)
			}
		}
		if len(values) > 0 {
			version.Values = append(version.Values, values)
		}
	}

	return version
}

type version struct {
	Stability string

	// Nightly
	Nightly bool

	// Meta contains any data that isn't directly comparable. We can only use
	// this to spot a difference in the versions, but not on a semantic level.
	Meta []string

	// Values are any numeric parts of the date. It's sorted from major to
	// minor. Note that this can be an arbitrary number between 1 and N.
	// If the actual version didn't contain a number, we will simply fill a `1`.
	Values [][]uint64
}

func split(version string) [][]string {
	var parts [][]string

	var nextStart int
	var lastSeparator rune

	// We can have multiple groups of numbers. For example semver + date. Often
	// these use different separator, in order to be able to differentiate
	// between the groups.
	for index, char := range version {
		switch char {
		case '.', '-', '_', '/', '\\', ';', ':':
			part := string(version[nextStart:index])
			nextStart = index + 1

			if len(parts) == 0 {
				parts = append(parts, []string{})
			}
			lastGroup := parts[len(parts)-1]
			// If we find someting like "pre-release", we want to ingore the
			// release, as the classification is "pre" either way. Things like
			// pre-beta don't exist, so we needn't worry about that.
			if len(lastGroup) > 0 && lastGroup[len(lastGroup)-1] == "pre" && part == "release" {
				continue
			}
			parts[len(parts)-1] = append(lastGroup, part)

			// add new group if there are still more parts
			if lastSeparator != char && index != len(version)-1 && lastSeparator != 0 {
				parts = append(parts, []string{})
			}
			lastSeparator = char
		}
	}

	// Single token, such as a commit hash
	if nextStart == 0 {
		return [][]string{{version}}
	}

	if lastPart := string(version[nextStart:]); lastPart != "" {
		parts[len(parts)-1] = append(parts[len(parts)-1], lastPart)
	} else {
	}

	return parts
}
