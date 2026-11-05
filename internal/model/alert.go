package model

import "time"

type AlertConfig struct {
	TrendWindowDays  int     `yaml:"trend_window_days"`
	CooldownMinutes  int     `yaml:"cooldown_minutes"`
	SellBelowAvgPct  float64 `yaml:"sell_below_avg_pct"`
	Change48hPct     float64 `yaml:"change_48h_pct"`
	MaxPerCycle      int     `yaml:"max_per_cycle"`
}

type AlertState struct {
	LastFired map[string]time.Time
}

func NewAlertState() *AlertState {
	return &AlertState{LastFired: make(map[string]time.Time)}
}

func (s *AlertState) CanFire(key string, cooldown time.Duration) bool {
	t, ok := s.LastFired[key]
	return !ok || time.Since(t) >= cooldown
}

func (s *AlertState) Record(key string) {
	s.LastFired[key] = time.Now()
}
