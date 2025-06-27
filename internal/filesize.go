package internal

import (
	"errors"
	"fmt"
)

// Unit is defined storage unit
type Unit uint64

const (
	// B is Byte
	B Unit = 1
	// KB is Kilobyte
	KB = B << 10
	// MB is Megabyte
	MB = KB << 10
	// GB is GigaByte
	GB = MB << 10
	// TB is Terabyte
	TB = GB << 10
	// PB is PetaByte
	PB = TB << 10
	// EB is Extrabyte
	EB = PB << 10
)

func (u Unit) isValid() bool {
	switch u {
	case B, KB, MB, GB, TB, PB, EB:
		return true
	}
	return false
}

func (u Unit) toString() string {
	var unitStr string
	switch u {
	case B:
		unitStr = "B"
	case KB:
		unitStr = "KB"
	case MB:
		unitStr = "MB"
	case GB:
		unitStr = "GB"
	case TB:
		unitStr = "TB"
	case PB:
		unitStr = "PB"
	case EB:
		unitStr = "EB"
	}
	return unitStr
}

// Byte is defined as uint64, which is equal 8 bit
type Byte uint64

// Convert converts byte to other unit, it will return err if input unit is invalid
func (b Byte) Convert(unit Unit) (float64, error) {
	if !unit.isValid() {
		return 0, errors.New("Invalid unit")
	}
	return float64(b) / float64(unit), nil
}

// ConvertToString converts byte to other unit then convert this to string, it will return err if input unit is invalid
// Result will be corrected to 1 decimal number
func (b Byte) ConvertToString(unit Unit) (string, error) {
	res, err := b.Convert(unit)
	if err != nil {
		return "", err
	}
	// Special case for Byte, don't show decimal number
	if unit == B {
		return fmt.Sprintf("%d%s", uint64(res), unit.toString()), nil
	}
	return fmt.Sprintf("%.1f%s", res, unit.toString()), nil
}

// ToString converts byte to a string which is human-readable
func (b Byte) ToString() (string, error) {
	var unit Unit
	switch {
	case b >= Byte(EB):
		unit = EB
	case b >= Byte(PB):
		unit = PB
	case b >= Byte(TB):
		unit = TB
	case b >= Byte(GB):
		unit = GB
	case b >= Byte(MB):
		unit = MB
	case b >= Byte(KB):
		unit = KB
	default:
		unit = B
	}
	res, err := b.ConvertToString(unit)
	return res, err
}
