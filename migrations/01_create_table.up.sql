DROP TABLE IF EXISTS fuel CASCADE;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE fuel (
  id  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  imo int NOT NULL,
  draught float8 ,
  speed float8,
  beaufort float8,
  consumption double precision
);

CREATE INDEX idx_imo ON fuel(imo);
CREATE INDEX idx_draught ON fuel(draught);

