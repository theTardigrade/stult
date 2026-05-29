package main

import (
	"encoding/binary"
	"fmt"
)

const bundleFooterMagicString = "STULTBUNDLE1"

const bundleFooterZipSizeBytes = 8

func makeBundleFooter(zipSize uint64) []byte {
	footer := make([]byte, bundleFooterSize())

	binary.LittleEndian.PutUint64(footer[:bundleFooterZipSizeBytes], zipSize)
	copy(footer[bundleFooterZipSizeBytes:], []byte(bundleFooterMagicString))

	return footer
}

func parseBundleFooter(footer []byte) (uint64, bool, error) {
	if len(footer) != bundleFooterSize() {
		return 0, false, fmt.Errorf("Invalid bundle footer size")
	}

	magic := string(footer[bundleFooterZipSizeBytes:])
	if magic != bundleFooterMagicString {
		return 0, false, nil
	}

	zipSize := binary.LittleEndian.Uint64(footer[:bundleFooterZipSizeBytes])

	if zipSize == 0 {
		return 0, false, fmt.Errorf("Invalid embedded bundle: archive is empty")
	}

	return zipSize, true, nil
}

func bundleFooterSize() int {
	return bundleFooterZipSizeBytes + len(bundleFooterMagicString)
}

func embeddedBundleStart(bytes []byte) (int, bool, error) {
	footerSize := bundleFooterSize()

	if len(bytes) < footerSize {
		return 0, false, nil
	}

	footer := bytes[len(bytes)-footerSize:]

	zipSize, found, err := parseBundleFooter(footer)
	if err != nil {
		return 0, false, err
	}

	if !found {
		return 0, false, nil
	}

	if zipSize > uint64(len(bytes)-footerSize) {
		return 0, false, fmt.Errorf("Invalid existing embedded bundle: archive size is larger than executable")
	}

	return len(bytes) - footerSize - int(zipSize), true, nil
}
