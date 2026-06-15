-- 1. On active l'extension qui permet de découper le texte en blocs de 3 lettres
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 2. On crée l'index GIN adapté aux recherches partielles (ILIKE)
CREATE INDEX IF NOT EXISTS invoices_client_name_idx ON invoices USING GIN (client_name gin_trgm_ops);

-- 3. Pour le statut, un index standard (B-Tree) suffit amplement car c'est une égalité exacte (=)
CREATE INDEX IF NOT EXISTS invoices_status_idx ON invoices (status);