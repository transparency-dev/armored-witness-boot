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
	"encoding/json"
	"fmt"
	"log"

	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/usdhc"
	"golang.org/x/mod/sumdb/note"

	"github.com/usbarmory/armory-boot/exec"

	"github.com/transparency-dev/armored-witness-boot/config"
	"github.com/transparency-dev/armored-witness-common/release/firmware"
)

const (
	// Quorum defines the number of required authentication signatures for
	// unikernel loading.
	Quorum = 2

	expectedBlockSize = 512
	osConfBlock       = 0x5000
)

var (
	Build    string
	Revision string

	OSLogOrigin         string
	OSLogVerifier       string
	OSManifestVerifiers string
)

// DMA region for target kernel boot
var mem *dma.Region

func init() {
	var err error

	log.SetFlags(0)

	if imx6ul.Native {
		imx6ul.SetARMFreq(imx6ul.Freq528)
	}

	dma.Init(dmaStart, dmaSize)

	if mem, err = dma.NewRegion(memoryStart, memorySize, false); err != nil {
		panic("could not allocate dma region")
	}

	mem.Reserve(memorySize, 0)
}

func preLaunch() {
	usbarmory.LED("blue", false)
	usbarmory.LED("white", false)
}

// read reads the firmware bundle for the trusted OS from internal storage, the
// OS proof bundle is *not* verified by this function.
func read(card *usdhc.USDHC) (fw *firmware.Bundle, err error) {
	blockSize := card.Info().BlockSize
	if blockSize != expectedBlockSize {
		return nil, fmt.Errorf("h/w invariant error - expected MMC blocksize %d, found %d", expectedBlockSize, blockSize)
	}

	buf, err := card.Read(osConfBlock*expectedBlockSize, config.MaxLength)
	if err != nil {
		return nil, err
	}

	conf := &config.Config{}
	if err = conf.Decode(buf); err != nil {
		return nil, err
	}

	fw = &firmware.Bundle{
		Checkpoint:     conf.Bundle.Checkpoint,
		Index:          conf.Bundle.LogIndex,
		InclusionProof: conf.Bundle.InclusionProof,
		Manifest:       conf.Bundle.Manifest,
	}

	fw.Firmware, err = card.Read(conf.Offset, conf.Size)
	if err != nil {
		return nil, fmt.Errorf("Failed to read firmware: %v", err)
	}

	return fw, nil
}

func main() {
	card := usbarmory.MMC

	usbarmory.LED("blue", false)
	usbarmory.LED("white", false)

	if len(OSManifestVerifiers) == 0 {
		panic("armored-witness-boot: missing public keys, aborting")
	}

	if err := card.Detect(); err != nil {
		panic(fmt.Sprintf("armored-witness-boot: boot media error, %v\n", err))
	}

	usbarmory.LED("blue", true)

	logVerifier, err := note.NewVerifier(OSLogVerifier)
	if err != nil {
		panic(fmt.Sprintf("armored-witness-boot: Invalid OSLogVerifier: %v", err))
	}
	log.Printf("armored-witness-boot: log verifier: %s", logVerifier.Name())

	manifestVerifiers, err := manifestVerifiers()
	if err != nil {
		panic(fmt.Sprintf("armored-witness-boot: Invalid OSManifestVerifiers: %v", err))
	}

	log.Printf("armored-witness-boot: loading configuration & kernel at USDHC%d@%d\n", card.Index, config.Offset)
	os, err := read(card)
	if err != nil {
		panic(fmt.Sprintf("armored-witness-boot: Failed to read OS firmware bundle: %v", err))
	}

	bv := &firmware.BundleVerifier{
		LogOrigin:         OSLogOrigin,
		LogVerifer:        logVerifier,
		ManifestVerifiers: manifestVerifiers,
	}
	if err := bv.Verify(*os); err != nil {
		panic(fmt.Sprintf("armored-witness-boot: kernel verification error, %v", err))
	}

	usbarmory.LED("white", true)

	log.Print("armored-witness-boot: verified kernel")

	image := &exec.ELFImage{
		Region: mem,
		ELF:    os.Firmware,
	}

	if err = image.Load(); err != nil {
		panic(fmt.Sprintf("load error, %v\n", err))
	}

	log.Printf("armored-witness-boot: starting kernel@%.8x\n", image.Entry())

	if err = image.Boot(preLaunch); err != nil {
		panic(fmt.Sprintf("armored-witness-boot: load error, %v\n", err))
	}
}

func manifestVerifiers() ([]note.Verifier, error) {
	var manifestKeys []string
	if err := json.Unmarshal([]byte(OSManifestVerifiers), &manifestKeys); err != nil {
		return nil, fmt.Errorf("invalid OSManifestVerifiers format: %v", err)
	}
	manifestVerifiers := []note.Verifier{}
	for _, v := range manifestKeys {
		mv, err := note.NewVerifier(v)
		if err != nil {
			return nil, fmt.Errorf("invalid OSManifestVerifier %q: %v", v, err)
		}
		manifestVerifiers = append(manifestVerifiers, mv)
		log.Printf("armored-witness-boot: kernel verifier: %v", v)
	}
	if l := len(manifestVerifiers); l != Quorum {
		return nil, fmt.Errorf("insufficient number of kernel manifest verifiers %d, need quorum of %d", l, Quorum)
	}

	return manifestVerifiers, nil
}
