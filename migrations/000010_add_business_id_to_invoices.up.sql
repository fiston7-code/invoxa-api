-- 1. Ajout de la colonne obligatoire (liée au profil unique de l'entreprise)
ALTER TABLE invoices ADD COLUMN business_id bigint NOT NULL;

-- 2. Liaison par clé étrangère vers la table business_profiles
ALTER TABLE invoices 
ADD CONSTRAINT invoices_business_id_fkey 
FOREIGN KEY (business_id) REFERENCES business_profiles(id) ON DELETE CASCADE;

-- 3. Suppression de l'ancienne contrainte UNIQUE globale qui bloquait les numéros de facture
ALTER TABLE invoices DROP CONSTRAINT invoices_invoice_number_key;

-- 4. Création de la nouvelle contrainte UNIQUE : le numéro est unique UNIQUEMENT pour une même entreprise
ALTER TABLE invoices ADD CONSTRAINT invoices_business_invoice_number_key UNIQUE (business_id, invoice_number);

-- 5. Index de performance pour charger instantanément le Dashboard
CREATE INDEX invoices_business_id_idx ON invoices(business_id);