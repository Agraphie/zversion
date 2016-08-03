package ssh

import (
	"github.com/agraphie/zversion/util"
	"regexp"
	"sync/atomic"
)

const ONLY_NUMBERS_REGEXP_STRING = `(?:\.\d+)+`
const OPENSSH_REGEXP_STRING = `(?i)(?:(?:OpenSSH)(?:(?:\s|/|_)(\d+(?:\.\d+){0,2}(?:p\d)?))?)`
const ROMSSHELL_REGEXP_STRING = `(?i)(?:(?:RomSShell)(?:(?:\s|/|_)(\d+(?:\.\d+){0,2}))?)`
const DROPBEAR_REGEXP_STRING = `(?i)(?:(?:dropbear)(?:(?:\s|/|_)(\d+(?:\.\d+){0,2}))?)`
const SFTP_REGEXP_STRING = `(?i)(?:(?:mod_sftp)(?:(?:\s|/|_)(\d+(?:\.\d+){0,2}))?)`
const COREFTP_REGEXP_STRING = `(?i)(?:(?:CoreFTP)(?:(?:\s|/|_)(\d+(?:\.\d+){0,2}))?)`
const APPGATE_REGEXP_STRING = `(?i)(?:(?:AppGateSSH)(?:(?:\s|/|_)(\d+(?:\.\d+){0,2}))?)`
const TECTIA_SERVER_REGEXP_STRING = `(?i)((\d+\.?)+)(?:\sSSH\sTectia\sServer)`
const SSHLIB_REGEXP_STRING = `(?i)((\d+\.?)+)(?:.sshlib)`

var onlyNumbersRegex = regexp.MustCompile(ONLY_NUMBERS_REGEXP_STRING)
var openSSHRegex = regexp.MustCompile(OPENSSH_REGEXP_STRING)
var romSShellRegex = regexp.MustCompile(ROMSSHELL_REGEXP_STRING)
var dropbearRegx = regexp.MustCompile(DROPBEAR_REGEXP_STRING)
var sftpRegx = regexp.MustCompile(SFTP_REGEXP_STRING)
var coreFTPRegx = regexp.MustCompile(COREFTP_REGEXP_STRING)
var appGateSSHRegx = regexp.MustCompile(APPGATE_REGEXP_STRING)
var tectiaServerRegex = regexp.MustCompile(TECTIA_SERVER_REGEXP_STRING)
var sshLibRegex = regexp.MustCompile(SSHLIB_REGEXP_STRING)

var m map[string]*regexp.Regexp = map[string]*regexp.Regexp{
	"OpenSSH":           openSSHRegex,
	"RomSShell":         romSShellRegex,
	"dropbear":          dropbearRegx,
	"sftp":              sftpRegx,
	"CoreFTP":           coreFTPRegx,
	"AppGateSSH":        appGateSSHRegx,
	"SSH Tectia Server": tectiaServerRegex,
	"sshlib":            sshLibRegex,
}

var softwareBannerNotCleaned uint64 = 0
var softwareBannerCleaned uint64 = 0

func cleanAndAssign(software string, sshEntry *SSHEntry) {
	numbersMatch := onlyNumbersRegex.FindStringSubmatch(software)
	if numbersMatch != nil {
		software = sshEntry.Raw_banner
	}
	for k, v := range m {
		match := v.FindStringSubmatch(software)
		if match != nil {
			version := util.AppendZeroToVersion(match[1])
			sshEntry.Vendor = k
			sshEntry.CanonicalVersion = util.MakeVersionCanonical(version)
			sshEntry.SoftwareVersion = version
			atomic.AddUint64(&softwareBannerCleaned, 1)

			return
		}
	}
	sshEntry.Vendor = software
	atomic.AddUint64(&softwareBannerNotCleaned, 1)
}
