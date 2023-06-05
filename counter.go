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

//go:embed templates/markdown.tmpl
var markdownTemplate string

var (
	importRegexBlock  = regexp.MustCompile(`^\s*"(.+)"\s*$`)
	importRegexSingle = regexp.MustCompile(`\s*"(.+)"\s*$`)
)

type PackageCount struct {
	Package string
	Count   int
}

type Result struct {
	Internal []PackageCount
	External []PackageCount
}

type PackageCounter struct {
	dir        string
	exclude    string
	lte        int
	gte        int
	moduleName string
	result     Result
}

func NewPackageCounter(dir string, moduleName string, exclude string, lte, gte int) *PackageCounter {
	return &PackageCounter{
		dir:        dir,
		lte:        lte,
		gte:        gte,
		exclude:    exclude,
		moduleName: moduleName,
	}
}

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
