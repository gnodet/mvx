package version

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Version represents a semantic version
type Version struct {
	Major int
	Minor int
	Patch int
	Pre   string // Pre-release identifier (e.g., "rc1", "beta2")
	Build string // Build metadata
}

// Spec represents a version specification that can match multiple versions
type Spec struct {
	Raw        string
	Constraint string // "exact", "major", "minor", "range", "latest"
	Major      int
	Minor      int
	Patch      int
	Pre        string
}

// ParseVersion parses a version string into a Version struct
func ParseVersion(v string) (*Version, error) {
	// Remove 'v' prefix if present
	v = strings.TrimPrefix(v, "v")
	
	// Regex for semantic version: major.minor.patch[-pre][+build]
	re := regexp.MustCompile(`^(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-([a-zA-Z0-9\-\.]+))?(?:\+([a-zA-Z0-9\-\.]+))?$`)
	matches := re.FindStringSubmatch(v)
	
	if matches == nil {
		return nil, fmt.Errorf("invalid version format: %s", v)
	}
	
	major, _ := strconv.Atoi(matches[1])
	minor := 0
	patch := 0
	
	if matches[2] != "" {
		minor, _ = strconv.Atoi(matches[2])
	}
	if matches[3] != "" {
		patch, _ = strconv.Atoi(matches[3])
	}
	
	return &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Pre:   matches[4],
		Build: matches[5],
	}, nil
}

// ParseSpec parses a version specification
func ParseSpec(spec string) (*Spec, error) {
	spec = strings.TrimSpace(spec)
	
	// Handle special cases
	if spec == "latest" || spec == "" {
		return &Spec{
			Raw:        spec,
			Constraint: "latest",
		}, nil
	}
	
	// Handle range specifications (future enhancement)
	if strings.Contains(spec, ">=") || strings.Contains(spec, "<=") || strings.Contains(spec, "~") || strings.Contains(spec, "^") {
		return nil, fmt.Errorf("range specifications not yet implemented: %s", spec)
	}
	
	// Parse as version
	v, err := ParseVersion(spec)
	if err != nil {
		return nil, err
	}
	
	// Determine constraint type based on how specific the version is
	constraint := "exact"
	if strings.Count(spec, ".") == 0 {
		// Only major version specified (e.g., "3")
		constraint = "major"
	} else if strings.Count(spec, ".") == 1 {
		// Major.minor specified (e.g., "3.9")
		constraint = "minor"
	}
	
	return &Spec{
		Raw:        spec,
		Constraint: constraint,
		Major:      v.Major,
		Minor:      v.Minor,
		Patch:      v.Patch,
		Pre:        v.Pre,
	}, nil
}

// String returns the string representation of a version
func (v *Version) String() string {
	result := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Pre != "" {
		result += "-" + v.Pre
	}
	if v.Build != "" {
		result += "+" + v.Build
	}
	return result
}

// Compare compares two versions. Returns -1 if v < other, 0 if equal, 1 if v > other
func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}
	
	// Handle pre-release versions
	if v.Pre == "" && other.Pre != "" {
		return 1 // Release version > pre-release
	}
	if v.Pre != "" && other.Pre == "" {
		return -1 // Pre-release < release version
	}
	if v.Pre != other.Pre {
		return strings.Compare(v.Pre, other.Pre)
	}
	
	return 0
}

// Matches checks if a version matches a specification
func (s *Spec) Matches(v *Version) bool {
	switch s.Constraint {
	case "latest":
		return true // Latest matches any version (resolver will pick the highest)
	case "exact":
		return s.Major == v.Major && s.Minor == v.Minor && s.Patch == v.Patch && s.Pre == v.Pre
	case "major":
		return s.Major == v.Major
	case "minor":
		return s.Major == v.Major && s.Minor == v.Minor
	default:
		return false
	}
}

// Resolve finds the best matching version from a list of available versions
func (s *Spec) Resolve(availableVersions []string) (string, error) {
	if len(availableVersions) == 0 {
		return "", fmt.Errorf("no versions available")
	}
	
	// Parse all available versions
	var versions []*Version
	var versionMap = make(map[string]string) // version -> original string
	
	for _, vStr := range availableVersions {
		v, err := ParseVersion(vStr)
		if err != nil {
			continue // Skip invalid versions
		}
		versions = append(versions, v)
		versionMap[v.String()] = vStr
	}
	
	if len(versions) == 0 {
		return "", fmt.Errorf("no valid versions found")
	}
	
	// Filter matching versions
	var matching []*Version
	for _, v := range versions {
		if s.Matches(v) {
			matching = append(matching, v)
		}
	}
	
	if len(matching) == 0 {
		return "", fmt.Errorf("no versions match specification %s", s.Raw)
	}
	
	// Sort versions (highest first)
	sort.Slice(matching, func(i, j int) bool {
		return matching[i].Compare(matching[j]) > 0
	})
	
	// Return the highest matching version
	best := matching[0]
	if original, exists := versionMap[best.String()]; exists {
		return original, nil
	}
	return best.String(), nil
}

// SortVersions sorts a slice of version strings in descending order (newest first)
func SortVersions(versions []string) []string {
	var parsed []*Version
	var versionMap = make(map[string]string)
	
	for _, vStr := range versions {
		v, err := ParseVersion(vStr)
		if err != nil {
			continue
		}
		parsed = append(parsed, v)
		versionMap[v.String()] = vStr
	}
	
	sort.Slice(parsed, func(i, j int) bool {
		return parsed[i].Compare(parsed[j]) > 0
	})
	
	var result []string
	for _, v := range parsed {
		if original, exists := versionMap[v.String()]; exists {
			result = append(result, original)
		} else {
			result = append(result, v.String())
		}
	}
	
	return result
}
