package nut20

import (
	"crypto/sha256"

	"github.com/OpenTollGate/gonuts-tollgate/cashu"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// NUT #20: This NUT defines signature-based authentication for mint quote redemption.

// buildMessageToSign constructs the NUT-20 spec message:
// quote_id || hex(B_) of each output, concatenated in request order.
//
// This is the format cdk mints verify (cashubtc/cdk nut20.rs
// legacy_mint_quote_msg_to_sign), accepted both by cdk versions that predate
// the domain-separated variant and by newer ones via their legacy fallback.
// The domain-separated "Cashu_MintQuoteSig_v1" framing is NOT understood by
// cdk mints currently deployed (verified against testnut.cashu.space:
// spec format mints, domain-separated is rejected with
// "Signature missing or invalid").
func buildMessageToSign(quoteId string, blindedMessages cashu.BlindedMessages) []byte {
	var msg []byte
	msg = append(msg, quoteId...)
	for _, bm := range blindedMessages {
		msg = append(msg, bm.B_...)
	}
	return msg
}

func SignMintQuote(
	privateKey *secp256k1.PrivateKey,
	quoteId string,
	blindedMessages cashu.BlindedMessages,
) (*schnorr.Signature, error) {
	msg := buildMessageToSign(quoteId, blindedMessages)
	hash := sha256.Sum256(msg)
	sig, err := schnorr.Sign(privateKey, hash[:])
	if err != nil {
		return nil, err
	}

	return sig, nil
}

// NUT #20: `pubkey` is the compressed secp256k1 public key (33 bytes, hex-encoded) that will be required for signature verification during the minting operation.
func VerifyMintQuoteSignature(
	signature *schnorr.Signature,
	quoteId string,
	blindedMessages cashu.BlindedMessages,
	publicKey *secp256k1.PublicKey,
) bool {
	msg := buildMessageToSign(quoteId, blindedMessages)
	hash := sha256.Sum256(msg)
	return signature.Verify(hash[:], publicKey)
}
