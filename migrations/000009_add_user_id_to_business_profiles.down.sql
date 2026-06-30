-- Supprimer l'index d'abord
DROP INDEX IF EXISTS business_profiles_user_id_idx;

-- Supprimer la colonne ensuite
ALTER TABLE business_profiles 
DROP COLUMN IF EXISTS user_id;