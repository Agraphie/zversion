package http1

import (
	"github.com/agraphie/zversion/util"
	"regexp"
	"sync/atomic"
)

const MICROSOFT_IIS_SERVER_REGEX_STRING = `(?i)(?:(?:Microsoft.)?IIS(?:(?:\s|/)(\d+(?:\.\d)(?:\.[1-])?))?)`
const APACHE_SERVER_REGEX_STRING = `(?i)(?:Apache(?:(?:\s|/)(\d+(?:\.\d+){0,2}(?:-(?:M|B)\d)?))?)`

const BASE_REGEX = `(?:(?:\s|/|-)(?:.*(?:\s|/|-))?(\d+(?:\.\d+){0,2}))?)`
const LIGHTHTTPD_SERVER_REGEX_STRING = `(?i)(?:lighttpd(?:(?:\s|/|-)(\d+(?:\.\d+){0,2}))?)`
const NGINX_CLOUDFLARE_SERVER_REGEX_STRING = `(?i)(?:cloudflare-nginx` + BASE_REGEX
const NGINX_SERVER_REGEX_STRING = `(?i)(?:nginx` + BASE_REGEX
const ATS_SERVER_REGEX_STRING = `(?i)(?:ATS` + BASE_REGEX
const BOA_SERVER_REGEX_STRING = `(?i)(?:boa(?:(?:\s|/|-)(?:.*(?:\s|/|-))?(\d+(?:\.\d+){0,2}(?:(?:rc)\d+)?))?)`
const ALLEGRO_SOFTWARE_ROMPAGER_SERVER_REGEX_STRING = `(?i)(?:Allegro-Software-RomPager` + BASE_REGEX
const ALLEGRO_SERVE_SERVER_REGEX_STRING = `(?i)(?:AllegroServe` + BASE_REGEX
const SQUID_SERVER_REGEX_STRING = `(?i)(?:Squid(?:(?:\s|/|-)(\d+(?:\.\d+){0,2}))?)`
const TENGINE_SERVER_REGEX_STRING = `(?i)(?:Tengine(?:(?:\s|/|-)(\d+(?:\.\d+){0,2}))?)`
const JETTY_SERVER_REGEX_STRING = `(?i)(?:jetty(?:(?:\s|/|\(|-)(\d+(?:\.\d+){0,2}(\.rc\d)?))?)`
const ROM_PAGER_SERVER_REGEX_STRING = `(?i)(?:RomPager(?:(?:\s|/|-)(\d+(?:\.\d+){0,2}))?)`
const MICRO_HTTPD_PAGER_SERVER_REGEX_STRING = `(?i)(?:micro_httpd(?:(?:\s|/|-)(\d+(?:\.\d+){0,2}))?)`
const MINI_HTTPD_PAGER_SERVER_REGEX_STRING = `(?i)(?:mini_httpd(?:(?:\s|/|-)(\d+(?:\.\d+){0,2}))?)`
const AOL_SERVER_REGEX_STRING = `(?i)(?:AOLserver` + BASE_REGEX
const ABYSS_SERVER_REGEX_STRING = `(?i)(?:Abyss(?:(?:\s|/|-)(\d+(?:\.\d+){0,3}(?:-X\d)?)))`
const AGRANAT_SERVER_REGEX_STRING = `(?i)(?:Agranat-EmWeb` + BASE_REGEX
const MICROSOFT_HTTPAPI_SERVER_REGEX_STRING = `(?i)(?:Microsoft-HTTPAPI` + BASE_REGEX
const CHERRYPY_SERVER_REGEX_STRING = `(?i)(?:CherryPy` + BASE_REGEX
const CHEROKEE_SERVER_REGEX_STRING = `(?i)(?:Cherokee` + BASE_REGEX
const COMMUNIGATE_SERVER_REGEX_STRING = `(?i)(?:CommuniGatePro` + BASE_REGEX
const EDGEPRISM_SERVER_REGEX_STRING = `(?i)(?:EdgePrism(?:(?:\s|/|-)(?:.*(?:\s|/|-))?(\d+(?:\.\d+){0,3}))?)`
const FLYWHEEL_SERVER_REGEX_STRING = `(?i)(?:Flywheel(?:(?:\s|/|-)(?:.*(?:\s|/|-))?(\d+(?:\.\d+){0,3}))?)`

//const GLASSFISH_SERVER_REGEX_STRING = `(?i)(?:GlassFish(?:(?:\s|/|-)(?:.*(?:\s|/|-))?(\d+(?:\.\d+){0,3}))?)`
const GOAHEAD_SERVER_REGEX_STRING = `(?i)(?:GoAhead(?:(?:\s|/|-)(?:.*?(?:\s|/|-))?(\d+(?:\.\d+){0,2}))?)`
const IDEA_WEB_SERVER_SERVER_REGEX_STRING = `(?i)(?:IdeaWebServer(?:(?:\s|/)v(\d+(?:\.\d+){0,2}))?)`
const INDY_SERVER_REGEX_STRING = `(?i)(?:Indy(?:(?:\s|/)(\d+(?:\.\d+){0,2}))?)`
const MBEDTHIS_SERVER_REGEX_STRING = `(?i)(?:Mbedthis` + BASE_REGEX
const PRTG_SERVER_REGEX_STRING = `(?i)(?:PRTG(?:(?:\s|/|-)(?:.*(?:\s|/|-))?(\d+(?:\.\d+){0,3}))?)`
const KANGLE_SERVER_REGEX_STRING = `(?i)(?:Kangle(?:(?:\s|/|-)(?:.*(?:\s|/|-))?(\d+(?:\.\d+){0,3}))?)`
const THTTPD_SERVER_REGEX_STRING = `(?i)(?:thttpd(?:(?:\s|/|-)(\d+(?:\.\d+){0,2}))?)`

const SERVER_FIELD_REGEXP_STRING = `(?:(?:\r\n)Server:\s(.*)\r\n)`

var microsoftIISRegex = regexp.MustCompile(MICROSOFT_IIS_SERVER_REGEX_STRING)
var apacheRegex = regexp.MustCompile(APACHE_SERVER_REGEX_STRING)
var nginxCloudflareRegex = regexp.MustCompile(NGINX_CLOUDFLARE_SERVER_REGEX_STRING)
var nginxRegex = regexp.MustCompile(NGINX_SERVER_REGEX_STRING)
var lighttpdRegex = regexp.MustCompile(LIGHTHTTPD_SERVER_REGEX_STRING)
var atsRegex = regexp.MustCompile(ATS_SERVER_REGEX_STRING)
var boaRegex = regexp.MustCompile(BOA_SERVER_REGEX_STRING)
var allegroSoftwareRomPagerRegex = regexp.MustCompile(ALLEGRO_SOFTWARE_ROMPAGER_SERVER_REGEX_STRING)
var allegroServeRegex = regexp.MustCompile(ALLEGRO_SERVE_SERVER_REGEX_STRING)
var squidRegex = regexp.MustCompile(SQUID_SERVER_REGEX_STRING)
var tengineRegex = regexp.MustCompile(TENGINE_SERVER_REGEX_STRING)
var jettyRegex = regexp.MustCompile(JETTY_SERVER_REGEX_STRING)
var romPagerRegex = regexp.MustCompile(ROM_PAGER_SERVER_REGEX_STRING)
var microHttpdRegex = regexp.MustCompile(MICRO_HTTPD_PAGER_SERVER_REGEX_STRING)
var miniHttpdRegex = regexp.MustCompile(MINI_HTTPD_PAGER_SERVER_REGEX_STRING)
var aolServerRegex = regexp.MustCompile(AOL_SERVER_REGEX_STRING)
var abyssServerRegex = regexp.MustCompile(ABYSS_SERVER_REGEX_STRING)
var agranatServerRegex = regexp.MustCompile(AGRANAT_SERVER_REGEX_STRING)
var microsoftHttpApiRegex = regexp.MustCompile(MICROSOFT_HTTPAPI_SERVER_REGEX_STRING)
var cherryPyRegex = regexp.MustCompile(CHERRYPY_SERVER_REGEX_STRING)
var cherokeeRegex = regexp.MustCompile(CHEROKEE_SERVER_REGEX_STRING)
var communiGateProRegex = regexp.MustCompile(COMMUNIGATE_SERVER_REGEX_STRING)
var edgePrismRegex = regexp.MustCompile(EDGEPRISM_SERVER_REGEX_STRING)
var flywheelRegex = regexp.MustCompile(FLYWHEEL_SERVER_REGEX_STRING)

//var glassfishRegex = regexp.MustCompile(GLASSFISH_SERVER_REGEX_STRING)
var goaheadRegex = regexp.MustCompile(GOAHEAD_SERVER_REGEX_STRING)
var ideaWebServerRegex = regexp.MustCompile(IDEA_WEB_SERVER_SERVER_REGEX_STRING)
var IndyRegex = regexp.MustCompile(INDY_SERVER_REGEX_STRING)
var mbedthisRegex = regexp.MustCompile(MBEDTHIS_SERVER_REGEX_STRING)
var prtgRegex = regexp.MustCompile(PRTG_SERVER_REGEX_STRING)
var kangleRegex = regexp.MustCompile(KANGLE_SERVER_REGEX_STRING)
var thttpdRegex = regexp.MustCompile(THTTPD_SERVER_REGEX_STRING)

var m map[string]*regexp.Regexp = map[string]*regexp.Regexp{
	"Microsoft-IIS":    microsoftIISRegex,
	"Apache":           apacheRegex,
	"cloudflare-nginx": nginxCloudflareRegex,
	"nginx":            nginxRegex,
	"lighttpd":         lighttpdRegex,
	"ATS":              atsRegex,
	"BOA":              boaRegex,
	"Allegro Software RomPager": allegroSoftwareRomPagerRegex,
	"AllegroServe":              allegroServeRegex,
	"squid":                     squidRegex,
	"tengine":                   tengineRegex,
	"jetty":                     jettyRegex,
	"RomPager":                  romPagerRegex,
	"mini_httpd":                miniHttpdRegex,
	"micro_httpd":               microHttpdRegex,
	"AOL Server":                aolServerRegex,
	"Abyss":                     abyssServerRegex,
	"Agranat-EmWeb":             agranatServerRegex,
	"Microsoft-HTTPAPI":         microsoftHttpApiRegex,
	"CherryPy":                  cherryPyRegex,
	"Cherokee":                  cherokeeRegex,
	"CommuniGatePro":            communiGateProRegex,
	"EdgePrism":                 edgePrismRegex,
	"Flywheel":                  flywheelRegex,
	"GoAhead":                   goaheadRegex,
	"IdeaWebServer":             ideaWebServerRegex,
	"Indy":                      IndyRegex,
	"mbedthis":                  mbedthisRegex,
	"PRTG":                      prtgRegex,
	"Kangle":                    kangleRegex,
	"thttpd":                    thttpdRegex,
}

var serverFieldRegexp = regexp.MustCompile(SERVER_FIELD_REGEXP_STRING)

var serverHeaderNotCleaned uint64 = 0
var serverHeaderCleaned uint64 = 0

func cleanAndAssign(agentString string, httpEntry *ZversionEntry) {
	for k, v := range m {
		match := v.FindStringSubmatch(agentString)
		if k == "nginx" && nginxCloudflareRegex.FindStringSubmatch(agentString) != nil {
			k = "cloudflare-nginx"
		}
		if match != nil {
			version := util.AppendZeroToVersion(match[1])
			canonicalVersion := util.MakeVersionCanonical(version)
			httpEntry.Agents = append(httpEntry.Agents, Server{Vendor: k, Version: version, CanonicalVersion: canonicalVersion})
			atomic.AddUint64(&serverHeaderCleaned, 1)

			return
		}
	}
	httpEntry.Agents = append(httpEntry.Agents, Server{Vendor: agentString})
	atomic.AddUint64(&serverHeaderNotCleaned, 1)

}
