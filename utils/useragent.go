package utils

import (
	"strings"
)

func ParseUserAgent(ua string) string {
	if ua == "" {
		return "unknown"
	}

	lower := strings.ToLower(ua)

	browser := "unknown"
	version := ""

	switch {
	case strings.Contains(lower, "firefox"):
		browser = "Firefox"
	case strings.Contains(lower, "chrome"):
		browser = "Chrome"
	case strings.Contains(lower, "safari") && !strings.Contains(lower, "chrome"):
		browser = "Safari"
	case strings.Contains(lower, "edge"):
		browser = "Edge"
	}

	// naive version extraction (good enough for admin view)
	if i := strings.Index(lower, "firefox/"); i != -1 {
		version = ua[i+8:]
	} else if i := strings.Index(lower, "chrome/"); i != -1 {
		version = ua[i+7:]
	} else if i := strings.Index(lower, "safari/"); i != -1 {
		version = ua[i+7:]
	}

	os := ""
	switch {
	case strings.Contains(lower, "windows"):
		os = "Windows"
	case strings.Contains(lower, "linux"):
		os = "Linux"
	case strings.Contains(lower, "mac"):
		os = "macOS"
	case strings.Contains(lower, "android"):
		os = "Android"
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "ios"):
		os = "iOS"
	}

	if os != "" && version != "" {
		return browser + " " + version + " · " + os
	}

	if version != "" {
		return browser + " " + version
	}

	if os != "" {
		return browser + " · " + os
	}

	return browser
}
