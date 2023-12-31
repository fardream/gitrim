package gitrim

import (
	"fmt"
	"path"
	"slices"
	"strings"
)

// PatternFilterSegment is a segment in [PatternFilter]
type PatternFilterSegment string

// PatternFilter filters the entries according to a restricted pattern of [gitignore]
//
//   - `**` is for multi level directories, and it can only appear once in the match.
//   - `*` is for match one level of names.
//   - `!` and escapes are unsupported.
//   - paths are always relative to the root. For example, `LICENSE` will only match `LICENSE` in the root of the repo. To match `LICENSE` at all directory levels, use `**/LICENSE`.
//
// [gitignore]: https://git-scm.com/docs/gitignore
type PatternFilter struct {
	inputPattern    string
	filterSegments  []PatternFilterSegment
	multiLevelIndex int
	// isDirOnly indicates if the filter is for directories only.
	// this is false indicating this matches files and directories.
	isDirOnly bool

	beforefilters []PatternFilterSegment
	afterfilters  []PatternFilterSegment
}

var _ Filter = (*PatternFilter)(nil)

func NewPatternFilter(pattern string) (*PatternFilter, error) {
	trimmedpattern := strings.TrimSpace(pattern)
	p := &PatternFilter{
		inputPattern:    trimmedpattern,
		multiLevelIndex: -1,
	}

	// remove trailing **/ or **
	if strings.HasSuffix(trimmedpattern, "**/") {
		trimmedpattern = strings.TrimSuffix(trimmedpattern, "**/")
	} else {
		// if strings.HasSuffix(trimmedpattern, "**") is unnecessary
		trimmedpattern = strings.TrimSuffix(trimmedpattern, "**")
	}

	logger.Debug("pattern", "input", pattern, "trimmed", trimmedpattern)

	if trimmedpattern == "/" || trimmedpattern == "" {
		return nil, fmt.Errorf("'%s' is invalid pattern", trimmedpattern)
	}

	p.isDirOnly = strings.HasSuffix(p.inputPattern, "/")
	segs := strings.Split(p.inputPattern, "/")
	p.filterSegments = make([]PatternFilterSegment, 0, len(segs))
	for _, s := range segs {
		p.filterSegments = append(p.filterSegments, PatternFilterSegment(s))
	}
	if len(p.filterSegments) == 0 {
		return nil, fmt.Errorf("input pattern %s has zero path segments", pattern)
	}

	// if the pattern ends with /, the last segment is empty
	if p.isDirOnly {
		// even after this, the segs should still has at least 1 element
		p.filterSegments = p.filterSegments[:len(p.filterSegments)-1]
	}

	// if the first element is empty, there is a root / at the start of the pattern.
	if p.filterSegments[0] == "" {
		p.filterSegments = p.filterSegments[1:]
	}

	if len(p.filterSegments) == 0 {
		return nil, fmt.Errorf("zero path segment left after removing leading/trailing white spaces: '%s'", trimmedpattern)
	}

	for idx, seg := range p.filterSegments {
		// check on the segment
		if seg == "**" {
			if p.multiLevelIndex >= 0 {
				return nil, fmt.Errorf("at most 1 ** pattern can appear in pattern, but %s has more than 1", trimmedpattern)
			}
			if idx == len(p.filterSegments)-1 {
				return nil, fmt.Errorf("trailing ** or **/ hasn't been removed")
			}
			p.multiLevelIndex = idx
		} else if strings.Contains(string(seg), "**") {
			return nil, fmt.Errorf("segment: %s contains **, which is invalid", seg)
		} else {
			_, err := path.Match(string(seg), "abc")
			if err != nil {
				return nil, fmt.Errorf("pattern segment %s is not valid: %w", seg, err)
			}
		}
	}

	if p.multiLevelIndex >= 0 {
		p.beforefilters = p.filterSegments[:p.multiLevelIndex]
		p.afterfilters = []PatternFilterSegment{}
		if len(p.filterSegments) > p.multiLevelIndex+1 {
			p.afterfilters = p.filterSegments[p.multiLevelIndex+1:]
		}
		logger.Debug("multi-level-filter", "before", p.beforefilters, "after", p.afterfilters)
	}

	return p, nil
}

func (f *PatternFilter) Filter(paths []string, isdir bool) FilterResult {
	if f.multiLevelIndex < 0 {
		// not multiLevelIndex, use simple filter
		return nonMultiLevelFilter(isdir, paths, f.filterSegments, f.isDirOnly)
	}

	beforefilters := f.beforefilters[:]
	afterfilters := f.afterfilters[:]

	predirpaths := paths[:]
	if !isdir {
		predirpaths = predirpaths[:len(predirpaths)-1]
	}

	if len(predirpaths) >= f.multiLevelIndex {
		predirpaths = predirpaths[:f.multiLevelIndex]
	}

	remainingpaths := paths[len(predirpaths):]

	beforesult := PatternDirFilter(predirpaths, beforefilters)
	switch beforesult {
	case FilterResult_In:
		if len(afterfilters) == 0 {
			return FilterResult_In
		}
		if isdir {
			if len(remainingpaths) == 0 {
				return FilterResult_DirDive
			}
			r := FilterResult_DirDive
			for i := 0; i < len(remainingpaths)-len(afterfilters)+1; i++ {
				afterpaths := remainingpaths[i:]
				tr := nonMultiLevelFilter(isdir, afterpaths, afterfilters, f.isDirOnly)
				if tr > r {
					r = tr
				}
				if r == FilterResult_In {
					return r
				}
			}

			return r
		} else {
			end := len(remainingpaths) - len(afterfilters) + 1
			for start := 0; start < end; start++ {
				tr := nonMultiLevelFilter(isdir, remainingpaths[start:], afterfilters, f.isDirOnly)

				if tr == FilterResult_In {
					return FilterResult_In
				}
			}
			return FilterResult_Out
		}
	case FilterResult_DirDive:
		if !isdir || len(remainingpaths) > 0 {
			return FilterResult_Out
		}
		return FilterResult_DirDive
	case FilterResult_Out:
		fallthrough
	default:
		return FilterResult_Out
	}
}

func nonMultiLevelFilter(isdir bool, paths []string, filters []PatternFilterSegment, filterIsDirOnly bool) FilterResult {
	switch {
	case isdir:
		// input is dir, do DirFilter
		return PatternDirFilter(paths, filters)
	case !isdir && filterIsDirOnly:
		// input is a file, so it will only be in if its dir is in
		if PatternDirFilter(paths[:len(paths)-1], filters) != FilterResult_In {
			return FilterResult_Out
		} else {
			return FilterResult_In
		}
	case !isdir && !filterIsDirOnly:
		if len(paths) < len(filters) {
			return FilterResult_Out
		}

		return PatternDirFilter(paths, filters)
	default:
		return FilterResult_Out
	}
}

// PatternDirFilter filters the directory according to a directory filter.
//
// The result is "In", if filters match all the leading path segments, and there are zero or more path trailing.
// Below are two examples of "In"
//
//	// path matches all filter segments, and path has extra segments
//	| p | p | p | p
//	| f | f | f
//	// path matches all filter segments, and path has no extra segmetns
//	| p | p | p
//	| f | f | f
//
// The result is "DirDive", if size of path segments is smaller than filters, and those path segments match the corresponding filters
//
//	| p | p | p
//	| f | f | f | f
//
// For empty paths or filtersegs, it will always return "Out".
func PatternDirFilter(paths []string, filtersegs []PatternFilterSegment) FilterResult {
	if len(paths) == 0 || len(filtersegs) == 0 {
		return FilterResult_Out
	}

	if len(paths) >= len(filtersegs) {
		for i, fseg := range filtersegs {
			matched, err := path.Match(string(fseg), paths[i])
			if err != nil {
				logger.Warn("failed match", "pattern", fseg, "name", paths[i], "error", err.Error())
				return FilterResult_Out
			}
			if !matched {
				return FilterResult_Out
			}
		}

		return FilterResult_In
	} else {
		for i, p := range paths {
			matched, err := path.Match(string(filtersegs[i]), p)
			if err != nil {
				logger.Warn("failed match", "pattern", filtersegs[i], "name", p, "error", err.Error())
				return FilterResult_Out
			}
			if !matched {
				return FilterResult_Out
			}
		}

		return FilterResult_DirDive
	}
}

// LoadPatternFilterFromString loads the string content of a pattern file like .gitignore.
// If ignoreUnsupported is set to false, the loader will error if the any unsupported patterns like ! (reverse) is encountered.
func LoadPatternFilterFromString(str string, ignoreUnsupported bool) ([]*PatternFilter, error) {
	lines := strings.Split(str, "\n")
	result := make([]*PatternFilter, 0, len(lines))

	for i, line := range lines {
		line := strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "!") {
			if ignoreUnsupported {
				continue
			}
			return nil, fmt.Errorf("line %d of input file %s contains unsupported pattern", i, line)
		}

		filter, err := NewPatternFilter(line)
		if err != nil {
			return nil, fmt.Errorf("failed to generate pattern for line %d (%s): %w", i, line, err)
		}

		result = append(result, filter)
	}

	return result, nil
}

// LoadPatternStringFromString loads from the string content of a pattern file like .gitignore.
// Similar to [LoadPatternFilterFromString], a false ignoreUnsupported will error if unsupported patterns are encountered.
//
// The result are lexigraphically sorted with [slices.Sort] and then feed into [slices.Compact] to remove duplicates.
func LoadPatternStringFromString(str string, ignoreUnsupported bool) ([]string, error) {
	lines := strings.Split(str, "\n")
	result := make([]string, 0, len(lines))

	for i, line := range lines {
		line := strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "!") {
			if ignoreUnsupported {
				continue
			}
			return nil, fmt.Errorf("line %d of input file %s contains unsupported pattern", i, line)
		}

		result = append(result, line)
	}

	slices.Sort(result)
	result = slices.Compact(result)

	return result, nil
}
