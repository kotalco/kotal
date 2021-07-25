package v1alpha1

// Default defaults genesis block parameters
func (g *Genesis) Default() {
	if g.Coinbase == "" {
		g.Coinbase = DefaultCoinbase
	}

	if g.Difficulty == "" {
		g.Difficulty = DefaultDifficulty
	}

	if g.Forks == nil {
		g.Forks = &Forks{}
	}

	if g.MixHash == "" {
		g.MixHash = DefaultMixHash
	}

	if g.GasLimit == "" {
		g.GasLimit = DefaultGasLimit
	}

	if g.Nonce == "" {
		g.Nonce = DefaultNonce
	}

	if g.Timestamp == "" {
		g.Timestamp = DefaultTimestamp
	}

	if g.Clique != nil {
		if g.Clique.BlockPeriod == 0 {
			g.Clique.BlockPeriod = DefaultCliqueBlockPeriod
		}
		if g.Clique.EpochLength == 0 {
			g.Clique.EpochLength = DefaultCliqueEpochLength
		}
	}

	if g.IBFT2 != nil {
		if g.IBFT2.BlockPeriod == 0 {
			g.IBFT2.BlockPeriod = DefaultIBFT2BlockPeriod
		}
		if g.IBFT2.EpochLength == 0 {
			g.IBFT2.EpochLength = DefaultIBFT2EpochLength
		}
		if g.IBFT2.RequestTimeout == 0 {
			g.IBFT2.RequestTimeout = DefaultIBFT2RequestTimeout
		}
		if g.IBFT2.MessageQueueLimit == 0 {
			g.IBFT2.MessageQueueLimit = DefaultIBFT2MessageQueueLimit
		}
		if g.IBFT2.DuplicateMessageLimit == 0 {
			g.IBFT2.DuplicateMessageLimit = DefaultIBFT2DuplicateMessageLimit
		}
		if g.IBFT2.FutureMessagesLimit == 0 {
			g.IBFT2.FutureMessagesLimit = DefaultIBFT2FutureMessagesLimit
		}
		if g.IBFT2.FutureMessagesMaxDistance == 0 {
			g.IBFT2.FutureMessagesMaxDistance = DefaultIBFT2FutureMessagesMaxDistance
		}
	}
}
