-- 1. Ajouter la colonne user_id à la table business_profiles
ALTER TABLE business_profiles 
ADD COLUMN user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE;

-- 2. Ajouter un index pour accélérer les requêtes de recherche par utilisateur
CREATE INDEX IF NOT EXISTS business_profiles_user_id_idx ON business_profiles(user_id);