package http1

import "regexp"

const MICROSOFT_IIS_SERVER_REGEX_STRING = `(?i)(?:Microsoft.IIS(?:(?:\s|/)(\d+(?:\.\d){0,2})){0,1})`
const APACHE_SERVER_REGEX_STRING = `(?i)(?:Apache(?:(?:\s|/)(\d+(?:\.\d+){0,2}(?:-(?:M|B)\d)?)){0,1})`

const BASE_REGEX = `(?:(?:\s|/|-)(?:.*(?:\s|/|-))?(\d+(?:\.\d+){0,2})){0,1})`
const LIGHTHTTPD_SERVER_REGEX_STRING = `(?i)(?:Lighthttpd` + BASE_REGEX
const NGINX_SERVER_REGEX_STRING = `(?i)(?:nginx` + BASE_REGEX
const ATS_SERVER_REGEX_STRING = `(?i)(?:ATS` + BASE_REGEX
const BOA_SERVER_REGEX_STRING = `(?i)(?:boa(?:(?:\s|/|-)(?:.*(?:\s|/|-))?(\d+(?:\.\d+){0,2}(?:(?:rc)\d+)?)){0,1})`
const ALLEGRO_SOFTWARE_ROMPAGER_SERVER_REGEX_STRING = `(?i)(?:Allegro-Software-RomPager` + BASE_REGEX
const ALLEGRO_SERVE_SERVER_REGEX_STRING = `(?i)(?:AllegroServe` + BASE_REGEX
const SQUID_SERVER_REGEX_STRING = `(?i)(?:Squid(?:(?:\s|/|-)(\d+(?:\.\d+){0,2})){0,1})`
const TENGINE_SERVER_REGEX_STRING = `(?i)(?:Tengine(?:(?:\s|/|-)(\d+(?:\.\d+){0,2})){0,1})`
const JETTY_SERVER_REGEX_STRING = `(?i)(?:jetty(?:(?:\s|/|\(|-)(\d+(?:\.\d+){0,2}(\.rc\d)?)){0,1})`
const ROM_PAGER_SERVER_REGEX_STRING = `(?i)(?:RomPager(?:(?:\s|/|-)(\d+(?:\.\d+){0,2})){0,1})`
const MICRO_HTTPD_PAGER_SERVER_REGEX_STRING = `(?i)(?:micro_httpd(?:(?:\s|/|-)(\d+(?:\.\d+){0,2})){0,1})`
const MINI_HTTPD_PAGER_SERVER_REGEX_STRING = `(?i)(?:mini_httpd(?:(?:\s|/|-)(\d+(?:\.\d+){0,2})){0,1})`

const SERVER_FIELD_REGEXP_STRING = `(?:(?:\r\n)Server:\s(.*)\r\n)`

var microsoftIISRegex = regexp.MustCompile(MICROSOFT_IIS_SERVER_REGEX_STRING)
var apacheRegex = regexp.MustCompile(APACHE_SERVER_REGEX_STRING)
var nginxRegex = regexp.MustCompile(NGINX_SERVER_REGEX_STRING)
var lighthttpdRegex = regexp.MustCompile(LIGHTHTTPD_SERVER_REGEX_STRING)
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
var serverFieldRegexp = regexp.MustCompile(SERVER_FIELD_REGEXP_STRING)

var notCleaned = 0

func cleanAndAssign(agentString string, httpEntry *ZversionEntry) {
	IISMatch := microsoftIISRegex.FindStringSubmatch(agentString)
	if IISMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleIISServer(IISMatch))
		return
	}

	apacheMatch := apacheRegex.FindStringSubmatch(agentString)
	if apacheMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleApacheServer(apacheMatch))
		return
	}

	nginxMatch := nginxRegex.FindStringSubmatch(agentString)
	if nginxMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleNginxServer(nginxMatch))
		return
	}

	lighthttpdMatch := lighthttpdRegex.FindStringSubmatch(agentString)
	if lighthttpdMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleLighthttpdServer(lighthttpdMatch))
		return
	}

	atsMatch := atsRegex.FindStringSubmatch(agentString)
	if atsMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleATSServer(atsMatch))
		return
	}

	boaMatch := boaRegex.FindStringSubmatch(agentString)
	if boaMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleBOAServer(boaMatch))
		return
	}

	allegroSoftwareRomPagerMatch := allegroSoftwareRomPagerRegex.FindStringSubmatch(agentString)
	if allegroSoftwareRomPagerMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleAllegroSoftwareRomPagerServer(allegroSoftwareRomPagerMatch))
		return
	}

	allegroServeMatch := allegroServeRegex.FindStringSubmatch(agentString)
	if allegroServeMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleAllegroServeServer(allegroServeMatch))
		return
	}

	squidMatch := squidRegex.FindStringSubmatch(agentString)
	if squidMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleSquidServer(squidMatch))
		return
	}

	tengineMatch := tengineRegex.FindStringSubmatch(agentString)
	if tengineMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleTengineServer(tengineMatch))
		return
	}

	jettyMatch := jettyRegex.FindStringSubmatch(agentString)
	if jettyMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleJettyServer(jettyMatch))
		return
	}

	romPagerMatch := romPagerRegex.FindStringSubmatch(agentString)
	if romPagerMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleRomPagerServer(romPagerMatch))
		return
	}

	microHttpdMatch := microHttpdRegex.FindStringSubmatch(agentString)
	if microHttpdMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleMicroHttpdServer(microHttpdMatch))
		return
	}

	miniHttpdMatch := miniHttpdRegex.FindStringSubmatch(agentString)
	if miniHttpdMatch != nil {
		httpEntry.Agents = append(httpEntry.Agents, handleMiniHttpdServer(miniHttpdMatch))
		return
	}

	httpEntry.Agents = append(httpEntry.Agents, Server{Agent: agentString})
	notCleaned++
}

func handleMiniHttpdServer(serverString []string) Server {
	server := "mini_httpd"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}

func handleMicroHttpdServer(serverString []string) Server {
	server := "micro_httpd"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}

func handleRomPagerServer(serverString []string) Server {
	server := "RomPager"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}

func handleJettyServer(serverString []string) Server {
	server := "Jetty"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}

func handleTengineServer(serverString []string) Server {
	server := "Tengine"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}

func handleSquidServer(serverString []string) Server {
	server := "Squid"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}

func handleAllegroServeServer(serverString []string) Server {
	server := "AllegroServe"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}

func handleAllegroSoftwareRomPagerServer(serverString []string) Server {
	server := "Allegro-Software-RomPager"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}

func handleBOAServer(serverString []string) Server {
	server := "BOA"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}

func handleATSServer(serverString []string) Server {
	server := "ATS"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}

func handleLighthttpdServer(serverString []string) Server {
	server := "lighthttpd"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}

func handleIISServer(serverString []string) Server {
	server := "Microsoft-IIS"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}
func handleApacheServer(serverString []string) Server {
	server := "Apache"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}
func handleNginxServer(serverString []string) Server {
	server := "nginx"
	version := appendZero(serverString[1])

	return Server{Agent: server, Version: version}
}

func appendZero(version string) string {
	if len(version) == 1 {
		version = version + ".0"
	}

	return version
}
