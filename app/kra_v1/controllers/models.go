package controllers


type KraTaxInfoWrapper struct {
	TotalBets        int     `json:"total_bets"`
	TotalStake       float64 `json:"total_stake"`
	ExciseDutyStake  float64 `json:"excise_duty_stake"`
	ExciseDutyPaid   float64 `json:"excise_duty_paid"`
	ExciseDutyUnpaid float64 `json:"excise_duty_unpaid"`
	TotalWinnings    float64 `json:"total_winnings"`    //Total winnings
	TotalWinningBets float64 `json:"totalwinning_bets"` //no of won bets
	WHTOnWinning     float64 `json:"WHTOn_winnings"`    //WHT on Winnings
	WHTPaid          float64 `json:"WHT_paid"`         //Paid successfully
	WHTUnpaid        float64 `json:"WHT_Unpaid"`       //Not yet paid 
	PassKey          string `json:"pass_key"`
}

type Version struct {
	Version string `json:"Version"`
}


