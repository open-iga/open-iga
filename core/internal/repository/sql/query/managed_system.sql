INSERT INTO managed_system(name, connector_url, connector_sha)
VALUES ($1, $2, #3)
RETURNING *;
