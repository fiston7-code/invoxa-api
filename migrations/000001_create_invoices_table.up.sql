CREATE TABLE IF NOT EXISTS invoices (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    invoice_number citext NOT NULL,
    invoice_date timestamp(0) with time zone NOT NULL,
    
    business_name text NOT NULL,
    business_logo_url text NOT NULL,
    business_rccm text,
    
    client_name text NOT NULL,
    client_phone text,
    client_email citext,
    client_address text,
    
    -- Le montant total stocké en centimes (bigint) pour éviter les erreurs de float en Go
    total_amount bigint NOT NULL, 
    currency text NOT NULL DEFAULT 'USD',
    
    note_title text,
    note_text text,
    footer_address text,
    footer_phone text,
    footer_email text,
    
    status text NOT NULL DEFAULT 'draft',
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    version integer NOT NULL DEFAULT 1
);


CREATE TABLE IF NOT EXISTS invoice_items (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  
    invoice_id bigint NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    description text NOT NULL,
    quantity integer NOT NULL,
    unit_price bigint NOT NULL, 
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);