-- Apenas para o sqlc entender tipos (não roda no banco).

-- Enums
CREATE TYPE "ThreatStatus" AS ENUM (
  'LC',
  'CR',
  'NT',
  'EN',
  'VU'
);

CREATE TYPE "OriginType" AS ENUM (
  'EX',
  'EXI',
  'N'
);

CREATE TYPE "LawScope" AS ENUM (
  'Federal',
  'State',
  'Municipal'
);

-- Tabela SpeciesLegislation (species_details)
CREATE TABLE species_details (
  id varchar(36) PRIMARY KEY,
  law_scope "LawScope" NOT NULL,
  law_id varchar(100) NOT NULL,
  is_law_active boolean NOT NULL DEFAULT true,
  species_form_factor numeric NOT NULL,
  is_species_protected boolean NOT NULL DEFAULT false,
  species_threat_status "ThreatStatus" NOT NULL,
  successional_ecology "OriginType" NOT NULL,
  created_at timestamp NOT NULL DEFAULT now(),
  updated_at timestamp NOT NULL
);

-- Tabela Species
CREATE TABLE species (
  id varchar(36) PRIMARY KEY,
  scientific_name varchar(255) NOT NULL UNIQUE,
  family varchar(255) NOT NULL,
  popular_name varchar(255),
  species_detail_id varchar(36) NOT NULL UNIQUE,
  created_at timestamp NOT NULL DEFAULT now(),
  updated_at timestamp NOT NULL,
  FOREIGN KEY (species_detail_id) REFERENCES species_details (id)
);

CREATE INDEX idx_species_species_detail_id ON species (species_detail_id);

