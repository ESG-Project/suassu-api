package utils

import (
	"database/sql"
	"fmt"
	"strconv"

	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

// ToNullString converte *string para sql.NullString
func ToNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

// FromNullString converte sql.NullString para *string
func FromNullString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	str := ns.String
	return &str
}

// ToNullInt64 converte *int64 para sql.NullInt64 (para futuras necessidades)
func ToNullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

// FromNullInt64 converte sql.NullInt64 para *int64
func FromNullInt64(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int64
}

// ToNullFloat64 converte *float64 para sql.NullFloat64
func ToNullFloat64(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

// FromNullFloat64 converte sql.NullFloat64 para *float64
func FromNullFloat64(nf sql.NullFloat64) *float64 {
	if !nf.Valid {
		return nil
	}
	return &nf.Float64
}

// Float64ToString converte float64 para string (para tipos numeric do PostgreSQL via sqlc)
func Float64ToString(f float64) string {
	return fmt.Sprintf("%v", f)
}

// StringToFloat64 converte string para float64 (para tipos numeric do PostgreSQL via sqlc)
func StringToFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// NullStringToNullFloat64 converte sql.NullString para *float64
func NullStringToNullFloat64(ns sql.NullString) *float64 {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	f, err := strconv.ParseFloat(ns.String, 64)
	if err != nil {
		return nil
	}
	return &f
}

// Float64PtrToString converte *float64 para sql.NullString
func Float64PtrToString(f *float64) sql.NullString {
	if f == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: fmt.Sprintf("%v", *f), Valid: true}
}

// ToNullSpeciesHabit converte *string para sqlc.NullSpeciesHabit
func ToNullSpeciesHabit(s *string) sqlc.NullSpeciesHabit {
	if s == nil {
		return sqlc.NullSpeciesHabit{Valid: false}
	}
	return sqlc.NullSpeciesHabit{SpeciesHabit: sqlc.SpeciesHabit(*s), Valid: true}
}

// FromNullSpeciesHabit converte sqlc.NullSpeciesHabit para *string
func FromNullSpeciesHabit(nh sqlc.NullSpeciesHabit) *string {
	if !nh.Valid {
		return nil
	}
	str := string(nh.SpeciesHabit)
	return &str
}

// StringToNullString converte string para sql.NullString
func StringToNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}
