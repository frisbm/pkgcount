package pkgcount

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/frisbm/pkgcount/internal/models"
	"github.com/frisbm/pkgcount/internal/resultgroup"
)

const (
	importone  = "import \""
	importmany = "import ("
)

var (
	importRegexBlock  = regexp.MustCompile(`^[\w_]*\s*"(.+)"\s*$`)
	importRegexSingle = regexp.MustCompile(`\s*"(.+)"\s*$`)

	knownSkipDirs = map[string]struct{}{
		"vendor":       {},
		".git":         {},
		".idea":        {},
		".vscode":      {},
		"node_modules": {},
	}
)

func Run(ctx context.Context, args models.Args) error {
	if err := args.Validate(); err != nil {
		return fmt.Errorf("args failed validation: %w", err)
	}

	moduleName, err := getModuleName(ctx, args.Dir)
	if err != nil {
		return fmt.Errorf("failed to get module name: %w", err)
	}
	args.ModuleName = moduleName

	results, err := countPackages(ctx, args)
	if err != nil {
		return fmt.Errorf("failed to count packages: %w", err)
	}

	var output []byte
	if args.Unrendered {
		output, err = results.Unrendered()
		if err != nil {
			return fmt.Errorf("failed to get unrendered markdown: %w", err)
		}
	} else {
		output, err = results.Rendered()
		if err != nil {
			return fmt.Errorf("failed to get rendered markdown: %w", err)
		}
	}

	if args.Out != "" {
		err = os.WriteFile(args.Out, output, 0777)
		if err != nil {
			return fmt.Errorf("failed to write output to file: %w", err)
		}
		return nil
	}

	_, err = os.Stdout.Write(output)
	if err != nil {
		return fmt.Errorf("failed to write output to stdout: %w", err)
	}
	return nil
}

func getModuleName(ctx context.Context, dir string) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "-m")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute 'go list -m': %w", err)
	}

	moduleName := strings.TrimSpace(string(output))
	// see https://github.com/golang/go/blob/c2c4a32f9e57ac9f7102deeba8273bcd2b205d3c/src/cmd/go/internal/modload/modfile.go#L665
	if moduleName == "" || moduleName == "command-line-arguments" {
		return "", fmt.Errorf("no go.mod file found in directory: %s", dir)
	}
	return moduleName, nil
}

// countImports returns the internal and external package counts that are used in a Go codebase.
func countPackages(ctx context.Context, args models.Args) (models.Result, error) {
	rg, _ := resultgroup.New[[2]*models.Counter](ctx, [2]*models.Counter{models.NewCounter(), models.NewCounter()})

	err := filepath.Walk(args.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walking directory: %w", err)
		}
		if info.IsDir() {
			if _, ok := knownSkipDirs[info.Name()]; ok {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		if args.ExcludeRegExp != nil {
			if args.ExcludeRegExp.MatchString(path) {
				return nil
			}
		}
		rg.Go(calculateCounts(path, args.ModuleName))
		return nil
	})

	if err != nil {
		return models.Result{}, fmt.Errorf("counting imports for files: %w", err)
	}

	c, err := rg.Wait()
	if err != nil {
		return models.Result{}, fmt.Errorf("waiting for goroutines: %w", err)
	}

	res := calculateResults(args.Gte, args.Lte, c)

	slices.SortFunc(res.Internal, sortPackageCounts)
	slices.SortFunc(res.External, sortPackageCounts)

	return res, nil
}

func sortPackageCounts(i, j models.PackageCount) int {
	diff := j.Count - i.Count
	if diff == 0 {
		return strings.Compare(i.Package, j.Package)
	}
	return diff
}

func calculateCounts(path, modname string) func(c *[2]*models.Counter) error {
	return func(counts *[2]*models.Counter) error {
		// Extract import statements from the Go file
		imports, err := extractImports(path)
		if err != nil {
			return fmt.Errorf("extracting imports: %w", err)
		}
		// Increment the package count for each import
		for _, pkg := range imports {
			if strings.Contains(pkg, modname) {
				counts[0].Add(pkg)
				continue
			}
			counts[1].Add(pkg)
		}
		return nil
	}
}

func calculateResults(gte, lte int, c [2]*models.Counter) models.Result {
	internalCounts := c[0].Counts()
	externalCounts := c[1].Counts()

	result := models.Result{
		Internal: make([]models.PackageCount, 0, len(internalCounts)),
		External: make([]models.PackageCount, 0, len(externalCounts)),
	}
	for pkg, count := range internalCounts {
		if count == nil {
			continue
		}
		val := int(count.Load())
		if val >= gte && val <= lte {
			result.Internal = append(result.Internal, models.PackageCount{Package: pkg, Count: val})
		}
	}
	for pkg, count := range externalCounts {
		if count == nil {
			continue
		}
		val := int(count.Load())
		if val >= gte && val <= lte {
			result.External = append(result.External, models.PackageCount{Package: pkg, Count: val})
		}
	}
	return result
}

// extractImports returns the imports declared in a Go file.
func extractImports(file string) (imports []string, err error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer func(f *os.File) {
		err = errors.Join(err, f.Close())
	}(f)

	scanner := bufio.NewScanner(f)

	readingImports := false
	for scanner.Scan() {
		if err = scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}

		line := strings.TrimSpace(scanner.Text())
		// single import
		if strings.HasPrefix(line, importone) {
			if match := importRegexSingle.FindStringSubmatch(line); match != nil {
				imports = append(imports, match[1])
				break
			}
		}

		// multiple imports
		if strings.HasPrefix(line, importmany) {
			readingImports = true
		}

		if readingImports {
			if strings.HasPrefix(line, ")") {
				break
			}
			if match := importRegexBlock.FindStringSubmatch(line); match != nil {
				imports = append(imports, match[1])
			}
		}
	}

	return imports, nil
}
