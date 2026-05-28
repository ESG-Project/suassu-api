-- Schema para refresh tokens (apenas para o sqlc entender tipos).
CREATE TABLE refresh_token (
  id text PRIMARY KEY,
  user_id text NOT NULL,
  token_hash text NOT NULL,
  expires_at timestamp with time zone NOT NULL,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  revoked_at timestamp with time zone,
  FOREIGN KEY (user_id) REFERENCES "User" (id) ON DELETE CASCADE
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_token (user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_token (token_hash);
