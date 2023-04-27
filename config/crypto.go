// https://github.com/usbarmory/armory-witness-boot
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source cod is governed by the license
// that can be found in the LICENSE file.

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
