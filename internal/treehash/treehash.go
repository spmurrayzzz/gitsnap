package treehash

import "regexp"

type Hash string

var valid = regexp.MustCompile(`^[0-9a-f]{40}$`)

func Parse(s string) (Hash, bool) {
	if !valid.MatchString(s) {
		return "", false
	}
	return Hash(s), true
}
