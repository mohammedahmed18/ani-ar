// Unpacker for Dean Edward's p.a.c.k.e.r
// Ported from: https://github.com/beautifier/js-beautify/blob/main/python/jsbeautifier/unpackers/packer.py
package egydead

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type DEUnpacker struct {
	Source string
	Prefix string
	Suffix string
}

func NewDEUnpacker(source string) (*DEUnpacker, bool) {
	deu := &DEUnpacker{
		Source: source,
	}
	beginOffset := -1
	indexes := regexp.MustCompile(`eval[ ]*\([ ]*function[ ]*\([ ]*p[ ]*,[ ]*a[ ]*,[ ]*c[ ]*,[ ]*k[ ]*,[ ]*e[ ]*,[ ]*`).FindStringIndex(source)
	if indexes != nil {
		beginOffset = indexes[0]
		deu.Prefix = source[:beginOffset]
	}
	if beginOffset == -1 {
		return nil, false
	}
	sourceEnd := source[beginOffset:]
	if arr := strings.SplitN(sourceEnd, "')))", 2); arr[0] == sourceEnd {
		if arr2 := strings.SplitN(sourceEnd, "}))", 2); len(arr2) > 1 {
			deu.Suffix = arr2[1]
		}
	} else {
		deu.Suffix = arr[1]
	}
	return deu, true
}

func filterArgs(source string) (string, []string, int, int, error) {
	juicers := []string{
		`(?s)}\('(.*)', *(\d+|\[\]), *(\d+), *'(.*)'\.split\('\|'\), *(\d+), *(.*)\)\)`,
		`(?s)}\('(.*)', *(\d+|\[\]), *(\d+), *'(.*)'\.split\('\|'\)`,
	}
	for _, juicer := range juicers {
		args := regexp.MustCompile(juicer).FindStringSubmatch(source)
		if args != nil {
			if args[2] == "[]" {
				args[2] = "62"
			}
			payload := args[1]
			symtab := strings.Split(args[4], "|")
			base, _ := strconv.Atoi(args[2])
			count, _ := strconv.Atoi(args[3])
			return payload, symtab, base, count, nil
		}
	}
	return "", nil, 0, 0, errors.New("unexpected code structure")
}

func replaceStrings(source, prefix, suffix string) string {
	match := regexp.MustCompile(`(?s)var *(_\w+)\=\["(.*?)"\];`).FindStringSubmatch(source)
	if match == nil {
		return prefix + source + suffix
	}
	startpoint := len(match[0])
	lookup := strings.Split(match[2], `","`)
	variable := match[1] + `[%d]`
	for i, value := range lookup {
		source = strings.ReplaceAll(source, fmt.Sprintf(variable, i), `"`+value+`"`)
	}
	return source[startpoint:]
}

func (deu *DEUnpacker) Unpack() (string, error) {
	payload, symtab, base, count, err := filterArgs(deu.Source)
	if err != nil {
		return "", err
	}
	if count != len(symtab) {
		return "", errors.New("malformed p.a.c.k.e.r. symtab")
	}
	ub, err := newUnbaser(base)
	if err != nil {
		return "", err
	}
	payload = strings.ReplaceAll(payload, `\\`, `\`)
	payload = strings.ReplaceAll(payload, `\'`, `'`)
	re := regexp.MustCompile(`\b\w+\b`)
	deu.Source = re.ReplaceAllStringFunc(payload, func(word string) string {
		idx := ub.unbase(word)
		if idx >= len(symtab) || symtab[idx] == "" {
			return word
		}
		return symtab[idx]
	})
	return replaceStrings(deu.Source, deu.Prefix, deu.Suffix), nil
}

type unbaser struct {
	base    int
	baseTen map[rune]int
}

func newUnbaser(base int) (*unbaser, error) {
	var alphabet string
	if 2 <= base && base <= 62 {
		alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"[:base]
	} else if base == 95 {
		alphabet = ` !"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_abcdefghijklmnopqrstuvwxyz{|}~`
	} else {
		return nil, errors.New("unsupported base encoding")
	}
	ub := &unbaser{
		base:    base,
		baseTen: make(map[rune]int, len(alphabet)),
	}
	for i, c := range alphabet {
		ub.baseTen[c] = i
	}
	return ub, nil
}

func (ub *unbaser) unbase(num string) int {
	n := 0
	for _, digit := range num {
		n = (n * ub.base) + ub.baseTen[digit]
	}
	return n
}
