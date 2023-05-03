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

//go:build linkramsize

package config

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/usbarmory/armory-boot/config"
)

// Verify authenticates the armory-witness-boot configuration hash signatures.
func (c *Config) Verify(buf []byte, pubKeys string, quorum int) (err error) {
	if err = json.Unmarshal([]byte(pubKeys), &c.pubKeys); err != nil {
		return
	}

	if len(pubKeys) < quorum {
		return fmt.Errorf("invalid configuration, at least %d authentication keys are required", quorum)
	}

	if len(c.Signatures) < quorum {
		return fmt.Errorf("invalid configuration, at least %d authentication signatures are required", quorum)
	}

	n := 0

	for i, pubKey := range c.pubKeys {
		log.Printf("armory-witness-boot: authenticating kernel (%s)", pubKey)

		if err = config.Verify(buf, c.Signatures[i], pubKey); err != nil {
			log.Printf("armory-witness-boot: %s, %v", pubKey, err)
			continue
		}

		n += 1
	}

	if n < quorum {
		return fmt.Errorf("invalid configuration, at least %d valid signatures are required", quorum)
	}

	return
}
