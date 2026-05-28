-- name: CreateSpecies :one
INSERT INTO public.species (
    id,
    scientific_name,
    family,
    popular_name,
    habit,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
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
    s.created_at,
    s.updated_at
FROM public.species s
WHERE s.scientific_name = $1
LIMIT 1;

-- name: ListSpecies :many
SELECT 
    s.id,
    s.scientific_name,
    s.family,
    s.popular_name,
    s.habit,
    s.created_at,
    s.updated_at
FROM public.species s
ORDER BY s.scientific_name ASC
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

