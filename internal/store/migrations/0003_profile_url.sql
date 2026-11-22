-- +goose Up
-- profile_url is the address of the congressman's page on the chamber's own
-- website, so the site can link to the source that confirms what it shows.
ALTER TABLE congressmen ADD COLUMN profile_url TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE congressmen DROP COLUMN profile_url;
