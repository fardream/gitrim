package gitrim

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/diff"
)

// FilePatchError is an error containing the information about the invalid file patch.
type FilePatchError struct {
	FromFile string
	ToFile   string
}

func (e *FilePatchError) ErrorFiles() []string {
	if e == nil {
		return nil
	}
	switch {
	case e.FromFile != "" && e.ToFile != "":
		return []string{e.FromFile, e.ToFile}
	case e.FromFile != "":
		return []string{e.FromFile}
	case e.ToFile != "":
		return []string{e.ToFile}
	default:
		return nil
	}
}

func (e *FilePatchError) Error() string {
	errfs := make([]string, 0, 2)
	if e.FromFile != "" {
		errfs = append(errfs, fmt.Sprintf("invalid from path: %s", e.FromFile))
	}
	if e.ToFile != "" {
		errfs = append(errfs, fmt.Sprintf("invalid to path: %s", e.ToFile))
	}

	return strings.Join(errfs, "|")
}

// FilePatchCheckResult contains the result from [CheckFilePatchAgainstFilter]
type FilePatchCheckResult struct {
	Errors []*FilePatchError
}

func (f *FilePatchCheckResult) ErrorSlice() []error {
	if f == nil || len(f.Errors) == 0 {
		return nil
	}

	errs := make([]error, 0, len(f.Errors))

	for _, e := range f.Errors {
		errs = append(errs, e)
	}

	return errs
}

func (f *FilePatchCheckResult) ToError() error {
	errs := f.ErrorSlice()
	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

// CheckFilePatchAgainstFilter checks the [diff.FilePath] against the [Filter], to make sure both the from and to file are allowed under the filter.
// The returned list of [error] contains all [FilePatchError] which indicate the files flagged by the filter.
func CheckFilePatchAgainstFilter(filepatches []diff.FilePatch, filter Filter) *FilePatchCheckResult {
	r := &FilePatchCheckResult{}

	for _, afile := range filepatches {
		fromfile, tofile := afile.Files()

		fromfilename := ""
		if fromfile != nil {
			fromfilename = fromfile.Path()
		}
		tofilename := ""
		if tofile != nil {
			tofilename = tofile.Path()
		}

		var thiserr *FilePatchError
		if fromfile != nil && !filter.Filter(strings.Split(fromfilename, "/"), false).IsIn() {
			if thiserr == nil {
				thiserr = new(FilePatchError)
			}
			thiserr.FromFile = fromfilename
		}
		if tofile != nil && !filter.Filter(strings.Split(tofilename, "/"), false).IsIn() {
			if thiserr == nil {
				thiserr = new(FilePatchError)
			}
			thiserr.ToFile = tofilename
		}
		if thiserr != nil {
			r.Errors = append(r.Errors, thiserr)
		}
	}

	return r
}
