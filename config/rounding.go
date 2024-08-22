package config

import (
	"math/big"
)

// RoundToNearestCent rounds a float64 to the nearest cent using big.Float
func RoundToNearestCent(amount float64) float64 {
	// Create a new big.Float with the float64 amount
	bf := new(big.Float).SetFloat64(amount)
	
	// Multiply by 100 to shift the decimal point
	bf.Mul(bf, big.NewFloat(100))
	
	// Round to the nearest integer
	rounded := new(big.Int)
	bf.Int(rounded)
	
	// Convert the rounded big.Int back to big.Float
	bf.SetInt(rounded)
	
	// Divide by 100 to get the final rounded amount
	bf.Quo(bf, big.NewFloat(100))
	
	// Return the float64 value of the result
	roundedAmount, _ := bf.Float64()
	return roundedAmount
}

