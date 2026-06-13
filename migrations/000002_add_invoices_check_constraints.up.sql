-- Sécurité sur la table des factures
ALTER TABLE invoices ADD CONSTRAINT invoices_total_amount_check CHECK (total_amount >= 0);
ALTER TABLE invoices ADD CONSTRAINT invoices_version_check CHECK (version >= 1);

-- Sécurité sur les lignes d'articles
ALTER TABLE invoice_items ADD CONSTRAINT invoice_items_quantity_check CHECK (quantity >= 1);
ALTER TABLE invoice_items ADD CONSTRAINT invoice_items_unit_price_check CHECK (unit_price >= 0);