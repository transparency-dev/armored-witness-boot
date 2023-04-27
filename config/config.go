// https://github.com/usbarmory/armory-witness-boot
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package config provides parsing for the armory-witness-boot configuration
// file format.
package config

import (
	"bytes"
	"encoding/gob"
)

const (
	Offset    = 10485760
	MaxLength = 40960
)

// Config represents the armory-witness-boot configuration.
type Config struct {
	// Offset is the MMC/SD card offset to an ELF unikernel image (e.g. TamaGo).
	Offset int64
	// Size is the unikernel length.
	Size int64
	// Signatures are the unikernel signify/minisign signatures.
	Signatures [][]byte

	pubKeys []string
}

// Encode serializes the configuration.
func (c *Config) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(c)

	return buf.Bytes(), err
}

// Decode deserializes the configuration.
func (c *Config) Decode(buf []byte) (err error) {
	// TODO: Go encoding/gob makes the following commitment:
	//
	// "Any future changes to the package will endeavor to maintain
	// compatibility with streams encoded using previous versions"
	//
	// Do we treat this as sufficient considering that we will throw away
	// the secure boot signing keys for this firmware?
	return gob.NewDecoder(bytes.NewBuffer(buf)).Decode(c)
}
