package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	isatty "github.com/mattn/go-isatty"
	lsd "github.com/mattn/go-lsd"
	unicodeclass "github.com/mattn/go-unicodeclass"
)

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

		words := unicodeclass.Split(strings.ToLower(srcscan.Text()))
		scan := bufio.NewScanner(in)
		lno := 0
		result := make([][][4][]string, distance+1)
		for scan.Scan() {
			lno++
			linewords := unicodeclass.Split(scan.Text())
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

				if found == len(words) {
					lnostr := []string{}
					lnostr = append(lnostr, strconv.Itoa(lno))
					result[max] = append(result[max], [4][]string{
						linewords[:i],
						linewords[i : i+found],
						linewords[i+found:],
						lnostr,
					})
					break
				}
			}
		}

		for i := 1; i <= distance; i++ {
			if 0 < len(result[i]) && len(result[i]) < 20 {
				if isatty.IsTerminal(os.Stdout.Fd()) {
					fmt.Fprintf(out, "\nIN: %s: L.%d:%s\n",
						srcfile,
						srclno,
						color.GreenString(srcscan.Text()))
				} else {
					fmt.Fprintf(out, "\nIN: %s: L.%d:%s\n",
						srcfile,
						srclno,
						srcscan.Text())
				}

				for j := 0; j < len(result[i]); j++ {
					left := strings.Join(result[i][j][0], "")
					middle := strings.Join(result[i][j][1], "")
					right := strings.Join(result[i][j][2], "")

					if isatty.IsTerminal(os.Stdout.Fd()) {
						fmt.Fprintf(out, "%s (d=%d): L.%s:%s\n",
							file,
							i,
							result[i][j][3][0],
							left+color.RedString(middle)+right)
					} else {
						fmt.Fprintf(out, "%s (d=%d): L.%s:%s\n",
							file,
							i,
							result[i][j][3][0],
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
