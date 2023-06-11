package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"golang.org/x/exp/slices"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"
)

// markdownTemplate is the markdown template that will be used to generate output.
//
//go:embed templates/markdown.tmpl
var markdownTemplate string

var (
	importRegexBlock  = regexp.MustCompile(`^\s*"(.+)"\s*$`)
	importRegexSingle = regexp.MustCompile(`\s*"(.+)"\s*$`)
)

// PackageCount contains a package and its count.
type PackageCount struct {
	Package string
	Count   int
}

// Result contains the internal and external package counts.
type Result struct {
	Internal []PackageCount
	External []PackageCount
}

// PackageCounter counts internal and external packages in a Go codebase
type PackageCounter struct {
	dir        string
	exclude    string
	lte        int
	gte        int
	moduleName string
	result     Result
}

// NewPackageCounter returns a new PackageCounter.
// It takes in the directory to scan, module name to exclude, regex patterns to exclude files and counts-gte and counts-lte
// which can be used to filter packages in the results.
func NewPackageCounter(dir string, moduleName string, exclude string, lte, gte int) *PackageCounter {
	return &PackageCounter{
		dir:        dir,
		lte:        lte,
		gte:        gte,
		exclude:    exclude,
		moduleName: moduleName,
	}
}

// CountPackages counts the number of internal and external packages in the specified directory.
func (pc *PackageCounter) CountPackages() error {
	goFiles, err := pc.findGoFiles()
	if err != nil {
		return fmt.Errorf("could not find .go files: %s", err.Error())
	}

	// Count package occurrences in the Go files
	result, err := pc.countPackages(goFiles)
	if err != nil {
		return fmt.Errorf("could not count packages: %s", err.Error())
	}

	// Sort the package counts
	slices.SortFunc(result.Internal, sortPackageCounts)
	slices.SortFunc(result.External, sortPackageCounts)

	pc.result = result
	return nil
}

func sortPackageCounts(i, j PackageCount) bool {
	return i.Count > j.Count
}

// GenerateMarkdown generates markdown output with the internal and external package counts.
func (pc *PackageCounter) GenerateMarkdown() (string, error) {
	funcMap := map[string]any{
		"isEmpty": func(items []PackageCount) bool {
			return len(items) == 0
		},
	}
	tmpl, err := template.New("markdown.tmpl").Funcs(funcMap).Parse(markdownTemplate)
	if err != nil {
		return "", err
	}

	var res bytes.Buffer
	if err = tmpl.Execute(&res, pc.result); err != nil {
		return "", err
	}
	return res.String(), nil
}

// findGoFiles returns a slice of paths to Go files found in pc.dir, that match exclude patterns if any.
func (pc *PackageCounter) findGoFiles() ([]string, error) {
	var goFiles []string

	err := filepath.Walk(pc.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			if pc.exclude != "" {
				compiled, err := regexp.Compile(pc.exclude)
				if err != nil {
					log.Fatalf("bad regular expression: '%s' error: %s", pc.exclude, err.Error())
				}
				if compiled.MatchString(path) {
					return nil
				}
			}
			goFiles = append(goFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return goFiles, nil
}

// countPackages returns the internal and external package counts that are used in a Go codebase.
func (pc *PackageCounter) countPackages(files []string) (Result, error) {
	internalPackageCounts := make(map[string]int)
	externalPackageCounts := make(map[string]int)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()

			// Extract import statements from the Go file
			imports, err := pc.extractImports(file)
			if err != nil {
				fmt.Printf("Error extracting imports from %s: %v\n", file, err)
				return
			}

			// Increment the package count for each import
			mu.Lock()
			for _, pkg := range imports {
				if strings.Contains(pkg, pc.moduleName) {
					internalPackageCounts[pkg]++
				} else {
					externalPackageCounts[pkg]++
				}
			}
			mu.Unlock()
		}(file)
	}

	wg.Wait()

	var result Result
	for pkg, count := range internalPackageCounts {
		if count >= pc.gte && count <= pc.lte {
			result.Internal = append(result.Internal, PackageCount{Package: pkg, Count: count})
		}
	}
	for pkg, count := range externalPackageCounts {
		if count >= pc.gte && count <= pc.lte {
			result.External = append(result.External, PackageCount{Package: pkg, Count: count})
		}
	}

	return result, nil
}

// extractImports returns the imports declared in a Go file.
func (pc *PackageCounter) extractImports(file string) (imports []string, err error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		ferr := f.Close()
		if err != nil {
			err = ferr
		}
	}(f)

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "import (") {
			// Start of import block
			break
		}
		// single import
		if strings.HasPrefix(line, "import \"") {
			if match := importRegexSingle.FindStringSubmatch(line); match != nil {
				imports = append(imports, match[1])
			}
		}
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, ")") {
			// End of import block
			break
		}
		if match := importRegexBlock.FindStringSubmatch(line); match != nil {
			imports = append(imports, match[1])
		}
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return imports, nil
}
