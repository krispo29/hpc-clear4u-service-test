package inbound

import "github.com/shopspring/decimal"

func calcCustomsFee(totalHawb int) *CustomFeeModel {
	const maxHawbPerDeclaration = 40
	const feePerDeclaration = 200

	result := &CustomFeeModel{}

	// number of declarations
	result.TotalDeclaration = (totalHawb + maxHawbPerDeclaration - 1) / maxHawbPerDeclaration

	// total customs fee
	result.TotalFee = decimal.NewFromInt(int64(feePerDeclaration)).Mul(decimal.NewFromInt(int64(result.TotalDeclaration)))

	// exact fee per HAWB
	exactPerHawb := result.TotalFee.Div(decimal.NewFromInt(int64(totalHawb)))

	// floor (2 decimals) for everyone first
	result.FloorPerHawb = exactPerHawb.RoundFloor(2)

	result.PerHawbFees = make([]decimal.Decimal, totalHawb)
	for i := range result.PerHawbFees {
		result.PerHawbFees[i] = result.FloorPerHawb
	}

	// how much leftover to distribute
	sum := result.FloorPerHawb.Mul(decimal.NewFromInt(int64(totalHawb)))
	remainder := result.TotalFee.Sub(sum)

	// distribute remainder in +0.01 steps
	i := 0
	oneCent := decimal.NewFromFloat(0.01)
	for remainder.GreaterThan(decimal.Zero) {
		result.PerHawbFees[i] = result.PerHawbFees[i].Add(oneCent)
		remainder = remainder.Sub(oneCent)
		i++
	}

	return result
}

func calcOTCustomsFee(totalHawb int) *OTCustomFeeModel {
	const maxHawbPerDeclaration = 40
	const feePerDeclaration = 200

	result := &OTCustomFeeModel{}

	// number of declarations
	result.TotalDeclaration = (totalHawb + maxHawbPerDeclaration - 1) / maxHawbPerDeclaration

	// total customs fee
	result.TotalFee = decimal.NewFromInt(int64(feePerDeclaration)).Mul(decimal.NewFromInt(int64(result.TotalDeclaration)))

	// exact fee per HAWB
	exactPerHawb := result.TotalFee.Div(decimal.NewFromInt(int64(totalHawb)))

	// floor (2 decimals) for everyone first
	result.FloorPerHawb = exactPerHawb.RoundFloor(2)

	result.PerHawbFees = make([]decimal.Decimal, totalHawb)
	for i := range result.PerHawbFees {
		result.PerHawbFees[i] = result.FloorPerHawb
	}

	// how much leftover to distribute
	sum := result.FloorPerHawb.Mul(decimal.NewFromInt(int64(totalHawb)))
	remainder := result.TotalFee.Sub(sum)

	// distribute remainder in +0.01 steps
	i := 0
	oneCent := decimal.NewFromFloat(0.01)
	for remainder.GreaterThan(decimal.Zero) {
		result.PerHawbFees[i] = result.PerHawbFees[i].Add(oneCent)
		remainder = remainder.Sub(oneCent)
		i++
	}

	return result
}

func calcBankFee(totalHawb int) *BankFeeFeeModel {
	const maxHawbPerDeclaration = 40
	const feePerDeclaration = 70

	result := &BankFeeFeeModel{}

	// number of declarations
	result.TotalDeclaration = (totalHawb + maxHawbPerDeclaration - 1) / maxHawbPerDeclaration

	// total customs fee
	result.TotalFee = decimal.NewFromInt(int64(feePerDeclaration)).Mul(decimal.NewFromInt(int64(result.TotalDeclaration)))

	// exact fee per HAWB
	exactPerHawb := result.TotalFee.Div(decimal.NewFromInt(int64(totalHawb)))

	// floor (2 decimals) for everyone first
	result.FloorPerHawb = exactPerHawb.RoundFloor(2)

	result.PerHawbFees = make([]decimal.Decimal, totalHawb)
	for i := range result.PerHawbFees {
		result.PerHawbFees[i] = result.FloorPerHawb
	}

	// how much leftover to distribute
	sum := result.FloorPerHawb.Mul(decimal.NewFromInt(int64(totalHawb)))
	remainder := result.TotalFee.Sub(sum)

	// distribute remainder in +0.01 steps
	i := 0
	oneCent := decimal.NewFromFloat(0.01)
	for remainder.GreaterThan(decimal.Zero) {
		result.PerHawbFees[i] = result.PerHawbFees[i].Add(oneCent)
		remainder = remainder.Sub(oneCent)
		i++
	}

	return result
}

func calcCargoPermitFee(totalHawb int) *CargoPermitFeeModel {
	const maxHawbPerDeclaration = 40
	const feePerDeclaration = 150

	result := &CargoPermitFeeModel{}

	// number of declarations
	result.TotalDeclaration = (totalHawb + maxHawbPerDeclaration - 1) / maxHawbPerDeclaration

	// total customs fee
	result.TotalFee = decimal.NewFromInt(int64(feePerDeclaration)).Mul(decimal.NewFromInt(int64(result.TotalDeclaration)))

	// exact fee per HAWB
	exactPerHawb := result.TotalFee.Div(decimal.NewFromInt(int64(totalHawb)))

	// floor (2 decimals) for everyone first
	result.FloorPerHawb = exactPerHawb.RoundFloor(2)

	result.PerHawbFees = make([]decimal.Decimal, totalHawb)
	for i := range result.PerHawbFees {
		result.PerHawbFees[i] = result.FloorPerHawb
	}

	// how much leftover to distribute
	sum := result.FloorPerHawb.Mul(decimal.NewFromInt(int64(totalHawb)))
	remainder := result.TotalFee.Sub(sum)

	// distribute remainder in +0.01 steps
	i := 0
	oneCent := decimal.NewFromFloat(0.01)
	for remainder.GreaterThan(decimal.Zero) {
		result.PerHawbFees[i] = result.PerHawbFees[i].Add(oneCent)
		remainder = remainder.Sub(oneCent)
		i++
	}

	return result
}

func calcExpressDeliveryFee(totalHawb int) *ExpressDeliveryFeeModel {
	feePerMasterAirwayBill := decimal.NewFromInt(380)

	result := &ExpressDeliveryFeeModel{}

	// exact per HAWB
	exact := feePerMasterAirwayBill.Div(decimal.NewFromInt(int64(totalHawb)))

	// floor to 2 decimals
	result.FloorPerHawb = exact.RoundFloor(2)

	result.PerHawbFees = make([]decimal.Decimal, totalHawb)
	for i := 0; i < totalHawb; i++ {
		result.PerHawbFees[i] = result.FloorPerHawb
	}

	// how much leftover to distribute
	sum := result.FloorPerHawb.Mul(decimal.NewFromInt(int64(totalHawb)))
	remainder := feePerMasterAirwayBill.Sub(sum)

	// distribute remainder in +0.01 steps
	i := 0
	oneCent := decimal.NewFromFloat(0.01)
	for remainder.GreaterThan(decimal.Zero) {
		result.PerHawbFees[i] = result.PerHawbFees[i].Add(oneCent)
		remainder = remainder.Sub(oneCent)
		i++
	}

	return result
}
