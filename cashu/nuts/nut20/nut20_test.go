package nut20

import (
	"encoding/hex"
	"testing"

	"github.com/OpenTollGate/gonuts-tollgate/cashu"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func TestSignMintQuote(t *testing.T) {
	privateKey, _ := secp256k1.GeneratePrivateKey()

	tests := []struct {
		quoteId    string
		outputs    cashu.BlindedMessages
		privateKey *secp256k1.PrivateKey
	}{
		{
			quoteId: "9d745270-1405-46de-b5c5-e2762b4f5e00",
			outputs: cashu.BlindedMessages{
				cashu.BlindedMessage{
					Amount: 1,
					Id:     "00456a94ab4e1c46",
					B_:     "0342e5bcc77f5b2a3c2afb40bb591a1e27da83cddc968abdc0ec4904201a201834",
				},
				cashu.BlindedMessage{
					Amount: 1,
					Id:     "00456a94ab4e1c46",
					B_:     "032fd3c4dc49a2844a89998d5e9d5b0f0b00dde9310063acb8a92e2fdafa4126d4",
				},
				cashu.BlindedMessage{
					Amount: 1,
					Id:     "00456a94ab4e1c46",
					B_:     "033b6fde50b6a0dfe61ad148fff167ad9cf8308ded5f6f6b2fe000a036c464c311",
				},
				cashu.BlindedMessage{
					Amount: 1,
					Id:     "00456a94ab4e1c46",
					B_:     "02be5a55f03e5c0aaea77595d574bce92c6d57a2a0fb2b5955c0b87e4520e06b53",
				},
				cashu.BlindedMessage{
					Amount: 1,
					Id:     "00456a94ab4e1c46",
					B_:     "02209fc2873f28521cbdde7f7b3bb1521002463f5979686fd156f23fe6a8aa2b79",
				},
			},
			privateKey: privateKey,
		},
	}

	for _, test := range tests {
		sig, err := SignMintQuote(test.privateKey, test.quoteId, test.outputs)
		if err != nil {
			t.Fatalf("got unexpected error signing mint quote: %v", err)
		}

		if !VerifyMintQuoteSignature(sig, test.quoteId, test.outputs, test.privateKey.PubKey()) {
			t.Fatal("generated invalid signature on mint quote")
		}
	}
}

func TestVerifyMintQuoteSignature(t *testing.T) {
	outputs := cashu.BlindedMessages{
		cashu.BlindedMessage{
			Amount: 1,
			Id:     "00456a94ab4e1c46",
			B_:     "0342e5bcc77f5b2a3c2afb40bb591a1e27da83cddc968abdc0ec4904201a201834",
		},
		cashu.BlindedMessage{
			Amount: 1,
			Id:     "00456a94ab4e1c46",
			B_:     "032fd3c4dc49a2844a89998d5e9d5b0f0b00dde9310063acb8a92e2fdafa4126d4",
		},
	}
	quoteId := "9d745270-1405-46de-b5c5-e2762b4f5e00"

	privKey, _ := secp256k1.GeneratePrivateKey()

	sig, err := SignMintQuote(privKey, quoteId, outputs)
	if err != nil {
		t.Fatalf("signing failed: %v", err)
	}

	if !VerifyMintQuoteSignature(sig, quoteId, outputs, privKey.PubKey()) {
		t.Fatal("valid signature rejected")
	}

	otherPrivKey, _ := secp256k1.GeneratePrivateKey()
	if VerifyMintQuoteSignature(sig, quoteId, outputs, otherPrivKey.PubKey()) {
		t.Fatal("invalid signature accepted (wrong key)")
	}

	sigHex := hex.EncodeToString(sig.Serialize())
	pubHex := hex.EncodeToString(privKey.PubKey().SerializeCompressed())
	if len(sigHex) != 128 {
		t.Fatalf("signature hex length = %d, want 128", len(sigHex))
	}
	if len(pubHex) != 66 {
		t.Fatalf("pubkey hex length = %d, want 66", len(pubHex))
	}
}

func TestBuildMessageToSign(t *testing.T) {
	msg := buildMessageToSign("testquote", cashu.BlindedMessages{{Amount: 1, B_: "aa"}, {Amount: 2, B_: "bb"}})
	want := []byte("testquoteaabb")
	if string(msg) != string(want) {
		t.Fatalf("message = %q, want %q (quote id || concatenated B_ hex)", msg, want)
	}
}

// TestMessageToSignMatchesCdkVector pins gonuts' NUT-20 message construction
// against cashubtc/cdk's own cross-implementation vector (crates/cashu/src/
// nuts/nut20.rs, test_msg_to_sign + test_valid_signature): the exact quote id,
// five outputs, and the valid (pubkey, signature) pair cdk verifies. If either
// implementation changes its framing, this test breaks — which is the point:
// the deployed-mint interop contract lives in cdk, not in our test file.
func TestMessageToSignMatchesCdkVector(t *testing.T) {
	const quote = "9d745270-1405-46de-b5c5-e2762b4f5e00"
	outputs := cashu.BlindedMessages{
		{Amount: 1, Id: "00456a94ab4e1c46", B_: "0342e5bcc77f5b2a3c2afb40bb591a1e27da83cddc968abdc0ec4904201a201834"},
		{Amount: 1, Id: "00456a94ab4e1c46", B_: "032fd3c4dc49a2844a89998d5e9d5b0f0b00dde9310063acb8a92e2fdafa4126d4"},
		{Amount: 1, Id: "00456a94ab4e1c46", B_: "033b6fde50b6a0dfe61ad148fff167ad9cf8308ded5f6f6b2fe000a036c464c311"},
		{Amount: 1, Id: "00456a94ab4e1c46", B_: "02be5a55f03e5c0aaea77595d574bce92c6d57a2a0fb2b5955c0b87e4520e06b53"},
		{Amount: 1, Id: "00456a94ab4e1c46", B_: "02209fc2873f28521cbdde7f7b3bb1521002463f5979686fd156f23fe6a8aa2b79"},
	}

	// The message must be exactly the UTF-8 quote id followed by each B_ hex
	// string in request order — cdk's expected byte array, as one string.
	want := quote +
		"0342e5bcc77f5b2a3c2afb40bb591a1e27da83cddc968abdc0ec4904201a201834" +
		"032fd3c4dc49a2844a89998d5e9d5b0f0b00dde9310063acb8a92e2fdafa4126d4" +
		"033b6fde50b6a0dfe61ad148fff167ad9cf8308ded5f6f6b2fe000a036c464c311" +
		"02be5a55f03e5c0aaea77595d574bce92c6d57a2a0fb2b5955c0b87e4520e06b53" +
		"02209fc2873f28521cbdde7f7b3bb1521002463f5979686fd156f23fe6a8aa2b79"
	if got := string(buildMessageToSign(quote, outputs)); got != want {
		t.Fatalf("message does not match cdk's vector:\n got %s\nwant %s", got, want)
	}

	// cdk's own valid signature for this exact message and pubkey verifies —
	// the interop pin, not just the construction pin.
	pubKeyBytes, err := hex.DecodeString("03d56ce4e446a85bbdaa547b4ec2b073d40ff802831352b8272b7dd7a4de5a7cac")
	if err != nil {
		t.Fatal(err)
	}
	pubKey, err := secp256k1.ParsePubKey(pubKeyBytes)
	if err != nil {
		t.Fatal(err)
	}
	sigBytes, err := hex.DecodeString("d4b386f21f7aa7172f0994ee6e4dd966539484247ea71c99b81b8e09b1bb2acbc0026a43c221fd773471dc30d6a32b04692e6837ddaccf0830a63128308e4ee0")
	if err != nil {
		t.Fatal(err)
	}
	sig, err := schnorr.ParseSignature(sigBytes)
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyMintQuoteSignature(sig, quote, outputs, pubKey) {
		t.Fatal("cdk's valid vector signature does not verify — the message construction has drifted from what deployed mints check")
	}
}
