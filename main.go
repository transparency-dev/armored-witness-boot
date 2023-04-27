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
	"fmt"
	"log"

	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"

	"github.com/usbarmory/armory-boot/exec"

	"github.com/transparency-dev/armored-witness-boot/config"
	"github.com/transparency-dev/armored-witness-boot/storage"
)

// Quorum defines the number of required authentication signatures for
// unikernel loading.
const Quorum = 2

var (
	Build    string
	Revision string

	PublicKeys string
)

func init() {
	log.SetFlags(0)

	if err := imx6ul.SetARMFreq(900); err != nil {
		panic(fmt.Sprintf("cannot change ARM frequency, %v\n", err))
	}

	imx6ul.DCP.Init()
}

func preLaunch() {
	usbarmory.LED("blue", false)
	usbarmory.LED("white", false)
}

func main() {
	card := usbarmory.MMC

	usbarmory.LED("blue", false)
	usbarmory.LED("white", false)

	if len(PublicKeys) == 0 {
		panic("missing public keys, aborting")
	}

	if err := card.Detect(); err != nil {
		panic(fmt.Sprintf("boot media error, %v\n", err))
	}

	usbarmory.LED("blue", true)

	log.Printf("armory-witness-boot: loading configuration at USDHC%d@%d\n", card.Index, config.Offset)

	conf, err := storage.Configuration(card, config.Offset, config.MaxLength)

	if err != nil {
		panic(fmt.Sprintf("configuration read error, %v\n", err))
	}

	kernel, err := card.Read(conf.Offset, conf.Size)

	if err != nil {
		panic(fmt.Sprintf("kernel read error, %v\n", err))
	}

	if err = conf.Verify(kernel, PublicKeys, Quorum); err != nil {
		panic(fmt.Sprintf("configuration verification error, %v\n", err))
	}

	usbarmory.LED("white", true)

	log.Printf("armory-witness-boot: loaded kernel off:%x size:%d", conf.Offset, conf.Size)

	image := &exec.ELFImage{
		Region: mem,
		ELF:    kernel,
	}

	if err = image.Load(); err != nil {
		panic(fmt.Sprintf("load error, %v\n", err))
	}

	log.Printf("armory-witness-boot: starting kernel@%.8x\n", image.Entry())

	if err = image.Boot(preLaunch); err != nil {
		panic(fmt.Sprintf("load error, %v\n", err))
	}
}
