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

//go:build console
// +build console

package main

import (
	"fmt"
	"log"
	"runtime"
	"time"
	_ "unsafe"

	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

func init() {
	switch model, _ := usbarmory.Model(); model {
	case usbarmory.BETA, usbarmory.GAMMA:
		if debugConsole, err := usbarmory.DetectDebugAccessory(250 * time.Millisecond); err == nil {
			<-debugConsole
		}
	}

	banner := fmt.Sprintf("armored-witness-boot • %s/%s (%s) • %s %s • %s",
		runtime.GOOS, runtime.GOARCH, runtime.Version(),
		Revision, Build,
		imx6ul.Model())

	log.SetFlags(0)
	log.Printf("%s", banner)
}

//go:linkname printk runtime.printk
func printk(c byte) {
	usbarmory.UART2.Tx(c)
}
