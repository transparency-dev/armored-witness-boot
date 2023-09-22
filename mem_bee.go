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

//go:build bee
// +build bee

package main

import (
	"log"
	_ "unsafe"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/soc/nxp/bee"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = 0x90000000

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = 0x08000000

// DMA region for bootloader operation
const (
	dmaStart = 0x98000000
	dmaSize  = 0x08000000
)

// DMA region for target kernel boot
const (
	// This memory layout enables AES CTR encryption for all external RAM
	// through BEE, when used the target kernel must be compiled for the
	// aliased memory region.
	memoryStart = bee.AliasRegion0
	memorySize  = 0x10000000
)

func init() {
	if !imx6ul.Native {
		log.Fatalf("could not activate BEE: unsupported under emulation")
	}

	if imx6ul.BEE == nil {
		log.Fatalf("could not activate BEE: unsupported hardware")
	}

	// Encrypt 1GB of external RAM, this is the maximum extent either
	// covered by the BEE or available on USB armory Mk II boards.
	region0 := uint32(imx6ul.MMDC_BASE)
	region1 := region0 + bee.AliasRegionSize

	imx6ul.BEE.Init()
	defer imx6ul.BEE.Lock()

	if err := imx6ul.BEE.Enable(region0, region1); err != nil {
		log.Fatalf("could not activate BEE: %v", err)
	}

	imx6ul.ARM.ConfigureMMU(
		bee.AliasRegion0,
		bee.AliasRegion1 + bee.AliasRegionSize,
		0,
		arm.TTE_CACHEABLE | arm.TTE_BUFFERABLE | arm.TTE_SECTION | arm.TTE_AP_001<<10,
	)
}
