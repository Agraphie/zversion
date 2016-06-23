package http1

import (
	"github.com/agraphie/zversion/util"
	"regexp"
	"strings"
)

const CMS_EXTRACT_REGEX_META_TAG_STRING = `(?i)(?:meta name="(?:description|generator)" content="(Joomla|Wordpress|Drupal)!?.(\d+(?:\.\d+)*)?.*?")`
const CMS_EXTRACT_REGEX_HEADER_FIELD = `(Joomla|Wordpress|Drupal)!?.(\d+(?:\.\d+)*)?.*?`

var cmsExtractFromMetaTagRegexp = regexp.MustCompile(CMS_EXTRACT_REGEX_META_TAG_STRING)
var cmsExtractFromHeaderFieldRegexp = regexp.MustCompile(CMS_EXTRACT_REGEX_HEADER_FIELD)

func cleanAndAssignCMS(rawBody string, xContentEncodedByField []string, httpEntry *ZversionEntry) {
	httpEntry.CMS = make([]CMS, 0)
	if len(xContentEncodedByField) > 0 {
		for _, v := range xContentEncodedByField {
			match := cmsExtractFromHeaderFieldRegexp.FindStringSubmatch(v)
			if match != nil {
				assignCMS(match, httpEntry)
			}
		}
	} else {
		match := cmsExtractFromMetaTagRegexp.FindAllStringSubmatch(rawBody, -1)
		for _, v := range match {
			if match != nil {
				assignCMS(v, httpEntry)
			}
		}
	}
}

func assignCMS(match []string, httpEntry *ZversionEntry) {
	vendor := strings.Title(match[1])
	version := match[2]
	canonicalVersion := ""
	if version != "" {
		version = util.AppendZeroToVersion(version)
		canonicalVersion = util.MakeVersionCanonical(version)
	}
	httpEntry.CMS = append(httpEntry.CMS, CMS{Vendor: vendor, Version: version, CanonicalVersion: canonicalVersion})
}
