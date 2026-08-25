package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Benefit struct {
	Code       string   `json:"code"`
	Name       string   `json:"name"`
	Tiers      []string `json:"tiers"`
	Quota      int      `json:"quota"`
	WindowDays int      `json:"window_days"`
	Enabled    bool     `json:"enabled"`
}

type BenefitCatalog struct {
	items map[string]Benefit
}

func NewBenefitCatalog() *BenefitCatalog {
	return &BenefitCatalog{items: map[string]Benefit{
		"BASIC-GIFT":      {Code: "BASIC-GIFT", Name: "基础礼遇", Tiers: []string{"standard", "silver", "gold", "platinum"}, Quota: 1, WindowDays: 30, Enabled: true},
		"GOLD-CARE":       {Code: "GOLD-CARE", Name: "黄金关怀", Tiers: []string{"gold", "platinum"}, Quota: 2, WindowDays: 30, Enabled: true},
		"PLATINUM-LOUNGE": {Code: "PLATINUM-LOUNGE", Name: "铂金休息室", Tiers: []string{"platinum"}, Quota: 1, WindowDays: 90, Enabled: true},
	}}
}

func (c *BenefitCatalog) Add(benefit Benefit) error {
	if c == nil {
		return fmt.Errorf("catalog is nil")
	}
	benefit.Code = strings.ToUpper(strings.TrimSpace(benefit.Code))
	benefit.Name = strings.TrimSpace(benefit.Name)
	if benefit.Code == "" || benefit.Name == "" || benefit.Quota <= 0 || benefit.WindowDays <= 0 {
		return fmt.Errorf("invalid benefit definition")
	}
	if len(benefit.Tiers) == 0 {
		return fmt.Errorf("at least one tier is required")
	}
	if c.items == nil {
		c.items = map[string]Benefit{}
	}
	c.items[benefit.Code] = benefit
	return nil
}

func (c *BenefitCatalog) Get(code string) (Benefit, bool) {
	if c == nil {
		return Benefit{}, false
	}
	item, ok := c.items[strings.ToUpper(strings.TrimSpace(code))]
	return item, ok
}

func (c *BenefitCatalog) List() []Benefit {
	if c == nil {
		return nil
	}
	items := make([]Benefit, 0, len(c.items))
	for _, item := range c.items {
		item.Tiers = append([]string(nil), item.Tiers...)
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Code < items[j].Code })
	return items
}

func (c *BenefitCatalog) Eligible(user User, code string) (Benefit, error) {
	item, ok := c.Get(code)
	if !ok {
		return Benefit{}, fmt.Errorf("benefit %s not found", code)
	}
	if !item.Enabled || !user.Active {
		return Benefit{}, fmt.Errorf("benefit is not available")
	}
	for _, tier := range item.Tiers {
		if tier == user.Tier {
			return item, nil
		}
	}
	return Benefit{}, fmt.Errorf("tier %s is not eligible", user.Tier)
}

func (c *BenefitCatalog) WithinWindow(record Record, now time.Time) bool {
	item, ok := c.Get(record.BenefitCode)
	if !ok || now.Before(record.ReceivedAt) {
		return false
	}
	return now.Sub(record.ReceivedAt) <= time.Duration(item.WindowDays)*24*time.Hour
}

func (c *BenefitCatalog) QuotaFor(code string) int {
	item, ok := c.Get(code)
	if !ok {
		return 0
	}
	return item.Quota
}

func (c *BenefitCatalog) Disable(code string) bool {
	item, ok := c.Get(code)
	if !ok {
		return false
	}
	item.Enabled = false
	c.items[item.Code] = item
	return true
}

func (c *BenefitCatalog) Enable(code string) bool {
	item, ok := c.Get(code)
	if !ok {
		return false
	}
	item.Enabled = true
	c.items[item.Code] = item
	return true
}

func (c *BenefitCatalog) Codes() []string {
	items := c.List()
	codes := make([]string, 0, len(items))
	for _, item := range items {
		codes = append(codes, item.Code)
	}
	return codes
}

func (c *BenefitCatalog) Replace(items []Benefit) error {
	if c == nil {
		return fmt.Errorf("catalog is nil")
	}
	next := map[string]Benefit{}
	for _, item := range items {
		if err := (&BenefitCatalog{items: next}).Add(item); err != nil {
			return err
		}
	}
	c.items = next
	return nil
}

func (c *BenefitCatalog) Snapshot() map[string]Benefit {
	result := map[string]Benefit{}
	for _, item := range c.List() {
		result[item.Code] = item
	}
	return result
}
