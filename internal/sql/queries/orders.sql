-- name: CreateOrder :one
INSERT INTO orders(customerid, restaurantid, address, status, created_at)
VALUES (
        $1,
        $2,
        $3,
        'pending',
        NOW()
)
RETURNING *;

-- name: GetFullOrdersByUserID :many
SELECT
    orders.id AS order_id,
    orders.customerid,
    orders.restaurantid AS order_restaurant_id,
    orders.courierid,
    orders.status,
    orders.created_at,
    orders.address,

    orderitem.menu_item_id,
    orderitem.quanity,

    menuitem.name AS menu_item_name,
    menuitem.price

FROM orders
         JOIN orderitem ON orders.id = orderitem.order_id
         JOIN menuitem ON orderitem.menu_item_id = menuitem.id
WHERE orders.customerid = $1;

-- name: GetFullOrderByID :many
SELECT
    orders.id AS order_id,
    orders.customerid,
    orders.restaurantid AS order_restaurant_id,
    orders.courierid,
    orders.status,
    orders.created_at,
    orders.address,

    orderitem.menu_item_id,
    orderitem.quanity,

    menuitem.name AS menu_item_name,
    menuitem.price

FROM orders
         JOIN orderitem ON orders.id = orderitem.order_id
         JOIN menuitem ON orderitem.menu_item_id = menuitem.id
WHERE orders.id = $1;
