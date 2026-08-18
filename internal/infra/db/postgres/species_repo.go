package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ESG-Project/suassu-api/internal/app/types"
	"github.com/ESG-Project/suassu-api/internal/apperr"
	domainspecies "github.com/ESG-Project/suassu-api/internal/domain/species"
	"github.com/ESG-Project/suassu-api/internal/infra/db/postgres/utils"
	sqlc "github.com/ESG-Project/suassu-api/internal/infra/db/sqlc/gen"
)

type SpeciesRepo struct {
	q  *sqlc.Queries
	db dbtx
}

func NewSpeciesRepoFrom(d dbtx) *SpeciesRepo {
	return &SpeciesRepo{q: sqlc.New(d), db: d}
}

func NewSpeciesRepo(db *sql.DB) *SpeciesRepo {
	return &SpeciesRepo{q: sqlc.New(db), db: db}
}

func (r *SpeciesRepo) CreateSpecies(ctx context.Context, s *domainspecies.Species) error {
	_, err := r.q.CreateSpecies(ctx, sqlc.CreateSpeciesParams{
		ID:             s.ID,
		ScientificName: s.ScientificName,
		Family:         s.Family,
		PopularName:    utils.ToNullString(s.PopularName),
		Habit:          utils.ToNullSpeciesHabit(s.Habit),
		Status:         sqlc.SpeciesStatus(s.Status),
		Version:        s.Version,
		CreatedBy:      utils.ToNullString(s.CreatedBy),
		EnterpriseID:   utils.ToNullString(s.EnterpriseID),
		ParentID:       utils.ToNullString(s.ParentID),
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	})
	return err
}

func (r *SpeciesRepo) CreateLegislation(ctx context.Context, sl *domainspecies.SpeciesLegislation) error {
	_, err := r.q.CreateSpeciesLegislation(ctx, sqlc.CreateSpeciesLegislationParams{
		ID:                  sl.ID,
		LawScope:            sqlc.LawScope(sl.LawScope),
		LawID:               utils.ToNullString(sl.LawID),
		IsLawActive:         sl.IsLawActive,
		SpeciesFormFactor:   utils.Float64ToString(sl.SpeciesFormFactor),
		IsSpeciesProtected:  sl.IsSpeciesProtected,
		SpeciesThreatStatus: sqlc.ThreatStatus(sl.SpeciesThreatStatus),
		SpeciesOrigin:       sqlc.OriginType(sl.SpeciesOrigin),
		SuccessionalEcology: sqlc.SpeciesSuccessionalEcology(sl.SuccessionalEcology),
		SpeciesID:           utils.ToNullString(sl.SpeciesID),
		CreatedAt:           sl.CreatedAt,
		UpdatedAt:           sl.UpdatedAt,
	})
	return err
}

// legislationsFor busca e mapeia as legislações de uma espécie.
func (r *SpeciesRepo) legislationsFor(ctx context.Context, speciesID string) ([]types.LegislationData, error) {
	legislations, err := r.q.GetSpeciesLegislationsBySpeciesID(ctx, utils.StringToNullString(speciesID))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	data := make([]types.LegislationData, 0, len(legislations))
	for _, leg := range legislations {
		formFactor, _ := utils.StringToFloat64(leg.SpeciesFormFactor)
		data = append(data, types.LegislationData{
			ID:                  leg.ID,
			LawScope:            string(leg.LawScope),
			LawID:               utils.FromNullString(leg.LawID),
			IsLawActive:         leg.IsLawActive,
			SpeciesFormFactor:   formFactor,
			IsSpeciesProtected:  leg.IsSpeciesProtected,
			SpeciesThreatStatus: string(leg.SpeciesThreatStatus),
			SpeciesOrigin:       string(leg.SpeciesOrigin),
			SuccessionalEcology: string(leg.SuccessionalEcology),
			SpeciesID:           utils.FromNullString(leg.SpeciesID),
			CreatedAt:           leg.CreatedAt,
			UpdatedAt:           leg.UpdatedAt,
		})
	}
	return data, nil
}

// mapSpeciesRow converte uma linha sqlc + legislações no tipo de aplicação.
func mapSpeciesRow(row sqlc.Species, legislations []types.LegislationData) *types.SpeciesWithLegislation {
	return &types.SpeciesWithLegislation{
		ID:             row.ID,
		ScientificName: row.ScientificName,
		Family:         row.Family,
		PopularName:    utils.FromNullString(row.PopularName),
		Habit:          utils.FromNullSpeciesHabit(row.Habit),
		Status:         string(row.Status),
		Version:        row.Version,
		CreatedBy:      utils.FromNullString(row.CreatedBy),
		EnterpriseID:   utils.FromNullString(row.EnterpriseID),
		ParentID:       utils.FromNullString(row.ParentID),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
		Legislations:   legislations,
	}
}

func (r *SpeciesRepo) GetByID(ctx context.Context, id string) (*types.SpeciesWithLegislation, error) {
	row, err := r.q.GetSpeciesByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "species not found")
		}
		return nil, err
	}

	legislations, err := r.legislationsFor(ctx, row.ID)
	if err != nil {
		return nil, err
	}

	return mapSpeciesRow(row, legislations), nil
}

// GetMapByScientificNames busca várias espécies de uma vez e retorna
// um mapa scientificName -> speciesID. Uma única query SQL.
func (r *SpeciesRepo) GetMapByScientificNames(ctx context.Context, names []string) (map[string]string, error) {
	if len(names) == 0 {
		return make(map[string]string), nil
	}

	unique := make(map[string]struct{}, len(names))
	trimmed := make([]string, 0, len(names))
	for _, n := range names {
		t := strings.TrimSpace(n)
		if t == "" {
			continue
		}
		if _, ok := unique[t]; !ok {
			unique[t] = struct{}{}
			trimmed = append(trimmed, t)
		}
	}
	if len(trimmed) == 0 {
		return make(map[string]string), nil
	}

	placeholders := make([]string, 0, len(trimmed))
	args := make([]interface{}, 0, len(trimmed))
	i := 1
	for _, name := range trimmed {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		args = append(args, name)
		i++
	}

	// Compara pelo nome sem espaços à esquerda/direita para casar com dados legados no banco.
	query := fmt.Sprintf(
		`SELECT id, scientific_name FROM public.species WHERE trim(both from scientific_name) IN (%s)`,
		strings.Join(placeholders, ", "),
	)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string, len(trimmed))
	for rows.Next() {
		var id, scientificName string
		if err := rows.Scan(&id, &scientificName); err != nil {
			return nil, err
		}
		result[strings.TrimSpace(scientificName)] = id
	}

	return result, rows.Err()
}

func (r *SpeciesRepo) GetByScientificName(ctx context.Context, scientificName string) (*types.SpeciesWithLegislation, error) {
	row, err := r.q.GetSpeciesByScientificName(ctx, scientificName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.New(apperr.CodeNotFound, "species not found")
		}
		return nil, err
	}

	legislations, err := r.legislationsFor(ctx, row.ID)
	if err != nil {
		return nil, err
	}

	return mapSpeciesRow(row, legislations), nil
}

// GetNextVersion retorna a próxima versão para um nome científico (maior + 1).
func (r *SpeciesRepo) GetNextVersion(ctx context.Context, scientificName string) (int32, error) {
	return r.q.GetNextSpeciesVersion(ctx, scientificName)
}

// CountLineage conta quantas versões existem na linhagem (árvore ligada por
// parent_id) à qual a espécie informada pertence.
func (r *SpeciesRepo) CountLineage(ctx context.Context, speciesID string) (int32, error) {
	return r.q.CountSpeciesLineage(ctx, speciesID)
}

// mapSpeciesRows mapeia uma lista de linhas, carregando as legislações de cada uma.
func (r *SpeciesRepo) mapSpeciesRows(ctx context.Context, rows []sqlc.Species) ([]*types.SpeciesWithLegislation, error) {
	if len(rows) == 0 {
		return []*types.SpeciesWithLegislation{}, nil
	}

	// Busca as legislações de todas as espécies de uma vez (evita N+1: antes era
	// uma query por espécie, o que estourava o timeout ao listar o catálogo todo).
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	legBySpecies, err := r.legislationsForMany(ctx, ids)
	if err != nil {
		return nil, err
	}

	result := make([]*types.SpeciesWithLegislation, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapSpeciesRow(row, legBySpecies[row.ID]))
	}
	return result, nil
}

// legislationsForMany busca e agrupa as legislações de várias espécies em uma
// única query (WHERE species_id IN (...)), retornando um mapa species_id -> legislações.
func (r *SpeciesRepo) legislationsForMany(ctx context.Context, speciesIDs []string) (map[string][]types.LegislationData, error) {
	result := make(map[string][]types.LegislationData, len(speciesIDs))
	if len(speciesIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, 0, len(speciesIDs))
	args := make([]interface{}, 0, len(speciesIDs))
	for i, id := range speciesIDs {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		SELECT sl.id, sl.law_scope, sl.law_id, sl.is_law_active, sl.species_form_factor,
		       sl.is_species_protected, sl.species_threat_status, sl.species_origin,
		       sl.successional_ecology, sl.species_id, sl.created_at, sl.updated_at
		FROM public.species_legislations sl
		WHERE sl.species_id IN (%s)
		ORDER BY sl.created_at DESC`, strings.Join(placeholders, ", "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id                            string
			lawScope                      string
			lawID                         sql.NullString
			isLawActive                   bool
			formFactorStr                 string
			isSpeciesProtected            bool
			threatStatus, origin, ecology string
			speciesID                     sql.NullString
			createdAt, updatedAt          time.Time
		)
		if err := rows.Scan(
			&id, &lawScope, &lawID, &isLawActive, &formFactorStr,
			&isSpeciesProtected, &threatStatus, &origin, &ecology,
			&speciesID, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		formFactor, _ := utils.StringToFloat64(formFactorStr)
		sid := speciesID.String
		result[sid] = append(result[sid], types.LegislationData{
			ID:                  id,
			LawScope:            lawScope,
			LawID:               utils.FromNullString(lawID),
			IsLawActive:         isLawActive,
			SpeciesFormFactor:   formFactor,
			IsSpeciesProtected:  isSpeciesProtected,
			SpeciesThreatStatus: threatStatus,
			SpeciesOrigin:       origin,
			SuccessionalEcology: ecology,
			SpeciesID:           utils.FromNullString(speciesID),
			CreatedAt:           createdAt,
			UpdatedAt:           updatedAt,
		})
	}
	return result, rows.Err()
}

// List retorna o catálogo oficial (versão aprovada mais recente de cada nome).
func (r *SpeciesRepo) List(ctx context.Context, limit, offset int32) ([]*types.SpeciesWithLegislation, error) {
	rows, err := r.q.ListSpecies(ctx, sqlc.ListSpeciesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	return r.mapSpeciesRows(ctx, rows)
}

// ListVisible retorna as espécies visíveis para uma empresa
// (todas as aprovadas + as próprias pendentes/recusadas).
func (r *SpeciesRepo) ListVisible(ctx context.Context, enterpriseID string, limit, offset int32) ([]*types.SpeciesWithLegislation, error) {
	rows, err := r.q.ListVisibleSpecies(ctx, sqlc.ListVisibleSpeciesParams{
		EnterpriseID: utils.StringToNullString(enterpriseID),
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		return nil, err
	}
	return r.mapSpeciesRows(ctx, rows)
}

// ListVisiblePaged retorna as espécies visíveis da empresa aplicando filtros,
// ordenação e paginação no banco, junto da contagem total (para o grid).
func (r *SpeciesRepo) ListVisiblePaged(ctx context.Context, f types.SpeciesListFilter) ([]*types.SpeciesWithLegislation, int64, error) {
	// WHERE base: aprovadas globais + próprias (qualquer status) da empresa.
	where := []string{"(s.status = 'APPROVED' OR s.enterprise_id = $1)"}
	args := []interface{}{f.EnterpriseID}

	addLike := func(col, val string) {
		args = append(args, "%"+val+"%")
		where = append(where, fmt.Sprintf("%s ILIKE $%d", col, len(args)))
	}
	if f.Q != "" {
		args = append(args, "%"+f.Q+"%")
		n := len(args)
		where = append(where, fmt.Sprintf(
			"(s.scientific_name ILIKE $%d OR s.family ILIKE $%d OR COALESCE(s.popular_name,'') ILIKE $%d)", n, n, n))
	}
	if f.ScientificName != "" {
		addLike("s.scientific_name", f.ScientificName)
	}
	if f.Family != "" {
		addLike("s.family", f.Family)
	}
	if f.PopularName != "" {
		addLike("COALESCE(s.popular_name,'')", f.PopularName)
	}
	if f.Status != "" {
		args = append(args, f.Status)
		where = append(where, fmt.Sprintf("s.status::text = $%d", len(args)))
	}
	if f.VersionMin != nil {
		args = append(args, *f.VersionMin)
		where = append(where, fmt.Sprintf("s.version >= $%d", len(args)))
	}
	if f.VersionMax != nil {
		args = append(args, *f.VersionMax)
		where = append(where, fmt.Sprintf("s.version <= $%d", len(args)))
	}
	whereSQL := strings.Join(where, " AND ")

	// Contagem total (mesmo WHERE, sem paginação) para o grid.
	var total int64
	if err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM public.species s WHERE "+whereSQL, args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*types.SpeciesWithLegislation{}, 0, nil
	}

	// ORDER BY com whitelist de colunas (o sort vem do cliente; nunca interpolar cru).
	orderCol := map[string]string{
		"scientificName": "s.scientific_name",
		"family":         "s.family",
		"popularName":    "s.popular_name",
		"version":        "s.version",
		"status":         "s.status",
	}[f.Sort]
	if orderCol == "" {
		orderCol = "s.scientific_name"
	}
	order := "ASC"
	if strings.EqualFold(f.Order, "desc") {
		order = "DESC"
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	args = append(args, limit)
	limitPos := len(args)
	args = append(args, f.Offset)
	offsetPos := len(args)

	listSQL := fmt.Sprintf(`
		SELECT s.id, s.scientific_name, s.family, s.popular_name, s.habit,
		       s.status, s.version, s.created_by, s.enterprise_id, s.parent_id,
		       s.created_at, s.updated_at
		FROM public.species s
		WHERE %s
		ORDER BY %s %s, s.id ASC
		LIMIT $%d OFFSET $%d`, whereSQL, orderCol, order, limitPos, offsetPos)

	rows, err := r.db.QueryContext(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// Pré-aloca pelo tamanho real do resultado (min entre limit e total) para não
	// reservar memória à toa quando o limit é alto (ex.: export sem paginação).
	capHint := total
	if int64(limit) < capHint {
		capHint = int64(limit)
	}
	list := make([]sqlc.Species, 0, capHint)
	for rows.Next() {
		var row sqlc.Species
		if err := rows.Scan(
			&row.ID, &row.ScientificName, &row.Family, &row.PopularName, &row.Habit,
			&row.Status, &row.Version, &row.CreatedBy, &row.EnterpriseID, &row.ParentID,
			&row.CreatedAt, &row.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	mapped, err := r.mapSpeciesRows(ctx, list)
	if err != nil {
		return nil, 0, err
	}
	return mapped, total, nil
}

// ListPending retorna todas as espécies pendentes (fila do super-admin).
func (r *SpeciesRepo) ListPending(ctx context.Context, limit, offset int32) ([]*types.SpeciesWithLegislation, error) {
	rows, err := r.q.ListPendingSpecies(ctx, sqlc.ListPendingSpeciesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	return r.mapSpeciesRows(ctx, rows)
}

func (r *SpeciesRepo) UpdateSpecies(ctx context.Context, s *domainspecies.Species) error {
	return r.q.UpdateSpecies(ctx, sqlc.UpdateSpeciesParams{
		ID:             s.ID,
		ScientificName: s.ScientificName,
		Family:         s.Family,
		PopularName:    utils.ToNullString(s.PopularName),
		Habit:          utils.ToNullSpeciesHabit(s.Habit),
		UpdatedAt:      s.UpdatedAt,
	})
}

// UpdateStatus altera o status da espécie (aprovar/recusar).
func (r *SpeciesRepo) UpdateStatus(ctx context.Context, id, status string) error {
	return r.q.UpdateSpeciesStatus(ctx, sqlc.UpdateSpeciesStatusParams{
		ID:        id,
		Status:    sqlc.SpeciesStatus(status),
		UpdatedAt: time.Now(),
	})
}

func (r *SpeciesRepo) UpdateLegislation(ctx context.Context, sl *domainspecies.SpeciesLegislation) error {
	return r.q.UpdateSpeciesLegislation(ctx, sqlc.UpdateSpeciesLegislationParams{
		ID:                  sl.ID,
		LawScope:            sqlc.LawScope(sl.LawScope),
		LawID:               utils.ToNullString(sl.LawID),
		IsLawActive:         sl.IsLawActive,
		SpeciesFormFactor:   utils.Float64ToString(sl.SpeciesFormFactor),
		IsSpeciesProtected:  sl.IsSpeciesProtected,
		SpeciesThreatStatus: sqlc.ThreatStatus(sl.SpeciesThreatStatus),
		SpeciesOrigin:       sqlc.OriginType(sl.SpeciesOrigin),
		SuccessionalEcology: sqlc.SpeciesSuccessionalEcology(sl.SuccessionalEcology),
		UpdatedAt:           sl.UpdatedAt,
	})
}
