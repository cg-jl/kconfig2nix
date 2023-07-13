package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"
)

type Option struct {
	Name    string
	NixExpr string
}

func parseOption(opt string) Option {
	if opt[0] == '#' {
		space := strings.IndexRune(opt[2:], ' ') + 2
		return Option{Name: opt[2+7 : space], NixExpr: "no"}
	} else {
		eql := strings.IndexRune(opt, '=')
		value := opt[eql+1:]
		var result string = "freeform "  + value
		switch value {
		case "m":
			result = "module"
		case "y":
			result = "yes"

		default:
			break
		}
		return Option{Name: opt[7:eql], NixExpr: result}
	}
}

func parseFile(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	sc := bufio.NewScanner(file)

	res := make(map[string]string, 9000)

	for sc.Scan() {
		line := sc.Text()
		if !strings.Contains(line, "CONFIG") {
			continue
		}
		option := parseOption(line)
		res[option.Name] = option.NixExpr
	}

	if err = sc.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func mergeMaps(oldmap, newmap map[string]string) map[string]string {
	res := make(map[string]string, int(math.Max(float64(len(oldmap)), float64(len(newmap)))))

	for k, newv := range newmap {
		if oldv, ok := oldmap[k]; ok {
			if oldv != newv {
				res[k] = newv
			}
		}
	}

	return res
}


func main() {

    if len(os.Args) != 4 {
        fmt.Printf("usage: %s <new file> <nixos file> <overlay package name>", os.Args[0]);
        os.Exit(1)
    }

	newfile := os.Args[1]
	origfile := os.Args[2]
    name := os.Args[3]

	newmap, err := parseFile(newfile)
	if err != nil {
		panic(err)
	}

	origmap, err := parseFile(origfile)
	if err != nil {
		panic(err)
	}

	merged := mergeMaps(origmap, newmap)

    fmt.Println("{pkgs, lib, ...}: {")
    fmt.Printf("\t%q = pkgs.linuxPackagesFor (pkgs.linux_6_1_3.override {\n", name)
    fmt.Println("\t\tstructuredExtraConfig = with lib.kernel; {")

    for k, v := range merged {
        fmt.Printf("\t\t\t%s = %s;\n", k, v)
    }

    fmt.Println("\t\t};")
    fmt.Println("\t});")
    fmt.Println("}")

}
