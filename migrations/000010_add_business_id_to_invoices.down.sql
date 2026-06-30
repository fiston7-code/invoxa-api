-- 1. Supprimer l'index de performance sur business_id
DROP INDEX IF EXISTS invoices_business_id_idx;

-- 2. Supprimer la nouvelle contrainte unique combinée
ALTER TABLE invoices DROP CONSTRAINT IF EXISTS invoices_business_invoice_number_key;

-- 3. Rétablir l'ancienne contrainte UNIQUE globale sur le numéro de facture
ALTER TABLE invoices ADD CONSTRAINT invoices_invoice_number_key UNIQUE (invoice_number);

-- 4. Supprimer la clé étrangère reliant invoices à business_profiles
ALTER TABLE invoices DROP CONSTRAINT IF EXISTS invoices_business_id_fkey;

-- 5. Supprimer définitivement la colonne business_id
ALTER TABLE invoices DROP COLUMN IF EXISTS business_id;