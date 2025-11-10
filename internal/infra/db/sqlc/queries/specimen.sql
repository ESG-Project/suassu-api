-- name: CreateSpecimen :one
INSERT INTO public.specimen (
    id,
    portion,
    height,
    cap1,
    cap2,
    cap3,
    cap4,
    cap5,
    cap6,
    average_dap,
    basal_area,
    volume,
    register_date,
    phyto_analysis_id,
    specie_id,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING *;

-- name: GetSpecimenByID :one
SELECT 
    sp.id,
    sp.portion,
    sp.height,
    sp.cap1,
    sp.cap2,
    sp.cap3,
    sp.cap4,
    sp.cap5,
    sp.cap6,
    sp.average_dap,
    sp.basal_area,
    sp.volume,
    sp.register_date,
    sp.phyto_analysis_id,
    sp.specie_id,
    sp.created_at,
    sp.updated_at,
    s.scientific_name,
    s.family,
    s.popular_name
FROM public.specimen sp
INNER JOIN public.species s ON sp.specie_id = s.id
WHERE sp.id = $1
LIMIT 1;

-- name: ListSpecimensByPhytoAnalysis :many
SELECT 
    sp.id,
    sp.portion,
    sp.height,
    sp.cap1,
    sp.cap2,
    sp.cap3,
    sp.cap4,
    sp.cap5,
    sp.cap6,
    sp.average_dap,
    sp.basal_area,
    sp.volume,
    sp.register_date,
    sp.phyto_analysis_id,
    sp.specie_id,
    sp.created_at,
    sp.updated_at,
    s.scientific_name,
    s.family,
    s.popular_name
FROM public.specimen sp
INNER JOIN public.species s ON sp.specie_id = s.id
WHERE sp.phyto_analysis_id = $1
ORDER BY sp.portion ASC, sp.created_at ASC;

-- name: UpdateSpecimen :exec
UPDATE public.specimen
SET
    portion = $2,
    height = $3,
    cap1 = $4,
    cap2 = $5,
    cap3 = $6,
    cap4 = $7,
    cap5 = $8,
    cap6 = $9,
    average_dap = $10,
    basal_area = $11,
    volume = $12,
    register_date = $13,
    specie_id = $14,
    updated_at = $15
WHERE id = $1;

-- name: DeleteSpecimen :exec
DELETE FROM public.specimen
WHERE id = $1;

-- name: CountSpecimensByPhytoAnalysis :one
SELECT COUNT(*) as total
FROM public.specimen
WHERE phyto_analysis_id = $1;

