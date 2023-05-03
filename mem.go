// Copyright 2022 The ArmoredWitness Authors. All Rights Reserved.
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
	_ "unsafe"

	"github.com/usbarmory/tamago/dma"
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

func init() {
	dma.Init(dmaStart, dmaSize)

	mem, _ = dma.NewRegion(memoryStart, memorySize, false)
	mem.Reserve(memorySize, 0)
}
