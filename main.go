package main

import (
	"crypto/md5"
	"fmt"
	"image"
	"image/png"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/fatal"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unipdf/common"
	"github.com/richardwilkes/unipdf/extractor"
	"github.com/richardwilkes/unipdf/model"
	"github.com/yookoala/realpath"
)

func main() {
	cmdline.AppVersion = "0.1"
	cmdline.CopyrightStartYear = "2019"
	cmdline.CopyrightHolder = "Richard A. Wilkes"
	cmdline.AppIdentifier = "com.trollworks.eximgpdf"
	common.Log = &toLogger{level: common.LogLevelWarning}
	cl := cmdline.New(true)
	fatal.IfErr(extractImages(cl.Parse(os.Args[1:])))
}

func extractImages(paths []string) error {
	// If no paths specified, use the current directory
	if len(paths) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			return errs.Wrap(err)
		}
		paths = append(paths, wd)
	}

	// Determine the actual root paths and prune out paths that are a subset of another
	set, err := rootPaths(paths)
	if err != nil {
		return errs.Wrap(err)
	}

	// Process the files
	var list []string
	if list, err = collectFiles(set); err != nil {
		return errs.Wrap(err)
	}
	for _, path := range list {
		if err = processFile(path); err != nil {
			return errs.Wrap(err)
		}
	}
	return nil
}

func rootPaths(paths []string) (map[string]bool, error) {
	set := make(map[string]bool)
	for _, path := range paths {
		actual, err := realpath.Realpath(path)
		if err != nil {
			return nil, errs.Wrap(err)
		}
		if _, exists := set[actual]; !exists {
			add := true
			for one := range set {
				var left, right string
				if left, err = filepath.Rel(one, actual); err != nil {
					return nil, errs.Wrap(err)
				}
				if right, err = filepath.Rel(actual, one); err != nil {
					return nil, errs.Wrap(err)
				}
				prefixed := strings.HasPrefix(left, "..")
				if prefixed != strings.HasPrefix(right, "..") {
					if prefixed {
						delete(set, one)
					} else {
						add = false
						break
					}
				}
			}
			if add {
				set[actual] = true
			}
		}
	}
	return set, nil
}

func collectFiles(set map[string]bool) ([]string, error) {
	var list []string
	for path := range set {
		if err := filepath.Walk(path, func(p string, info os.FileInfo, _ error) error {
			// Prune out hidden directories and files
			name := info.Name()
			if strings.HasPrefix(name, ".") {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			// If this is a pdf file, add it to the list
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(p), ".pdf") {
				list = append(list, p)
			}
			return nil
		}); err != nil {
			return nil, errs.Wrap(err)
		}
	}
	sort.Slice(list, func(i, j int) bool { return txt.NaturalLess(list[i], list[j], true) })
	return list, nil
}

func processFile(path string) error {
	slog.Info("examining", "path", path)
	f, err := os.Open(path)
	if err != nil {
		return errs.Wrap(err)
	}
	defer xio.CloseIgnoringErrors(f)
	var pdfReader *model.PdfReader
	if pdfReader, err = model.NewPdfReader(f); err != nil {
		return errs.Wrap(err)
	}
	var encrypted bool
	if encrypted, err = pdfReader.IsEncrypted(); err != nil {
		return errs.Wrap(err)
	}
	if encrypted {
		var decrypted bool
		if decrypted, err = pdfReader.Decrypt(nil); err != nil {
			return errs.Wrap(err)
		}
		if !decrypted {
			return errs.New("unable to decrypt")
		}
	}
	var numPages int
	if numPages, err = pdfReader.GetNumPages(); err != nil {
		return errs.Wrap(err)
	}
	hashes := make(map[[md5.Size]byte]bool)
	for n := 1; n <= numPages; n++ {
		var page *model.PdfPage
		if page, err = pdfReader.GetPage(n); err != nil {
			return errs.Wrap(err)
		}
		var ex *extractor.Extractor
		if ex, err = extractor.New(page); err != nil {
			return errs.Wrap(err)
		}
		var pi *extractor.PageImages
		if pi, err = ex.ExtractPageImages(&extractor.ImageExtractOptions{IncludeInlineStencilMasks: true}); err != nil {
			return errs.Wrap(err)
		}
		dir := path[:len(path)-4] // All incoming paths have a ".pdf" suffix
		for i, one := range pi.Images {
			hash := md5.Sum(one.Image.Data)
			if !hashes[hash] {
				hashes[hash] = true
				var img image.Image
				if img, err = one.Image.ToGoImage(); err != nil {
					return errs.Wrap(err)
				}
				if !fs.IsDir(dir) {
					if err = os.Mkdir(dir, 0o755); err != nil {
						return errs.Wrap(err)
					}
				}
				p := filepath.Join(dir, fmt.Sprintf("p%d_i%d.png", n, i+1))
				slog.Info("creating", "path", p)
				var w *os.File
				if w, err = os.Create(p); err != nil {
					return errs.Wrap(err)
				}
				if err = png.Encode(w, img); err != nil {
					return errs.Wrap(err)
				}
				if err = w.Close(); err != nil {
					return errs.Wrap(err)
				}
			}
		}
	}
	return nil
}

type toLogger struct {
	level common.LogLevel
}

func (log *toLogger) Error(format string, args ...interface{}) {
	if log.level >= common.LogLevelError {
		slog.Error(fmt.Sprintf(format, args...))
	}
}

func (log *toLogger) Warning(format string, args ...interface{}) {
	if log.level >= common.LogLevelWarning {
		slog.Warn(fmt.Sprintf(format, args...))
	}
}

func (log *toLogger) Notice(format string, args ...interface{}) {
	if log.level >= common.LogLevelNotice {
		slog.Info(fmt.Sprintf(format, args...))
	}
}

func (log *toLogger) Info(format string, args ...interface{}) {
	if log.level >= common.LogLevelInfo {
		slog.Info(fmt.Sprintf(format, args...))
	}
}

func (log *toLogger) Debug(format string, args ...interface{}) {
	if log.level >= common.LogLevelDebug {
		slog.Debug(fmt.Sprintf(format, args...))
	}
}

func (log *toLogger) Trace(format string, args ...interface{}) {
	if log.level >= common.LogLevelTrace {
		slog.Debug(fmt.Sprintf(format, args...))
	}
}

func (log *toLogger) IsLogLevel(level common.LogLevel) bool {
	return log.level >= level
}
