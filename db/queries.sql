-- name: InsertWord :exec
INSERT INTO word (name) VALUES ($1)
ON CONFLICT (name) DO NOTHING;

-- name: InsertWordSymbol :exec
INSERT INTO word_symbol (word, pronunciation_num, position, symbol)
VALUES ($1, $2, $3, $4);
