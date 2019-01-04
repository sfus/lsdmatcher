package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	isatty "github.com/mattn/go-isatty"
	lsd "github.com/mattn/go-lsd"
	unicodeclass "github.com/mattn/go-unicodeclass"
)

type Match struct {
	Left   []string
	Middle []string
	Right  []string
	LineNo int
	LSD    int
	NLSD   float64 // Normalized Levenshtein Distance
}

func main() {
	var distance int
	flag.IntVar(&distance, "d", 2, "distance")
	flag.Parse()

	if flag.NArg() == 0 || flag.NArg() > 2 {
		flag.Usage()
		os.Exit(2)
	}

	var out io.Writer
	var src, in io.Reader
	var srcfile, file string

	if isatty.IsTerminal(os.Stdout.Fd()) {
		out = colorable.NewColorableStdout()
	} else {
		out = os.Stdout
	}

	srcfile = flag.Arg(0)

	s, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		os.Exit(1)
	}
	defer s.Close()
	src = s

	srcscan := bufio.NewScanner(src)
	srclno := 0
	for srcscan.Scan() {

		srclno++
		if flag.NArg() == 1 {
			file = "stdin"
			in = os.Stdin
		} else {
			file = flag.Arg(1)
			var err error
			f, err := os.Open(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
				os.Exit(1)
			}
			defer f.Close()
			in = f
		}

		srcline := srcscan.Text()
		words := unicodeclass.Split(strings.ToLower(srcline))
		scan := bufio.NewScanner(in)
		lno := 0
		result := make([][]Match, distance+1)
		for scan.Scan() {
			lno++
			line := scan.Text()
			linewords := unicodeclass.Split(line)
			for i := 0; i < len(linewords); i++ {
				if i+len(words) >= len(linewords) {
					break
				}
				found := 0
				max := 0
				for j := 0; j < len(words); j++ {
					d := lsd.StringDistance(words[j], strings.ToLower(linewords[i+found]))
					if d > distance {
						break
					} else if max < d {
						max = d
					}
					found++
				}
				longer := srcline
				if longer < line {
					longer = line
				}

				if found == len(words) {
					result[max] = append(result[max], Match{
						Left:   linewords[:i],
						Middle: linewords[i : i+found],
						Right:  linewords[i+found:],
						LineNo: lno,
						LSD:    max,
						NLSD:   float64(max) / float64(len(longer)),
					})
					break
				}
			}
		}

		for i := 1; i <= distance; i++ {
			if 0 < len(result[i]) {
				sort.Slice(result[i], func(j, k int) bool { return result[i][j].NLSD < result[i][k].NLSD })

				if isatty.IsTerminal(os.Stdout.Fd()) {
					fmt.Fprintf(out, "\nIN: %s: L.%d:%s\n",
						srcfile,
						srclno,
						color.GreenString(srcline))
				} else {
					fmt.Fprintf(out, "\nIN: %s: L.%d:%s\n",
						srcfile,
						srclno,
						srcscan.Text())
				}

				for j := 0; j < len(result[i]); j++ {
					left := strings.Join(result[i][j].Left, "")
					middle := strings.Join(result[i][j].Middle, "")
					right := strings.Join(result[i][j].Right, "")

					if isatty.IsTerminal(os.Stdout.Fd()) {
						fmt.Fprintf(out, "%s (d=%d, n=%f): L.%d:%s\n",
							file,
							i,
							result[i][j].NLSD,
							result[i][j].LineNo,
							left+color.RedString(middle)+right)
					} else {
						fmt.Fprintf(out, "%s (d=%d, n=%f): L.%d:%s\n",
							file,
							i,
							result[i][j].NLSD,
							result[i][j].LineNo,
							left+middle+right)
					}
				}
				break
			}
		}

		if err := scan.Err(); err != nil {
			log.Fatal(err)
		}

	}

}
