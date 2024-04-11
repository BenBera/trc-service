package models

type Settings struct {

	//WithdrawStatus withdraw status, 1 (active) 0 for suspended
	WithdrawStatus int64 `json:"withdraw_status" validate:"required" enums:"1,0"`

	//WithdrawalMinimumAmount minimum withdrawal amount
	WithdrawalMinimumAmount int64 `json:"withdrawal_minimum_amount" validate:"required"`

	//WithdrawalMaximumAmount max withdrawal amount
	WithdrawalMaximumAmount int64 `json:"withdrawal_maximum_amount" validate:"required"`

	//WithdrawalDailyLimit daily withdrawal limit
	WithdrawalDailyLimit int64 `json:"withdrawal_daily_limit" validate:"required"`

	//WithdrawalBeforeMinimumSpend  1 (active) 0 for suspended
	WithdrawalBeforeMinimumSpend int64 `json:"withdrawal_before_minimum_spend" validate:"required" enums:"1,0"`

	//CannotWithdrawMoreThanCumulativeDeposit  1 (active) 0 for suspended
	CannotWithdrawMoreThanCumulativeDeposit int64 `json:"cannot_withdraw_more_than_cumulative_deposit" validate:"required" enums:"1,0"`

	//CannotWithdrawMoreThanCumulativeStake 1 (active) 0 for suspended
	CannotWithdrawMoreThanCumulativeStake int64 `json:"cannot_withdraw_more_than_cumulative_stake" validate:"required" enums:"1,0"`

	//CannotWithdrawMoreThanCumulativeWinning  1 (active) 0 for suspended
	CannotWithdrawMoreThanCumulativeWinning int64 `json:"cannot_withdraw_more_than_cumulative_winning" validate:"required" enums:"1,0"`
}

type UpdateSetting struct {

	//FieldName name of the field to update
	FieldName string `json:"field_name"  enums:"withdraw_status,withdrawal_minimum_amount,withdrawal_maximum_amount,withdrawal_daily_limit,withdrawal_before_minimum_spend,cannot_withdraw_more_than_cumulative_deposit,cannot_withdraw_more_than_cumulative_stake, cannot_withdraw_more_than_cumulative_winning"`

	//FieldValue field value to update
	FieldValue string `json:"field_value"`
}
