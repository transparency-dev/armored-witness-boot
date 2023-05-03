// Copyright 2022 The Armored Witness Boot authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package storage provides utilities for reading armory-witness-boot
// configuration and kernel files.
package storage

import (
	"github.com/usbarmory/tamago/soc/nxp/usdhc"

	"github.com/transparency-dev/armored-witness-boot/config"
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
