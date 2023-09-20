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

package main

import (
	"log"
	_ "unsafe"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/soc/nxp/bee"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

// Override imx6ul.ramStart, usbarmory.ramSize and dma allocation, as the
// mapped kernel image needs to be within the first 128MiB of RAM.

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = 0x90000000

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = 0x08000000

// DMA region for target kernel boot
var mem *dma.Region

// DMA region for bootloader operation
const (
	dmaStart = 0x98000000
	dmaSize  = 0x08000000
)

// DMA region for target kernel boot
const (
	memoryStart = 0x80000000
	memorySize  = 0x10000000

	kernelOffset = 0x00800000
	paramsOffset = 0x07000000
	initrdOffset = 0x08000000
)

// BEE enables AES CTR encryption for all external RAM, when used the target
// kernel must be compiled for the aliased memory region.
const BEE = true

// Encrypt 1GB of external RAM, this is the maximum extent either
// covered by the BEE or available on USB armory Mk II boards.
func initRAMEncryption() {
	region0 := uint32(imx6ul.MMDC_BASE)
	region1 := region0 + bee.AliasRegionSize

	imx6ul.BEE.Init()
	defer imx6ul.BEE.Lock()

	if err := imx6ul.BEE.Enable(region0, region1); err != nil {
		log.Fatalf("could not activate BEE: %v", err)
	}

	flags := arm.TTE_CACHEABLE | arm.TTE_BUFFERABLE | arm.TTE_SECTION | arm.TTE_AP_001<<10
	imx6ul.ARM.ConfigureMMU(bee.AliasRegion0, bee.AliasRegion1 + bee.AliasRegionSize, 0, flags)
}

func init() {
	var start uint

	dma.Init(dmaStart, dmaSize)

	if BEE && imx6ul.Native && imx6ul.BEE != nil {
		initRAMEncryption()
		start = bee.AliasRegion0
	} else {
		start = memoryStart
	}

	mem, _ = dma.NewRegion(start, memorySize, false)
	mem.Reserve(memorySize, 0)
}
