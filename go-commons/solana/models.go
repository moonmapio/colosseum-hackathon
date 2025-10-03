package solana

type MintCreateLog struct {
	Signature string
	Slot      uint64
}

type LogsNotification struct {
	Method string `json:"method"`
	Params struct {
		Result struct {
			Context struct {
				Slot uint64 `json:"slot"`
			} `json:"context"`
			Value struct {
				Signature string   `json:"signature"`
				Err       any      `json:"err"`
				Logs      []string `json:"logs"`
			} `json:"value"`
		} `json:"result"`
		Subscription int `json:"subscription"`
	} `json:"params"`
}

type ProgramNotification struct {
	Method string `json:"method"`
	Params struct {
		Result struct {
			Context struct {
				Slot uint64 `json:"slot"`
			} `json:"context"`
			Value struct {
				Pubkey  string `json:"pubkey"`
				Account struct {
					Data struct {
						Program string `json:"program"`
						Parsed  struct {
							Type string `json:"type"`
							Info struct {
								Mint        string `json:"mint"`
								Owner       string `json:"owner"`
								TokenAmount struct {
									Amount   string   `json:"amount"`
									Decimals int      `json:"decimals"`
									UIAmount *float64 `json:"uiAmount"`
								} `json:"tokenAmount"`
							} `json:"info"`
						} `json:"parsed"`
					} `json:"data"`
					Owner string `json:"owner"`
				} `json:"account"`
			} `json:"value"`
		} `json:"result"`
		Subscription int `json:"subscription"`
	} `json:"params"`
}
