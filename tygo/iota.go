package tygo

import (
	"regexp"
	"strconv"
	"strings"
)

var iotaRegexp = regexp.MustCompile(`\biota\b`)

func isProbablyIotaType(valueString string) bool {
	return !strings.ContainsAny(valueString, "\"'`") && iotaRegexp.MatchString(valueString)
}

func replaceIotaValue(valueString string, iotaValue int) string {
	return iotaRegexp.ReplaceAllLiteralString(valueString, strconv.Itoa(iotaValue))
}
