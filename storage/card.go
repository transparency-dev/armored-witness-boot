// https://github.com/usbarmory/armory-witness-boot
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package storage provides utilities for reading armory-witness-boot
// configuration and kernel files.
package storage

import (
	"github.com/usbarmory/tamago/soc/nxp/usdhc"

	"github.com/usbarmory/armory-witness-boot/config"
)

// Configuration reads an armory-witness-boot configuration data gob from a
// fixed offset on an MMC/SD card.
func Configuration(card *usdhc.USDHC, offset int64, size int64) (c *config.Config, err error) {
	buf, err := card.Read(offset, size)

	if err != nil {
		return
	}

	c = &config.Config{}
	err = c.Decode(buf)

	return
}
