package common

import (
	"regexp"
	"strings"

	"paleotronic.com/core/settings"
)

var reExpr = regexp.MustCompile("(.+)(==|!=)(.+)")

func ResolveString(index int, in string) string {
	if !strings.HasPrefix(in, "@{") {
		return in
	}
	tmp := strings.TrimSuffix(strings.TrimPrefix(in, "@{"), "}")
	//log.Printf("condition is: %s", tmp)
	parts := strings.Split(tmp, ",")
	var condition, trues, falses string
	if len(parts) < 2 {
		return parts[0]
	}
	condition = parts[0]
	trues = parts[1]
	if len(parts) >= 3 {
		falses = parts[2]
	}
	//log.Printf("test: %s, if-true: %s, if-false: %s", condition, trues, falses)
	if !reExpr.MatchString(condition) {
		return trues
	}
	m := reExpr.FindAllStringSubmatch(condition, -1)
	varname := m[0][1]
	varvalue := m[0][3]
	test := m[0][2]

	//log.Printf("expression: var: %s, comparision: %s, value: %s", varname, test, varvalue)

	var actual string
	switch varname {
	case "diskii.sectors":
		if settings.DiskIIUse13Sectors[index] {
			actual = "13"
		} else {
			actual = "16"
		}
	default:
		return falses
	}

	//log.Printf("actual: %s", actual)

	switch test {
	case "==":
		if actual == varvalue {
			return trues
		}
		return falses
	case "!=":
		if actual != varvalue {
			return trues
		}
		return falses
	}
	return falses
}
