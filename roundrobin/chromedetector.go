package roundrobin

import (
	"strconv"
	"strings"
)

// IsUANewChrome does a dirty check of a user agent string and returns true if the user agent is
// _probably_ Chrome version 77 or greater
func IsUANewChrome(uaHeader string) bool {
	if chromeStrIdx := strings.LastIndex(uaHeader, "Chrome/"); chromeStrIdx != -1 {
		chromeSuffix := uaHeader[chromeStrIdx+7:]
		if majorVerSep := strings.Index(chromeSuffix, "."); majorVerSep != -1 {
			chromeMajorVer, error := strconv.Atoi(chromeSuffix[:majorVerSep])
			if error == nil && chromeMajorVer > 77 {
				return true
			}
		}
	}
	return false
}
