package http1

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
