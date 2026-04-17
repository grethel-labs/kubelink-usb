#!/usr/bin/env bash

set -euo pipefail

overall_min="${OVERALL_MIN_COVERAGE:-8}"
security_min="${SECURITY_MIN_COVERAGE:-80}"
usbip_min="${USBIP_MIN_COVERAGE:-50}"

tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT

overall_profile="${tmp_dir}/coverage-overall.out"
security_profile="${tmp_dir}/coverage-security.out"
usbip_profile="${tmp_dir}/coverage-usbip.out"

go test ./... -coverprofile="${overall_profile}" -covermode=atomic >/dev/null
go test ./internal/security -coverprofile="${security_profile}" -covermode=atomic >/dev/null
go test ./internal/usbip -coverprofile="${usbip_profile}" -covermode=atomic >/dev/null

coverage_value() {
	local file="$1"
	go tool cover -func="${file}" | awk '/^total:/ {gsub("%", "", $3); print $3}'
}

assert_minimum() {
	local label="$1"
	local actual="$2"
	local minimum="$3"
	if ! awk -v actual="${actual}" -v minimum="${minimum}" 'BEGIN {exit !(actual >= minimum)}'; then
		echo "${label} coverage ${actual}% is below required ${minimum}%"
		exit 1
	fi
}

overall_cov="$(coverage_value "${overall_profile}")"
security_cov="$(coverage_value "${security_profile}")"
usbip_cov="$(coverage_value "${usbip_profile}")"

echo "Coverage summary:"
echo "- overall: ${overall_cov}% (min ${overall_min}%)"
echo "- internal/security: ${security_cov}% (min ${security_min}%)"
echo "- internal/usbip: ${usbip_cov}% (min ${usbip_min}%)"

assert_minimum "overall" "${overall_cov}" "${overall_min}"
assert_minimum "internal/security" "${security_cov}" "${security_min}"
assert_minimum "internal/usbip" "${usbip_cov}" "${usbip_min}"
