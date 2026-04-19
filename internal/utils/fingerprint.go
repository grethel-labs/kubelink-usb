package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// dnsLabelUnsafe matches characters not allowed in DNS labels.
var dnsLabelUnsafe = regexp.MustCompile(`[^a-z0-9-]`)

// DeviceFingerprint returns a deterministic, DNS-label-safe name for a USB device.
//
// Intent: Generate reproducible CR names so reconnects update existing CRs instead of creating duplicates.
// Inputs: Node name, vendor ID, product ID, serial number (optional), and bus ID (fallback).
// Outputs: Lowercase, DNS-label-safe string of at most 63 characters.
// Errors: None.
func DeviceFingerprint(nodeName, vendorID, productID, serialNumber, busID string) string {
	var suffix string
	if serialNumber != "" {
		suffix = serialNumber
	} else {
		suffix = busID
	}

	raw := fmt.Sprintf("%s-%s-%s-%s", nodeName, vendorID, productID, suffix)
	return sanitizeDNSLabel(raw)
}

// sanitizeDNSLabel lowercases, replaces unsafe characters, collapses dashes,
// trims leading/trailing dashes, and truncates to 63 characters.
func sanitizeDNSLabel(s string) string {
	s = strings.ToLower(s)
	s = dnsLabelUnsafe.ReplaceAllString(s, "-")

	// Collapse multiple dashes.
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}

	s = strings.Trim(s, "-")

	if len(s) > 63 {
		s = s[:63]
		s = strings.TrimRight(s, "-")
	}

	if s == "" {
		return "unknown-device"
	}
	return s
}
