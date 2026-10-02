package main

import (
	"fmt"
	"time"

	"github.com/OpenTollGate/gonuts-tollgate/cashu/nuts/nut04"
	"github.com/OpenTollGate/gonuts-tollgate/wallet"
)

func main() {
	const mint = "https://testnut.cashu.space/"
	w, err := wallet.LoadWallet(wallet.Config{WalletPath: "/tmp/opencode/nutred-wallet", CurrentMintURL: mint})
	if err != nil {
		panic(err)
	}
	if _, err := w.AddMint(mint); err != nil {
		panic(err)
	}
	q, err := w.RequestMint(4, mint)
	if err != nil {
		panic(err)
	}
	fmt.Println("quote:", q.Quote, "state:", q.State)
	paid := false
	for i := 0; i < 30; i++ {
		s, err := w.MintQuoteState(q.Quote)
		if err != nil {
			fmt.Println("state err:", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if s.State == nut04.Paid {
			paid = true
			fmt.Println("PAID after", i*2, "s")
			break
		}
		fmt.Println("state:", s.State)
		time.Sleep(2 * time.Second)
	}
	if !paid {
		fmt.Println("NEVER PAID — abort")
		return
	}
	amt, err := w.MintTokens(q.Quote)
	fmt.Printf("MintTokens: amount=%d err=%v\n", amt, err)
}
