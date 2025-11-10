-- name: CreatePhytoAnalysis :one
INSERT INTO public.phyto_analysis (
    id,
    title,
    initial_date,
    portion_quantity,
    portion_area,
    total_area,
    sampled_area,
    description,
    project_id,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetPhytoAnalysisByID :one
SELECT 
    pa.id,
    pa.title,
    pa.initial_date,
    pa.portion_quantity,
    pa.portion_area,
    pa.total_area,
    pa.sampled_area,
    pa.description,
    pa.project_id,
    pa.created_at,
    pa.updated_at,
    p.title AS project_title,
    p.cnpj AS project_cnpj,
    p.activity AS project_activity,
    p."clientId" AS project_client_id
FROM public.phyto_analysis pa
INNER JOIN public."Project" p ON pa.project_id = p.id
WHERE pa.id = $1
LIMIT 1;

-- name: ListPhytoAnalysesByProject :many
SELECT 
    pa.id,
    pa.title,
    pa.initial_date,
    pa.portion_quantity,
    pa.portion_area,
    pa.total_area,
    pa.sampled_area,
    pa.description,
    pa.project_id,
    pa.created_at,
    pa.updated_at,
    p.title AS project_title,
    p.cnpj AS project_cnpj,
    p.activity AS project_activity
FROM public.phyto_analysis pa
INNER JOIN public."Project" p ON pa.project_id = p.id
WHERE pa.project_id = $1
ORDER BY pa.initial_date DESC, pa.created_at DESC;

-- name: ListAllPhytoAnalyses :many
SELECT 
    pa.id,
    pa.title,
    pa.initial_date,
    pa.portion_quantity,
    pa.portion_area,
    pa.total_area,
    pa.sampled_area,
    pa.description,
    pa.project_id,
    pa.created_at,
    pa.updated_at,
    p.title AS project_title,
    p.cnpj AS project_cnpj,
    p.activity AS project_activity,
    p."clientId" AS project_client_id
FROM public.phyto_analysis pa
INNER JOIN public."Project" p ON pa.project_id = p.id
ORDER BY pa.initial_date DESC, pa.created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListPhytoAnalysesByEnterprise :many
SELECT 
    pa.id,
    pa.title,
    pa.initial_date,
    pa.portion_quantity,
    pa.portion_area,
    pa.total_area,
    pa.sampled_area,
    pa.description,
    pa.project_id,
    pa.created_at,
    pa.updated_at,
    p.title AS project_title,
    p.cnpj AS project_cnpj,
    p.activity AS project_activity,
    p."clientId" AS project_client_id
FROM public.phyto_analysis pa
INNER JOIN public."Project" p ON pa.project_id = p.id
INNER JOIN public."Client" c ON p."clientId" = c.id
INNER JOIN public."User" u ON c."userId" = u.id
WHERE u."enterpriseId" = $1
ORDER BY pa.initial_date DESC, pa.created_at DESC;

-- name: UpdatePhytoAnalysis :exec
UPDATE public.phyto_analysis
SET
    title = $2,
    initial_date = $3,
    portion_quantity = $4,
    portion_area = $5,
    total_area = $6,
    sampled_area = $7,
    description = $8,
    updated_at = $9
WHERE id = $1;

-- name: DeletePhytoAnalysis :exec
DELETE FROM public.phyto_analysis
WHERE id = $1;

-- name: GetPhytoAnalysisWithSpecimens :many
SELECT 
    pa.id AS phyto_id,
    pa.title AS phyto_title,
    pa.initial_date,
    pa.portion_quantity,
    pa.portion_area,
    pa.total_area,
    pa.sampled_area,
    pa.description AS phyto_description,
    pa.project_id,
    pa.created_at AS phyto_created_at,
    pa.updated_at AS phyto_updated_at,
    p.title AS project_title,
    p.cnpj AS project_cnpj,
    p.activity AS project_activity,
    p."clientId" AS project_client_id,
    sp.id AS specimen_id,
    sp.portion,
    sp.height,
    sp.cap1,
    sp.cap2,
    sp.cap3,
    sp.cap4,
    sp.cap5,
    sp.cap6,
    sp.register_date,
    sp.specie_id,
    s.scientific_name,
    s.family,
    s.popular_name
FROM public.phyto_analysis pa
INNER JOIN public."Project" p ON pa.project_id = p.id
LEFT JOIN public.specimen sp ON sp.phyto_analysis_id = pa.id
LEFT JOIN public.species s ON sp.specie_id = s.id
WHERE pa.id = $1
ORDER BY sp.portion ASC, sp.created_at ASC;

