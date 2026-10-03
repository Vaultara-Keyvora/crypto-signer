package crypto

type MnemonicBitSize int16

const (
	Nemonic128 MnemonicBitSize = 128
	Nemonic256 MnemonicBitSize = 256
)

const (
	PacketEthChain    = 1
	PacketBSCChain    = 2
	PacketTRONChain   = 3
	PacketSOLANAChain = 4
)
