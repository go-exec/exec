package exec

import (
	"regexp"
	"strings"
)

func mergeOptions(maps ...map[string]*Option) (output map[string]*Option) {
	size := len(maps)
	if size == 0 {
		return output
	}
	if size == 1 {
		return maps[0]
	}
	output = make(map[string]*Option)
	for _, m := range maps {
		for k, v := range m {
			output[k] = v
		}
	}
	return output
}

func Parse(text string) string {
	re := regexp.MustCompile(`\{\{\s*([\w\.\/]+)\s*\}\}`)
	if !re.MatchString(text) {
		return text
	}
	return re.ReplaceAllStringFunc(text, func(str string) string {
		name := strings.TrimRight(strings.TrimLeft(str, "{{"), "}}")
		if Has(name) {
			return Parse(Get(name).String())
		} else {
			return str
		}
	})
}
