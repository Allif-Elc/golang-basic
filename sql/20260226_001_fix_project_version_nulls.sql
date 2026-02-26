-- Fix project version NULL values
-- Updates any existing NULL or empty version strings to default '1.0'

UPDATE projects
SET version = '1.0'
WHERE version IS NULL OR version = '';
