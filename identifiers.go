package identifiers

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/cmarkh/errs"
)

//reference docs: https://www.cusip.com/pdf/CUSIP_Intro_03.14.11.pdf

// FIGI takes a string containing an FIGI but possibly more than just the FIGI, strips it, validates it is a real FIGI, and returns just the FIGI
// An FIGI is a 12-character code that identifies a financial security.
func FIGI(figi string) (string, error) {
	if len(figi) < 12 {
		err := fmt.Errorf("FIGI must be at least 12 characters long. Provided: %s", figi)
		return "", err
	}
	figi = figi[0:12]

	ascii, err := ascii(figi[3:12])
	if err != nil {
		return "", err
	}

	if !ValidLuhn(ascii) {
		err := fmt.Errorf("FIGI failed the Luhn verification. Provided: %s", figi)
		return "", err
	}

	return figi, nil
}

// ISIN takes a string containing an ISIN but possibly more than just the ISIN, strips it, validates it is a real ISIN, and returns just the ISIN
// An ISIN is a 12-character code that identifies a financial security.
func ISIN(isin string) (string, error) {
	if len(isin) < 12 {
		err := fmt.Errorf("ISIN must be at least 12 characters long. Provided: %s", isin)
		return "", err
	}
	isin = isin[0:12]

	if isin[:3] == "BBG" { //just accept Bloomberg ID style
		return isin, nil
	}

	ascii, err := ascii(isin)
	if err != nil {
		if strings.HasSuffix(fmt.Sprint(err), "value out of range") {
			return isin, nil
		}

		return "", err
	}

	if !ValidLuhn(ascii) {
		err := fmt.Errorf("ISIN failed the Luhn verification. Provided: %s", isin)
		return "", err
	}

	return isin, nil
}

// CUSIP takes a string containing an CUSIP but possibly more than just the CUSIP, strips it, validates it is a real CUSIP, and returns just the CUSIP
// An CUSIP is a 9-character code that identifies a financial security.
func CUSIP(cusip string) (string, error) {
	if len(cusip) < 8 {
		err := fmt.Errorf("CUSIP must be at least 8 characters long. Provided: %s", cusip)
		return "", err
	}
	if len(cusip) == 8 {
		cusip = cusip[0:8]
	} else {
		cusip = cusip[0:9]
	}

	if cusip[:2] == "BL" { //just accept Bloomberg ID style
		return cusip, nil
	}

	if !Modulus10DoubleAddDouble(cusip) {
		err := fmt.Errorf("CUSIP failed the Modulus 10 Double Add Double verification. Provided: %s", cusip)
		errs.Log(err)
		return "", err
	}

	return cusip, nil
}

// Ascii converts the letters in the string to their ascii numbers
func ascii(str string) (ascii int, err error) {
	var new string
	for _, char := range str {
		if !unicode.IsDigit(char) {
			new += fmt.Sprint(int(char) - 55)
			continue
		}
		new += fmt.Sprintf("%c", char)
	}

	ascii, err = strconv.Atoi(new)
	if err != nil {
		return
	}
	return
}

// ValidLuhn check number is valid or not based on Luhn algorithm
func ValidLuhn(number int) bool {
	checksum := func(number int) int {
		var luhn int

		for i := 0; number > 0; i++ {
			cur := number % 10

			if i%2 == 0 { // even
				cur = cur * 2
				if cur > 9 {
					cur = cur%10 + cur/10
				}
			}

			luhn += cur
			number = number / 10
		}
		return luhn % 10
	}

	return (number%10+checksum(number/10))%10 == 0
}

// Modulus10DoubleAddDouble is the check digit algorithm for CUSIP verification
func Modulus10DoubleAddDouble(cusip string) bool {
	if len(cusip) != 9 {
		errs.Log(fmt.Errorf("CUSIP missing check digit. Assuming Passed. Provided: %s", cusip))
		return true
	}
	checkdigit := cusip[8] - '0'

	var sum int64
	for i, char := range cusip[:8] { //last digit is the check digit so skip it
		var intChar int64

		if !unicode.IsDigit(char) {
			intChar = int64(char - 'A' + 10) //The letter A will be 10; and the value of each subsequent letter will be the preceding letter’s value incremented by 1
		} else {
			intChar = int64(char - '0')
		}

		if i%2 != 0 { //if char index in cusip is odd, double it
			intChar *= 2
		}

		sum += intChar % 10
		for intChar = int64(intChar / 10); intChar != 0; intChar = int64(intChar / 10) { //add the individual digits, not whole number
			sum += intChar
		}
	}

	return int64(checkdigit) == (10 - sum%10) //the check num = 10 - the last digit of the sum
}
