package rpc

type TxTraceResult struct {
	TxHash string                  `json:"txHash,omitempty"`
	Result *TraceTransactionResult `json:"result,omitempty"`
	Error  string                  `json:"error,omitempty"`
}

type TraceTransactionResult struct {
	Type    string `json:"type,omitempty"`
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
	Value   string `json:"value,omitempty"`
	Gas     string `json:"gas,omitempty"`
	GasUsed string `json:"gasUsed,omitempty"`
	Input   string `json:"input,omitempty"`
	Output  string `json:"output,omitempty"`
	Time    string `json:"time,omitempty"`

	Calls []*TraceTransactionCall `json:"calls,omitempty"`
}

type TraceTransactionCall struct {
	Type    string `json:"type,omitempty"`
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
	Value   string `json:"value,omitempty"`
	Gas     string `json:"gas,omitempty"`
	GasUsed string `json:"gasUsed,omitempty"`
	Input   string `json:"input,omitempty"`
	Output  string `json:"output,omitempty"`

	Calls []*TraceTransactionCall `json:"calls,omitempty"`
}
