package main

import (
	"fmt"
	"lbe_crypto_signer/pkg/common/cmd"
)

func main() {
	if err := cmd.NewStartCmd().Execute(); err != nil {
		fmt.Println("cmd.NewStartCmd start failed. error = ", err)
	}
}
