-- Apenas para o sqlc entender tipos (não roda no banco).

-- Enum
CREATE TYPE species_change_status AS ENUM (
  'PENDING',
  'APPROVED',
  'REFUSED'
);

-- Tabela SpecieChange
CREATE TABLE species_changes (
  id varchar(36) PRIMARY KEY,
  field_changed varchar(255) NOT NULL,
  new_value varchar(255) NOT NULL,
  old_value varchar(255) NOT NULL,
  comment varchar(500) NOT NULL,
  status species_change_status NOT NULL,
  refuse_reason varchar(255),
  solicitation_date timestamp NOT NULL DEFAULT now(),
  evaluation_date timestamp NOT NULL,
  specie_id varchar(36) NOT NULL,
  solicitation_user_id varchar(36) NOT NULL,
  evaluation_user_id varchar(36),
  FOREIGN KEY (specie_id) REFERENCES species (id),
  FOREIGN KEY (solicitation_user_id) REFERENCES "User" (id)
);

