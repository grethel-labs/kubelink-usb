package utils

import "strings"

// ParseUSBIdentifiers extracts vendor and product IDs from common udev property format.
func ParseUSBIdentifiers(id string) (vendorID, productID string) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}
