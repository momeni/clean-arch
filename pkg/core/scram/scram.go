// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package scram exports the expected interfaces for Salted Challenge
// Response Authentication Mechanism (SCRAM). For the corresponding
// implementation, check the adapter layer.
//
// Interfaces should be defined based on the use cases requirements.
// For example, scram-sha-1 and scram-sha-256 as defined in RFC 7677
// and RFC 5802 can be used for generation of challenge and response
// messages using HMAC and SHA1 or SHA256 hash algorithms, so a client
// may authenticate to a server (and also verify the server identity).
// If we wanted to support that authentication mechanism, two interfaces
// would be required such as a ClientConversation and ServerConversation
// which could accept messages of the other party and create messages to
// be sent to that party. Also, a Builder interface might be required
// for instantiation of ClientConversation having a user/pass and
// instantiation of ServerConversation having a serialized hash string.
// Then, ClientConversation could be used to produce a serialized hash
// string which could be stored at server-side and used later for
// instantiation of ServerConversation instances.
// However, in our use cases, it is only required to generate a hash
// string with standard format (having a password, salt, and iteration
// count), so it can be passed to a PostgreSQL server.
// The server and client side SCRAM implementations are managed by the
// PostgreSQL server and its driver in the adapter layer, and so they
// are not needed in the use cases layer.
//
// Also, for example, if other hashing schemes (not only the SCRAM
// family) were desired, then a more abstract interface should be
// defined in the core layer and SCRAM related details would be kept
// in the adapter layer.
//
// See the Hasher interface for the expected SCRAM implementation
// features. This interface is used by the migrationuc package in order
// to change the database role passwords without sending the plaintext
// passwords in the relevant DDL queries (so their possible logging is
// not a threat). It is also used by the migration test cases.
package scram

// Hasher represents the expectations from a SCRAM hasher implementation
// which for a specific underlying hash function (e.g., SHA1 or SHA256)
// computes the storedKey and serverKey values whenever its Hash method
// is called with the relevant pass, salt, and iters arguments,
// representing password, random salt value, and hashing iterations
// count. Note that although username and authorization identifier are
// required in a SCRAM protocol, but they do no affect the storedKey and
// serverKey and so are not asked by the Hasher interface. A PBKDF2
// algorithm is computed in order to slow down a dictionary attack as
// detailed in RFC 5802.
type Hasher interface {
	// Hash computes a hash string following the standard scram hash
	// format, so it can be stored and used later for authentication.
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
	Hash(pass, salt string, iters int) (string, error)
}
