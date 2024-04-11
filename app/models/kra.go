package models

import "time"

type KRAHeader struct {
	OperatorPin     string `json:"operatorPin"`
	TransactionDate string `json:"transactionDate"`
	NoOfStakes      string `json:"noOfStakes"`
}

type KRAStakeInfo struct {
	BetID               string `json:"betId"`
	CustomerID          string `json:"customerId"`
	MobileNo            string `json:"mobileNo"`
	PunterAmt           string `json:"punterAmt"`
	StakeAmt            string `json:"stakeAmt"`
	Desc                string `json:"desc"`
	Odds                string `json:"odds"`
	StakeType           string `json:"stakeType"`
	DateOfStake         string `json:"dateOfStake"`
	ExciseAmt           string `json:"exciseAmt"`
	ExpectedOutcomeTime string `json:"expectedOutcomeTime"`
	WalletBalanceStake  string `json:"walletBalanceStake"`
}

type StakeDetail struct {
	StakeInfo KRAStakeInfo `json:"stakeInfo"`
}

type KRARequest struct {
	Hash    string        `json:"hash"`
	Header  KRAHeader     `json:"header"`
	Details []StakeDetail `json:"details"`
}

type StakeData struct {
	Request KRARequest `json:"Request"`
}

type KRAStakeResponse struct {
	Response struct {
		Result struct {
			ResponseCode string `json:"ResponseCode"`
			Message      string `json:"Message"`
			Status       string `json:"Status"`
		} `json:"RESULT"`
	} `json:"RESPONSE"`
}

type OutcomeHeader struct {
	OperatorPin     string `json:"operatorPin"`
	TransactionDate string `json:"transactionDate"`
	NoOfOutcomes    string `json:"noOfOutcomes"`
}

type OutcomeInfo struct {
	BetID          string `json:"betId"`
	Outcome        string `json:"outcome"`
	Status         int    `json:"status"`
	Outcomedate    int    `json:"outcomedate"`
	Payout         string `json:"payout"`
	Winnings       string `json:"winnings"`
	WithholdingTax string `json:"withholdingTax"`
}

type OutcomeDetail struct {
	OutcomeInfo OutcomeInfo `json:"outcomeInfo"`
}

type OutcomeRequest struct {
	Hash    string          `json:"hash"`
	Header  OutcomeHeader   `json:"header"`
	Details []OutcomeDetail `json:"details"`
}

type OutcomeData struct {
	Request OutcomeRequest `json:"Request"`
}
type GeneratePRN struct {
	Amount    float64   `json:"amount"`
	TaxType   string    `json:"tax_type"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}
type PayTax struct {
	Amount        float64 `json:"amount"`
	Prn           string  `json:"prn"`
	TransactionID int64   `json:"transaction_id"`
}
