/*
 *  Copyright 2018 The CovenantSQL Authors.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package types

import (
	pi "github.com/SQLess/SQLess/blockproducer/interfaces"
	"github.com/SQLess/SQLess/crypto"
	"github.com/SQLess/SQLess/crypto/asymmetric"
	"github.com/SQLess/SQLess/crypto/verifier"
	"github.com/SQLess/SQLess/proto"
)

//go:generate hsp

// Range defines a height range (from, to].
type Range struct {
	From, To uint32
}

// MinerIncome defines the income of miner.
type MinerIncome struct {
	Miner  proto.AccountAddress
	Income uint64
}

// UserCost defines the cost of user.
type UserCost struct {
	User   proto.AccountAddress
	Cost   uint64
	Miners []*MinerIncome
}

// UpdateBillingHeader defines the UpdateBilling transaction header.
type UpdateBillingHeader struct {
	Receiver proto.AccountAddress
	Nonce    pi.AccountNonce
	Users    []*UserCost
	Range    Range
	Version  int32 `hsp:"v,version"`
}

// UpdateBilling defines the UpdateBilling transaction.
type UpdateBilling struct {
	UpdateBillingHeader
	pi.TransactionTypeMixin
	verifier.DefaultHashSignVerifierImpl
}

// NewUpdateBilling returns new instance.
func NewUpdateBilling(header *UpdateBillingHeader) *UpdateBilling {
	return &UpdateBilling{
		UpdateBillingHeader:  *header,
		TransactionTypeMixin: *pi.NewTransactionTypeMixin(pi.TransactionTypeUpdateBilling),
	}
}

// GetAccountAddress implements interfaces/Transaction.GetAccountAddress.
func (ub *UpdateBilling) GetAccountAddress() proto.AccountAddress {
	addr, _ := crypto.PubKeyHash(ub.Signee)
	return addr
}

// GetAccountNonce implements interfaces/Transaction.GetAccountNonce.
func (ub *UpdateBilling) GetAccountNonce() pi.AccountNonce {
	return ub.Nonce
}

// Sign implements interfaces/Transaction.Sign.
func (ub *UpdateBilling) Sign(signer *asymmetric.PrivateKey) (err error) {
	return ub.DefaultHashSignVerifierImpl.Sign(&ub.UpdateBillingHeader, signer)
}

// Verify implements interfaces/Transaction.Verify.
func (ub *UpdateBilling) Verify() (err error) {
	return ub.DefaultHashSignVerifierImpl.Verify(&ub.UpdateBillingHeader)
}

func init() {
	pi.RegisterTransaction(pi.TransactionTypeUpdateBilling, (*UpdateBilling)(nil))
}
