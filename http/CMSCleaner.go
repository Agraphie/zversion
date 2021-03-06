package http

import (
	"github.com/agraphie/zversion/util"
	"regexp"
	"strings"
	"sync/atomic"
)

const CMS_EXTRACT_REGEX_META_TAG_STRING = `(?i)(?:meta name="(?:description|generator)" content="(Joomla|Wordpress|Drupal)!?.(\d+(?:\.\d+)*)?.*?")`
const CMS_EXTRACT_REGEX_HEADER_FIELD = `(Joomla|Wordpress|Drupal)!?.(\d+(?:\.\d+)*)?.*?`
const WORDPRESS_EXTRACT_FILE_PATH = `(href=.*(/wp-content/|/wp-includes/|/wp-json/))`

var cmsExtractFromMetaTagRegexp = regexp.MustCompile(CMS_EXTRACT_REGEX_META_TAG_STRING)
var cmsExtractFromHeaderFieldRegexp = regexp.MustCompile(CMS_EXTRACT_REGEX_HEADER_FIELD)
var wordpressExtractFromPath = regexp.MustCompile(WORDPRESS_EXTRACT_FILE_PATH)

var cmsCleaned uint64 = 0
var wpPath uint64 = 0
var wpMetaCleaned uint64 = 0
var wpXContendBy uint64 = 0

func cleanAndAssignCMS(rawBody string, xContentEncodedByField []string, httpEntry *ZversionEntry) {
	httpEntry.CMS = make([]CMS, 0)
	if len(xContentEncodedByField) > 0 {
		for _, v := range xContentEncodedByField {
			match := cmsExtractFromHeaderFieldRegexp.FindStringSubmatch(v)
			if match != nil {
				if strings.EqualFold(match[1], "WordPress") {
					atomic.AddUint64(&wpXContendBy, 1)
				}
				assignCMS(match, httpEntry)
				atomic.AddUint64(&cmsCleaned, 1)
			}
		}
	} else {
		match := cmsExtractFromMetaTagRegexp.FindAllStringSubmatch(rawBody, -1)
		if match != nil {
			for _, v := range match {
				if match != nil {
					if strings.EqualFold(v[1], "WordPress") {
						atomic.AddUint64(&wpMetaCleaned, 1)
					}
					assignCMS(v, httpEntry)
					atomic.AddUint64(&cmsCleaned, 1)
				}
			}
		} else {

			matchWordPress := wordpressExtractFromPath.FindAllStringSubmatch(rawBody, -1)
			if matchWordPress != nil {
				atomic.AddUint64(&wpPath, 1)
				newCMS := CMS{Vendor: "WordPress", Version: "", CanonicalVersion: ""}
				httpEntry.CMS = append(httpEntry.CMS, newCMS)
				atomic.AddUint64(&cmsCleaned, 1)

			}
		}
	}
}

func assignCMS(match []string, httpEntry *ZversionEntry) {
	var vendor string
	if strings.EqualFold(match[1], "WordPress") {
		vendor = "WordPress"
	} else if strings.EqualFold(match[1], "Joomla") {
		vendor = "Joomla"
	} else if strings.EqualFold(match[1], "Drupal") {
		vendor = "Drupal"
	}
	version := match[2]
	canonicalVersion := ""
	if version != "" {
		version = util.AppendZeroToVersion(version)
		canonicalVersion = util.MakeVersionCanonical(version)
	}

	newCMS := CMS{Vendor: vendor, Version: version, CanonicalVersion: canonicalVersion}
	for _, v := range httpEntry.CMS {
		if v == newCMS {
			return
		}
	}
	httpEntry.CMS = append(httpEntry.CMS, newCMS)
}
