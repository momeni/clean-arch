// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package scram presents an implementation of SCRAM-SHA-256 and
// SCRAM-SHA-1 mechanisms. See the SHA256 and SHA1 functions for their
// instantiation logic. When a mechanism for a specific underlying hash
// function is instantiated, it can be used for generation of hash
// strings in the SCRAM standard format.
// This format is also known as the scram encrypted password format,
// however, it may not be reversed (so no encryption/decryption is
// taking place).
package scram

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/xdg-go/scram"
)

// Mechanism provides a Salted Challenge Response Authentication
// Mechanism (SCRAM) having a fixed underlying hash algorithm.
//
// It implements the github.com/momeni/clean-arch/pkg/core/scram.Hasher
// interface, so it may be used in the use cases layer without any
// dependency on the actual implementation. This package relies on
// the github.com/xdg-go/scram module for the SCRAM implementation.
type Mechanism struct {
	hashGenerator scram.HashGeneratorFcn
	outLen        int // bytes
	name          string
}

// SHA1 returns a new Mechanism instance using the SHA1 as its
// underlying hash algorithm.
func SHA1() *Mechanism {
	return &Mechanism{
		hashGenerator: scram.SHA1,
		outLen:        160 / 8,
		name:          "SCRAM-SHA-1",
	}
}

// SHA256 returns a new Mechanism instance using the SHA256 as its
// underlying hash algorithm.
func SHA256() *Mechanism {
	return &Mechanism{
		hashGenerator: scram.SHA256,
		outLen:        256 / 8,
		name:          "SCRAM-SHA-256",
	}
}

// Hash computes a hash string following the standard scram hash format,
// so it can be stored and used later for authentication.
//
// The pass argument must be non-empty. The user and authzID params
// are not asked because they are not used in the hash output. The
// given password will be normalized accoriding to the SASLprep
// profile (defined by RFC 4013) of the stringprep algorithm (which
// is defined by RFC 3454) and any failure in that normalization
// returns an error.
//
// The salt must contain a base64 encoding of the desired salt
// bytes, otherwise, if an empty value is passed, a random salt will
// be generated and used instead.
// The iters must be at least equal to 4096. However, the RFC 7677
// recommends to use 15000 or more.
//
// In absence of errors, a hashed string will be returned which
// conforms to the following format.
//
//	SCRAM-{SHA-X}${iters}:{b64-salt}${b64-storedKey}:{b64-serverKey}
//
// This string (consisting only of ASCII printable letters) can
// be safely passed to an ALTER or CREATE ROLE query in order to
// update or create a database role with the desired password as
// accepted by the PostgreSQL DBMS without risking to send a
// plaintext password.
func (m *Mechanism) Hash(pass, salt string, iters int) (string, error) {
	switch {
	case pass == "":
		return "", errors.New("password must be non-empty")
	case iters < 4096:
		return "", fmt.Errorf("iters (%d) is less than 4096", iters)
	}
	if salt == "" {
		saltBytes := make([]byte, m.outLen)
		if _, err := rand.Read(saltBytes); err != nil {
			return "", fmt.Errorf("creating random salt: %w", err)
		}
		s := make([]byte, base64.StdEncoding.EncodedLen(m.outLen))
		base64.StdEncoding.Encode(s, saltBytes)
		salt = string(s)
	}
	sc, err := m.storedCredentials(pass, salt, iters)
	if err != nil {
		return "", fmt.Errorf("obtaining stored credentials: %w", err)
	}
	h := fmt.Sprintf(
		"%s$%d:%s$%s:%s",
		m.name,
		iters, salt,
		base64.StdEncoding.EncodeToString(sc.StoredKey),
		base64.StdEncoding.EncodeToString(sc.ServerKey),
	)
	return h, nil
}

func (m *Mechanism) storedCredentials(
	pass, salt string, iters int,
) (*scram.StoredCredentials, error) {
	c, err := m.hashGenerator.NewClient("username", pass, "authzID")
	if err != nil {
		return nil, fmt.Errorf("creating SCRAM client: %w", err)
	}
	saltBytes, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return nil, fmt.Errorf("decoding base64 salt: %w", err)
	}
	// Indeed, these options are not required because we do not call
	// the NewConversation method. However, we have it here for sake
	// of completeness (similar to explanation of the ClientConversation
	// and ServerConversation interfaces in pkg/core/scram/scram.go).
	c = c.WithMinIterations(iters).WithNonceGenerator(func() string {
		return salt
	})
	sc := c.GetStoredCredentials(scram.KeyFactors{
		Salt:  string(saltBytes),
		Iters: iters,
	})
	return &sc, nil
}
