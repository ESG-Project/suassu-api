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
  'FEDERAL',
  'STATE',
  'MUNICIPAL'
);

CREATE TYPE "SpeciesSuccessionalEcology" AS ENUM (
  'P',
  'IS',
  'S',
  'C',
  'LS',
  'MS',
  'AS'
);

CREATE TYPE "SpeciesHabit" AS ENUM (
  'ARB',
  'ANF',
  'ARV',
  'EME FIX',
  'FLU FIX',
  'FLU LIV',
  'HERB',
  'PAL',
  'TREP'
);

-- Tabela Species
CREATE TABLE species (
  id varchar(36) PRIMARY KEY,
  scientific_name varchar(255) NOT NULL UNIQUE,
  family varchar(255) NOT NULL,
  popular_name varchar(255),
  habit "SpeciesHabit",
  created_at timestamp NOT NULL DEFAULT now(),
  updated_at timestamp NOT NULL
);

-- Tabela SpeciesLegislation (species_legislations)
CREATE TABLE species_legislations (
  id varchar(36) PRIMARY KEY,
  law_scope "LawScope" NOT NULL,
  law_id varchar(100),
  is_law_active boolean NOT NULL DEFAULT true,
  species_form_factor numeric NOT NULL,
  is_species_protected boolean NOT NULL DEFAULT false,
  species_threat_status "ThreatStatus" NOT NULL,
  species_origin "OriginType" NOT NULL,
  successional_ecology "SpeciesSuccessionalEcology" NOT NULL,
  species_id varchar(36),
  created_at timestamp NOT NULL DEFAULT now(),
  updated_at timestamp NOT NULL,
  FOREIGN KEY (species_id) REFERENCES species (id)
);

CREATE INDEX idx_species_legislations_species_id ON species_legislations (species_id);

