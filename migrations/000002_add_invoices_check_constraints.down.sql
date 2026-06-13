ALTER TABLE invoices DROP CONSTRAINT IF EXISTS invoices_total_amount_check;
ALTER TABLE invoices DROP CONSTRAINT IF EXISTS invoices_version_check;

ALTER TABLE invoice_items DROP CONSTRAINT IF EXISTS invoice_items_quantity_check;
ALTER TABLE invoice_items DROP CONSTRAINT IF EXISTS invoice_items_unit_price_check;ALTER DATABASE greenlight OWNER TO greenlight;ALTER DATABASE greenlight OWNER TO greenlight;ALTER DATABASE greenlight OWNER TO greenlight;