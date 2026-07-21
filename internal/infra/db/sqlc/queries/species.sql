-- name: CreateSpecies :one
INSERT INTO public.species (
    id,
    scientific_name,
    family,
    popular_name,
    habit,
    status,
    version,
    created_by,
    enterprise_id,
    parent_id,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: CreateSpeciesLegislation :one
INSERT INTO public.species_legislations (
    id,
    law_scope,
    law_id,
    is_law_active,
    species_form_factor,
    is_species_protected,
    species_threat_status,
    species_origin,
    successional_ecology,
    species_id,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetSpeciesByID :one
SELECT
    s.id,
    s.scientific_name,
    s.family,
    s.popular_name,
    s.habit,
    s.status,
    s.version,
    s.created_by,
    s.enterprise_id,
    s.parent_id,
    s.created_at,
    s.updated_at
FROM public.species s
WHERE s.id = $1
LIMIT 1;

-- name: GetSpeciesByScientificName :one
SELECT
    s.id,
    s.scientific_name,
    s.family,
    s.popular_name,
    s.habit,
    s.status,
    s.version,
    s.created_by,
    s.enterprise_id,
    s.parent_id,
    s.created_at,
    s.updated_at
FROM public.species s
WHERE s.scientific_name = $1
LIMIT 1;

-- GetNextSpeciesVersion retorna a próxima versão para um nome científico
-- (maior versão existente + 1), independentemente do status.
-- name: GetNextSpeciesVersion :one
SELECT COALESCE(MAX(s.version), 0) + 1 AS next_version
FROM public.species s
WHERE s.scientific_name = $1;

-- CountSpeciesLineage conta todas as versões da linhagem (árvore) à qual a
-- espécie $1 pertence: sobe até a raiz (parent_id IS NULL) e desce por todos os
-- descendentes. É a base do cálculo da próxima versão (count + 1), independente
-- de mudanças no nome científico.
-- name: CountSpeciesLineage :one
WITH RECURSIVE up AS (
    SELECT sp.id AS id, sp.parent_id AS parent_id
    FROM public.species sp
    WHERE sp.id = $1
    UNION
    SELECT s.id AS id, s.parent_id AS parent_id
    FROM public.species s
    JOIN up ON s.id = up.parent_id
),
root AS (
    SELECT up.id AS id FROM up WHERE up.parent_id IS NULL LIMIT 1
),
down AS (
    SELECT sp2.id AS id, sp2.parent_id AS parent_id
    FROM public.species sp2
    WHERE sp2.id = (SELECT root.id FROM root)
    UNION
    SELECT s.id AS id, s.parent_id AS parent_id
    FROM public.species s
    JOIN down ON s.parent_id = down.id
)
SELECT COUNT(*)::int AS lineage_count FROM down;

-- ListSpecies retorna o catálogo oficial: apenas a versão APROVADA mais
-- recente de cada nome científico (visível globalmente para todas as empresas).
-- name: ListSpecies :many
SELECT DISTINCT ON (s.scientific_name)
    s.id,
    s.scientific_name,
    s.family,
    s.popular_name,
    s.habit,
    s.status,
    s.version,
    s.created_by,
    s.enterprise_id,
    s.parent_id,
    s.created_at,
    s.updated_at
FROM public.species s
WHERE s.status = 'APPROVED'
ORDER BY s.scientific_name ASC, s.version DESC
LIMIT $1 OFFSET $2;

-- ListVisibleSpecies retorna as espécies visíveis para uma empresa:
-- todas as APROVADAS (globais) + as próprias PENDING/REFUSED da empresa.
-- name: ListVisibleSpecies :many
SELECT
    s.id,
    s.scientific_name,
    s.family,
    s.popular_name,
    s.habit,
    s.status,
    s.version,
    s.created_by,
    s.enterprise_id,
    s.parent_id,
    s.created_at,
    s.updated_at
FROM public.species s
WHERE s.status = 'APPROVED' OR s.enterprise_id = $1
ORDER BY s.scientific_name ASC, s.version DESC
LIMIT $2 OFFSET $3;

-- ListPendingSpecies retorna todas as espécies pendentes (fila do super-admin).
-- name: ListPendingSpecies :many
SELECT
    s.id,
    s.scientific_name,
    s.family,
    s.popular_name,
    s.habit,
    s.status,
    s.version,
    s.created_by,
    s.enterprise_id,
    s.parent_id,
    s.created_at,
    s.updated_at
FROM public.species s
WHERE s.status = 'PENDING'
ORDER BY s.created_at ASC
LIMIT $1 OFFSET $2;

-- name: GetSpeciesLegislationsBySpeciesID :many
SELECT
    sl.id,
    sl.law_scope,
    sl.law_id,
    sl.is_law_active,
    sl.species_form_factor,
    sl.is_species_protected,
    sl.species_threat_status,
    sl.species_origin,
    sl.successional_ecology,
    sl.species_id,
    sl.created_at,
    sl.updated_at
FROM public.species_legislations sl
WHERE sl.species_id = $1
ORDER BY sl.created_at DESC;

-- name: UpdateSpecies :exec
UPDATE public.species
SET
    scientific_name = $2,
    family = $3,
    popular_name = $4,
    habit = $5,
    updated_at = $6
WHERE id = $1;

-- UpdateSpeciesStatus altera o status da espécie (aprovar/recusar).
-- name: UpdateSpeciesStatus :exec
UPDATE public.species
SET
    status = $2,
    updated_at = $3
WHERE id = $1;

-- name: UpdateSpeciesLegislation :exec
UPDATE public.species_legislations
SET
    law_scope = $2,
    law_id = $3,
    is_law_active = $4,
    species_form_factor = $5,
    is_species_protected = $6,
    species_threat_status = $7,
    species_origin = $8,
    successional_ecology = $9,
    updated_at = $10
WHERE id = $1;

-- name: DeleteSpeciesLegislation :exec
DELETE FROM public.species_legislations
WHERE id = $1;
