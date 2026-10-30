package domain

type PriceBreakdown struct {
	GrossAmount      int64 `json:"gross_amount"`
	TaxAmount        int64 `json:"tax_amount"`
	CommissionAmount int64 `json:"commission_amount"`
	DevPayoutAmount  int64 `json:"dev_payout_amount"`
}

type PricingCalculator struct {
	commissionBasisPoints int64 // 1500 = 15.00%
	taxBasisPoints        int64 // 1000 = 10.00%
}

func NewPricingCalculator() *PricingCalculator {
	return &PricingCalculator{
		commissionBasisPoints: 1500, // 15% fair store fee
		taxBasisPoints:        1000, // 10% VAT
	}
}

// CalculateBreakdown computes exact integer minor units ensuring zero rounding leakage
func (c *PricingCalculator) CalculateBreakdown(grossMinorUnits int64) PriceBreakdown {
	if grossMinorUnits <= 0 {
		return PriceBreakdown{}
	}

	// 1. Tax is 10% of gross (or gross * 10 / 110 for tax-inclusive)
	// Standard Korean/EU VAT inclusive: net_sales = gross * 10000 / (10000 + tax_bp)
	netSales := (grossMinorUnits * 10000) / (10000 + c.taxBasisPoints)
	tax := grossMinorUnits - netSales

	// 2. Commission: 15% of net sales
	commission := (netSales * c.commissionBasisPoints) / 10000

	// 3. Developer Payout: remainder
	devPayout := netSales - commission

	// Invariant Check & Adjustment: Gross == Commission + Tax + DevPayout
	total := commission + tax + devPayout
	diff := grossMinorUnits - total
	devPayout += diff // Allocate rounding remainder to developer

	return PriceBreakdown{
		GrossAmount:      grossMinorUnits,
		TaxAmount:        tax,
		CommissionAmount: commission,
		DevPayoutAmount:  devPayout,
	}
}
