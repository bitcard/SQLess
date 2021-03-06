/*
 * Copyright 2019 The CovenantSQL Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package route

import (
	"fmt"
	"net"
	"sync"

	"github.com/pkg/errors"

	"github.com/SQLess/SQLess/crypto"
	"github.com/SQLess/SQLess/crypto/asymmetric"
	"github.com/SQLess/SQLess/pow/cpuminer"
	"github.com/SQLess/SQLess/proto"
)

const (
	// ID is node id
	ID = "id."
	// PUBKEY is public key
	PUBKEY = "pub."
	// NONCE is nonce
	NONCE = "n."
	// ADDR is address
	ADDR = "addr."
)

// IPv6SeedClient is IPv6 DNS seed client
type IPv6SeedClient struct{}

// GetBPFromDNSSeed gets BP info from the IPv6 domain
func (isc *IPv6SeedClient) GetBPFromDNSSeed(BPDomain string) (BPNodes IDNodeMap, err error) {
	// Public key
	var pubKeyBuf []byte
	var pubBuf, nonceBuf, addrBuf, nodeIDBuf []byte
	var pubErr, nonceErr, addrErr, nodeIDErr error
	wg := new(sync.WaitGroup)
	wg.Add(4)

	f := func(host string) ([]net.IP, error) {
		return net.LookupIP(host)
	}
	// Public key
	go func() {
		defer wg.Done()
		pubBuf, pubErr = FromDomain(PUBKEY+BPDomain, f)
	}()
	// Nonce
	go func() {
		defer wg.Done()
		nonceBuf, nonceErr = FromDomain(NONCE+BPDomain, f)
	}()
	// Addr
	go func() {
		defer wg.Done()
		addrBuf, addrErr = FromDomain(ADDR+BPDomain, f)
	}()
	// NodeID
	go func() {
		defer wg.Done()
		nodeIDBuf, nodeIDErr = FromDomain(ID+BPDomain, f)
	}()

	wg.Wait()

	switch {
	case pubErr != nil:
		err = pubErr
		return
	case nonceErr != nil:
		err = nonceErr
		return
	case addrErr != nil:
		err = addrErr
		return
	case nodeIDErr != nil:
		err = nodeIDErr
		return
	}

	// For bug that trim the public header before or equal cql 0.7.0
	if len(pubBuf) == asymmetric.PublicKeyBytesLen-1 {
		pubKeyBuf = make([]byte, asymmetric.PublicKeyBytesLen)
		pubKeyBuf[0] = asymmetric.PublicKeyFormatHeader
		copy(pubKeyBuf[1:], pubBuf)
	} else if len(pubBuf) == 48 {
		pubKeyBuf, err = crypto.RemovePKCSPadding(pubBuf)
		if err != nil {
			return
		}
	} else {
		return nil, errors.Errorf("error public key bytes len: %d", len(pubBuf))
	}
	var pubKey asymmetric.PublicKey
	err = pubKey.UnmarshalBinary(pubKeyBuf)
	if err != nil {
		return
	}

	nonce, err := cpuminer.Uint256FromBytes(nonceBuf)
	if err != nil {
		return
	}

	addrBytes, err := crypto.RemovePKCSPadding(addrBuf)
	if err != nil {
		return
	}

	var nodeID proto.RawNodeID
	err = nodeID.SetBytes(nodeIDBuf)
	if err != nil {
		return
	}

	BPNodes = make(IDNodeMap)
	BPNodes[nodeID] = proto.Node{
		ID:        nodeID.ToNodeID(),
		Addr:      string(addrBytes),
		PublicKey: &pubKey,
		Nonce:     *nonce,
	}

	return
}

// GenBPIPv6 generates the IPv6 addrs contain BP info
func (isc *IPv6SeedClient) GenBPIPv6(node *proto.Node, domain string) (out string, err error) {
	// NodeID
	nodeIDIps, err := ToIPv6(node.ID.ToRawNodeID().AsBytes())
	if err != nil {
		return "", err
	}
	for i, ip := range nodeIDIps {
		out += fmt.Sprintf("%02d.%s%s	1	IN	AAAA	%s\n", i, ID, domain, ip)
	}

	pubKeyIps, err := ToIPv6(crypto.AddPKCSPadding(node.PublicKey.Serialize()))
	if err != nil {
		return "", err
	}
	for i, ip := range pubKeyIps {
		out += fmt.Sprintf("%02d.%s%s	1	IN	AAAA	%s\n", i, PUBKEY, domain, ip)
	}

	// Nonce
	nonceIps, err := ToIPv6(node.Nonce.Bytes())
	if err != nil {
		return "", err
	}
	for i, ip := range nonceIps {
		out += fmt.Sprintf("%02d.%s%s	1	IN	AAAA	%s\n", i, NONCE, domain, ip)
	}

	// Addr
	addrIps, err := ToIPv6(crypto.AddPKCSPadding([]byte(node.Addr)))
	if err != nil {
		return "", err
	}
	for i, ip := range addrIps {
		out += fmt.Sprintf("%02d.%s%s	1	IN	AAAA	%s\n", i, ADDR, domain, ip)
	}

	return
}
