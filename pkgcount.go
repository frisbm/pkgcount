package main

import (
	"flag"
	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/frisbm/pkgcount/utils"
	"log"
	"math"
	"os"
)

type Args struct {
	help       bool
	unrendered bool
	exclude    string
	lte        int
	gte        int
	out        string
	dir        string
}

func main() {
	var args Args

	flag.BoolVar(&args.help, "h", false, "Help")
	flag.BoolVar(&args.unrendered, "u", false, "Output markdown without being rendered")
	flag.StringVar(&args.out, "o", "", "Provide a name for a file output")
	flag.StringVar(&args.dir, "d", ".", "Directory to run")
	flag.StringVar(&args.exclude, "exclude", "", "Exclude files using a regular expression")
	flag.IntVar(&args.lte, "lte", math.MaxInt, "Return only packages with counts less than or equal to some number")
	flag.IntVar(&args.gte, "gte", 0, "Return only packages with counts greater than or equal to some number")

	flag.Parse()
	if args.help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	run(args)
}

func run(args Args) {
	moduleName, err := utils.GetModuleName(args.dir)
	if err != nil {
		log.Fatal(err)
	}
	pkgCounter := NewPackageCounter(args.dir, moduleName, args.exclude, args.lte, args.gte)

	err = pkgCounter.CountPackages()
	if err != nil {
		log.Fatal(err)
	}

	generatedMarkdown := pkgCounter.GenerateMarkdown()
	markdownBytes := []byte(generatedMarkdown)
	if !args.unrendered {
		markdownBytes = markdown.Render(generatedMarkdown, 80, 1)
	}

	if args.out == "" {
		_, err = os.Stdout.Write(markdownBytes)
		if err != nil {
			return
		}
	} else {
		err = os.WriteFile(args.out, markdownBytes, 0777)
		if err != nil {
			return
		}
	}
}
