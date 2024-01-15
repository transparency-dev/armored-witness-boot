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

// Package config provides parsing for the armored-witness-boot configuration
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

// ProofBundle represents a firmware transparency proof bundle.
type ProofBundle struct {
	// Checkpoint is a note-formatted checkpoint from a log which contains Manifest, below.
	Checkpoint []byte
	// Manifest contains metadata about a firmware release.
	Manifest []byte
	// LogIndex is the position within the log where Manifest was included.
	LogIndex uint64
	// InclusionProof is a proof for Manifest@Index being committed to by Checkpoint.
	InclusionProof [][]byte
}

// Config represents a firmware component configuration.
type Config struct {
	// Offset is the MMC/SD card offset to an ELF unikernel image (e.g. TamaGo).
	Offset int64
	// Size is the unikernel length.
	Size int64
	// Signatures are the unikernel signify/minisign signatures.
	Signatures [][]byte
	// Bundle contains firmware transparency artefacts relating to the firmware this config
	// references.
	Bundle ProofBundle
	// NewIdentity is whether the device should ignore its previous data and boot
	// as a new witness identity.
	NewIdentity bool
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
