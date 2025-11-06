-- name: CreateSpeciesLegislation :one
INSERT INTO public.species_details (
    id,
    law_scope,
    law_id,
    is_law_active,
    species_form_factor,
    is_species_protected,
    species_threat_status,
    successional_ecology,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: CreateSpecies :one
INSERT INTO public.species (
    id,
    scientific_name,
    family,
    popular_name,
    species_detail_id,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetSpeciesByID :one
SELECT 
    s.id,
    s.scientific_name,
    s.family,
    s.popular_name,
    s.species_detail_id,
    s.created_at,
    s.updated_at,
    sd.law_scope,
    sd.law_id,
    sd.is_law_active,
    sd.species_form_factor,
    sd.is_species_protected,
    sd.species_threat_status,
    sd.successional_ecology,
    sd.created_at AS detail_created_at,
    sd.updated_at AS detail_updated_at
FROM public.species s
INNER JOIN public.species_details sd ON s.species_detail_id = sd.id
WHERE s.id = $1
LIMIT 1;

-- name: GetSpeciesByScientificName :one
SELECT 
    s.id,
    s.scientific_name,
    s.family,
    s.popular_name,
    s.species_detail_id,
    s.created_at,
    s.updated_at,
    sd.law_scope,
    sd.law_id,
    sd.is_law_active,
    sd.species_form_factor,
    sd.is_species_protected,
    sd.species_threat_status,
    sd.successional_ecology,
    sd.created_at AS detail_created_at,
    sd.updated_at AS detail_updated_at
FROM public.species s
INNER JOIN public.species_details sd ON s.species_detail_id = sd.id
WHERE s.scientific_name = $1
LIMIT 1;

-- name: ListSpecies :many
SELECT 
    s.id,
    s.scientific_name,
    s.family,
    s.popular_name,
    s.species_detail_id,
    s.created_at,
    s.updated_at,
    sd.law_scope,
    sd.law_id,
    sd.is_law_active,
    sd.species_form_factor,
    sd.is_species_protected,
    sd.species_threat_status,
    sd.successional_ecology
FROM public.species s
INNER JOIN public.species_details sd ON s.species_detail_id = sd.id
ORDER BY s.scientific_name ASC
LIMIT $1 OFFSET $2;

-- name: UpdateSpecies :exec
UPDATE public.species
SET
    scientific_name = $2,
    family = $3,
    popular_name = $4,
    updated_at = $5
WHERE id = $1;

-- name: UpdateSpeciesLegislation :exec
UPDATE public.species_details
SET
    law_scope = $2,
    law_id = $3,
    is_law_active = $4,
    species_form_factor = $5,
    is_species_protected = $6,
    species_threat_status = $7,
    successional_ecology = $8,
    updated_at = $9
WHERE id = $1;

