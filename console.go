// https://github.com/usbarmory/armory-witness-boot
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build console
// +build console

package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

func init() {
	banner := fmt.Sprintf("armory-witness-boot • %s/%s (%s) • %s %s • %s",
		runtime.GOOS, runtime.GOARCH, runtime.Version(),
		Revision, Build,
		imx6ul.Model())

	log.SetFlags(0)
	log.Printf("%s", banner)
}
