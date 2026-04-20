package utils

import "strings"

// ParseUSBIdentifiers extracts vendor and product IDs from common udev property format.
// Input is expected as "vendorID:productID" (e.g. "1d6b:0002").
func ParseUSBIdentifiers(id string) (vendorID, productID string) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}
