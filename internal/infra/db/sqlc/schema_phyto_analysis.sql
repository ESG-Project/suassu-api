-- Apenas para o sqlc entender tipos (não roda no banco).
CREATE TABLE phyto_analysis (
  id varchar(36) PRIMARY KEY,
  title varchar(255) NOT NULL,
  initial_date timestamp NOT NULL,
  portion_quantity integer NOT NULL,
  portion_area numeric NOT NULL,
  total_area numeric NOT NULL,
  sampled_area numeric NOT NULL,
  description varchar(500),
  project_id varchar(36) NOT NULL,
  created_at timestamp NOT NULL DEFAULT now(),
  updated_at timestamp NOT NULL,
  FOREIGN KEY (project_id) REFERENCES "Project" (id)
);

CREATE INDEX idx_phyto_analysis_project_id ON phyto_analysis (project_id);

