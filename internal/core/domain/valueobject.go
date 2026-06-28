package domain

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type Money struct {
	cents int64
}

func NewMoneyFromFloat(amount float64) Money {
	return Money{cents: int64(math.Round(amount * 100))}
}

func (m Money) Float64() float64 {
	return float64(m.cents) / 100
}

func (m Money) Add(other Money) Money {
	return Money{cents: m.cents + other.cents}
}

func (m Money) Sub(other Money) Money {
	return Money{cents: m.cents - other.cents}
}

func (m Money) MulQty(qty int) Money {
	return Money{cents: m.cents * int64(qty)}
}

func (m Money) AtLeastZero() Money {
	if m.cents < 0 {
		return Money{}
	}
	return m
}

func (m Money) IsPositive() bool {
	return m.cents > 0
}

func (m Money) GreaterOrEqual(other Money) bool {
	return m.cents >= other.cents
}

func (m Money) String() string {
	return fmt.Sprintf("%.2f", m.Float64())
}

func (m Money) MarshalJSON() ([]byte, error) {
	return []byte(m.String()), nil
}

func (m *Money) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	if s == "null" || s == "" {
		m.cents = 0
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("valor monetário inválido: %w", err)
	}
	m.cents = int64(math.Round(f * 100))
	return nil
}

type CpfCnpj struct {
	digits string
}

var nonDigits = regexp.MustCompile(`\D`)

func NewCpfCnpj(raw string) (CpfCnpj, error) {
	digits := nonDigits.ReplaceAllString(raw, "")
	if digits == "" {
		return CpfCnpj{}, nil
	}
	switch len(digits) {
	case 11:
		if !validCPF(digits) {
			return CpfCnpj{}, fmt.Errorf("CPF inválido: %w", ErrValidation)
		}
	case 14:
		if !validCNPJ(digits) {
			return CpfCnpj{}, fmt.Errorf("CNPJ inválido: %w", ErrValidation)
		}
	default:
		return CpfCnpj{}, fmt.Errorf("documento deve ter 11 (CPF) ou 14 (CNPJ) dígitos: %w", ErrValidation)
	}
	return CpfCnpj{digits: digits}, nil
}

func NewCpfCnpjUnchecked(raw string) CpfCnpj {
	return CpfCnpj{digits: nonDigits.ReplaceAllString(raw, "")}
}

func (d CpfCnpj) String() string {
	return d.digits
}

func (d CpfCnpj) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.digits + `"`), nil
}

func (d *CpfCnpj) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	parsed, err := NewCpfCnpj(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

func allSameDigit(s string) bool {
	for i := 1; i < len(s); i++ {
		if s[i] != s[0] {
			return false
		}
	}
	return true
}

func checkDigit(digits string, weights []int) int {
	sum := 0
	for i, w := range weights {
		sum += int(digits[i]-'0') * w
	}
	rest := sum % 11
	if rest < 2 {
		return 0
	}
	return 11 - rest
}

func validCPF(d string) bool {
	if allSameDigit(d) {
		return false
	}
	dv1 := checkDigit(d[:9], []int{10, 9, 8, 7, 6, 5, 4, 3, 2})
	dv2 := checkDigit(d[:9]+strconv.Itoa(dv1), []int{11, 10, 9, 8, 7, 6, 5, 4, 3, 2})
	return int(d[9]-'0') == dv1 && int(d[10]-'0') == dv2
}

func validCNPJ(d string) bool {
	if allSameDigit(d) {
		return false
	}
	dv1 := checkDigit(d[:12], []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	dv2 := checkDigit(d[:12]+strconv.Itoa(dv1), []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	return int(d[12]-'0') == dv1 && int(d[13]-'0') == dv2
}

type Phone struct {
	digits string
}

func NewPhone(raw string) (Phone, error) {
	digits := normalizeBrazilianPhoneDigits(raw)
	if digits == "" {
		return Phone{}, nil
	}
	if len(digits) != 10 && len(digits) != 11 {
		return Phone{}, fmt.Errorf("telefone deve ter DDD + número (10 ou 11 dígitos): %w", ErrValidation)
	}
	return Phone{digits: digits}, nil
}

func NewPhoneUnchecked(raw string) Phone {
	return Phone{digits: normalizeBrazilianPhoneDigits(raw)}
}

func normalizeBrazilianPhoneDigits(raw string) string {
	digits := nonDigits.ReplaceAllString(raw, "")
	if len(digits) == 12 || len(digits) == 13 {
		digits = strings.TrimPrefix(digits, "55")
	}
	return digits
}

func (p Phone) String() string {
	return p.digits
}

func (p Phone) MarshalJSON() ([]byte, error) {
	return []byte(`"` + p.digits + `"`), nil
}

func (p *Phone) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	parsed, err := NewPhone(s)
	if err != nil {
		return err
	}
	*p = parsed
	return nil
}
