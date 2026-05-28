-- Apenas para o sqlc entender tipos (não roda no banco).
CREATE TABLE specimen (
  id varchar(36) PRIMARY KEY,
  portion varchar(50) NOT NULL,
  height numeric NOT NULL,
  cap1 numeric NOT NULL,
  cap2 numeric,
  cap3 numeric,
  cap4 numeric,
  cap5 numeric,
  cap6 numeric,
  register_date timestamp NOT NULL,
  phyto_analysis_id varchar(36) NOT NULL,
  specie_id varchar(36) NOT NULL,
  created_at timestamp NOT NULL DEFAULT now(),
  updated_at timestamp NOT NULL,
  FOREIGN KEY (phyto_analysis_id) REFERENCES phyto_analysis (id),
  FOREIGN KEY (specie_id) REFERENCES species (id)
);

