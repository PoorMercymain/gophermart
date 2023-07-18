package domain

type Goods struct {
	Match      string  `json:"match" valid:"-"`
	Reward     float64 `json:"reward" valid:"-"`
	RewardType string  `json:"reward_type" valid:"in(pt|%)"`
}
