package common

import "regexp"

var Alphanum *regexp.Regexp
var SingleChar *regexp.Regexp

func init() {
	var err error
	Alphanum, err = regexp.Compile("[a-zA-Z0-9]")
	if err != nil {
		panic(err)
	}
	SingleChar, err = regexp.Compile("^.{1,1}$")
	if err != nil {
		panic(err)
	}
}
