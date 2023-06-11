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

	flag.BoolVar(&args.help, "h", false, "Display help messages with argument list and descriptions.")
	flag.BoolVar(&args.unrendered, "u", false, "Retrieve markdown in unrendered format.")
	flag.StringVar(&args.out, "o", "", "Save output to file. Used with -u to preserve format.")
	flag.StringVar(&args.dir, "d", ".", "Specify directory or file path. Default is current directory.")
	flag.StringVar(&args.exclude, "exclude", "", "Exclude specific files, directories, or other entities with regex.")
	flag.IntVar(&args.lte, "lte", 0, "Display package counts less than or equal to specified integer.")
	flag.IntVar(&args.gte, "gte", 0, "Display package counts greater than or equal to specified integer.")

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
	if args.lte == 0 {
		args.lte = math.MaxInt
	}
	pkgCounter := NewPackageCounter(args.dir, moduleName, args.exclude, args.lte, args.gte)

	err = pkgCounter.CountPackages()
	if err != nil {
		log.Fatal(err)
	}

	generatedMarkdown, err := pkgCounter.GenerateMarkdown()
	if err != nil {
		log.Fatal(err)
	}

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
