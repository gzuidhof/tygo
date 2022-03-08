package tygo

import (
	"log"
	"strconv"
	"strings"
)

func isProbablyIotaType(groupType string) bool {
	groupType = strings.Trim(groupType, "()")
	return groupType == "iota" || strings.HasPrefix(groupType, "iota +") || strings.HasSuffix(groupType, "+ iota")
}

// Note: this could be done so much more elegantly, but this probably covers 99.9% of iota usecases
func basicIotaOffsetValueParse(groupType string) int {
	if !isProbablyIotaType(groupType) {
		panic("can't parse non-iota type")
	}
	groupType = strings.Trim(groupType, "()")
	if groupType == "iota" {
		return 0
	}
	parts := strings.Split(groupType, " + ")

	var numPart string
	if parts[0] == "iota" {
		numPart = parts[1]
	} else {
		numPart = parts[0]
	}

	addValue, err := strconv.ParseInt(numPart, 10, 64)
	if err != nil {
		log.Panicf("Failed to guesstimate initial iota value for \"%s\": %v", groupType, err)
	}
	return int(addValue)
}
