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
	"bytes"
	"encoding/gob"
	"flag"
	"io/fs"
	"io/ioutil"
	"log"
	"os"

	"github.com/usbarmory/armory-witness-boot/config"
)

type Flags struct {
	kernel string
	sig1   string
	sig2   string
	output string
}

var flags *Flags

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	flags = &Flags{}

	flag.StringVar(&flags.kernel, "k", "", "kernel image")
	flag.StringVar(&flags.sig1, "1", "", "signature #1 file")
	flag.StringVar(&flags.sig2, "2", "", "signature #2 file")
	flag.StringVar(&flags.output, "o", "", "output image")
}

func main() {
	var err error

	flag.Parse()

	if len(flags.kernel) <= 0 || len(flags.sig1) <= 0 || len(flags.sig2) <= 0 || len(flags.output) <= 0 {
		flag.PrintDefaults()
		return
	}

	elf, err := ioutil.ReadFile(flags.kernel)

	if err != nil {
		log.Fatal(err)
	}

	sig1, err := ioutil.ReadFile(flags.sig1)

	if err != nil {
		log.Fatal(err)
	}

	sig2, err := ioutil.ReadFile(flags.sig2)

	if err != nil {
		log.Fatal(err)
	}

	conf := &config.Config{
		Offset:     config.Offset + config.MaxLength,
		Size:       int64(len(elf)),
		Signatures: [][]byte{sig1, sig2},
	}

	buf := new(bytes.Buffer)

	if err = gob.NewEncoder(buf).Encode(conf); err != nil {
		log.Fatal(err)
	}

	pad := config.MaxLength - int64(buf.Len())

	buf.Write(make([]byte, pad))
	buf.Write(elf)

	if err = os.WriteFile(flags.output, buf.Bytes(), fs.ModeExclusive|0600); err != nil {
		log.Fatal(err)
	}

	log.Printf("written config gob and kernel (off:%d, len:%d) to %s", conf.Offset, conf.Size, flags.output)
}
