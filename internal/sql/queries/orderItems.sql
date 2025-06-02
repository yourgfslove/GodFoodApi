-- name: AddItem :exec
INSERT INTO orderitem(order_id, menu_item_id, quanity)
VALUES (
        $1,
        $2,
        $3
       );


-- name: AddItems :many
INSERT INTO orderitem(order_id, menu_item_id, quanity)
SELECT unnest($1::int[]), unnest($2::int[]), unnest($3::int[])
RETURNING *;