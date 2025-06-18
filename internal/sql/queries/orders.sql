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
    orders.status,
    orders.created_at,
    orders.address AS delivery_address,

    orderitem.menu_item_id,
    orderitem.quanity,

    menuitem.name AS menu_item_name,
    menuitem.price,

    restaurants.address AS restaurant_address,
    restaurants.user_name AS restaurant_name,
    restaurants.phone AS restaurant_phone,

    customer.user_name AS costomer_name,
    customer.phone AS customer_phone

FROM orders
         JOIN orderitem ON orders.id = orderitem.order_id
         JOIN menuitem ON orderitem.menu_item_id = menuitem.id
         JOIN users AS restaurants ON orders.restaurantid = restaurants.id
         JOIN users AS customer ON orders.customerid = customer.id
WHERE orders.customerid = $1;

-- name: GetFullOrderByID :many
SELECT
    orders.id AS order_id,
    orders.status,
    orders.created_at,
    orders.address AS delivery_address,

    orderitem.menu_item_id,
    orderitem.quanity,

    menuitem.name AS menu_item_name,
    menuitem.price,

    restaurants.address AS restaurant_address,
    restaurants.user_name AS restaurant_name,
    restaurants.phone AS restaurant_phone,

    customer.user_name AS costomer_name,
    customer.phone AS customer_phone,
    customer.id AS customer_id,

    courier.user_name AS courier_name

FROM orders
         JOIN orderitem ON orders.id = orderitem.order_id
         JOIN menuitem ON orderitem.menu_item_id = menuitem.id
         JOIN users AS restaurants ON orders.restaurantid = restaurants.id
         JOIN users AS customer ON orders.customerid = customer.id
        LEFT JOIN users AS courier ON orders.courierid = courier.id
WHERE orders.id = $1;


-- name: GetFullPendingOrders :many
SELECT
    orders.id AS order_id,
    orders.status,
    orders.created_at,
    orders.address AS delivery_address,

    orderitem.menu_item_id,
    orderitem.quanity,

    menuitem.name AS menu_item_name,
    menuitem.price,

    restaurants.address AS restaurant_address,
    restaurants.user_name AS restaurant_name,
    restaurants.phone AS restaurant_phone,

    customer.user_name AS costomer_name,
    customer.phone AS customer_phone

FROM orders
         JOIN orderitem ON orders.id = orderitem.order_id
         JOIN menuitem ON orderitem.menu_item_id = menuitem.id
         JOIN users AS restaurants ON orders.restaurantid = restaurants.id
        JOIN users AS customer ON orders.customerid = customer.id
WHERE orders.status = 'pending';



-- name: GetOrderStatusByID :one
SELECT orders.status FROM orders
WHERE orders.id = $1;

-- name: GetCurrentIDOrderForCourier :one
SELECT id FROM orders
WHERE status = 'delivering' AND courierid=$1;


-- name: UpdateCourierID :many
WITH updated_order AS (
    UPDATE orders
    SET courierid = $1,
        status = 'delivering'
    WHERE orders.id = $2
    RETURNING *
)
SELECT
    o.id AS order_id,
    o.status,
    o.created_at,
    o.address AS delivery_address,

    orderitem.menu_item_id,
    orderitem.quanity,

    menuitem.name AS menu_item_name,
    menuitem.price,

    restaurants.address AS restaurant_address,
    restaurants.user_name AS restaurant_name,
    restaurants.phone AS restaurant_phone,

    customer.user_name AS costomer_name,
    customer.phone AS customer_phone,
    customer.id AS customer_id,

    courier.user_name AS courier_name

FROM updated_order o
         JOIN orderitem ON o.id = orderitem.order_id
         JOIN menuitem ON orderitem.menu_item_id = menuitem.id
         JOIN users AS restaurants ON o.restaurantid = restaurants.id
         JOIN users AS customer ON o.customerid = customer.id
         LEFT JOIN users AS courier ON o.courierid = courier.id;


-- name: GetCurrentOrderForCourier :many
SELECT
    orders.id AS order_id,
    orders.status,
    orders.created_at,
    orders.address AS delivery_address,

    orderitem.menu_item_id,
    orderitem.quanity,

    menuitem.name AS menu_item_name,
    menuitem.price,

    restaurants.address AS restaurant_address,
    restaurants.user_name AS restaurant_name,
    restaurants.phone AS restaurant_phone,

    customer.phone AS customer_phone

FROM orders
         JOIN orderitem ON orders.id = orderitem.order_id
         JOIN menuitem ON orderitem.menu_item_id = menuitem.id
         JOIN users AS restaurants ON orders.restaurantid = restaurants.id
         JOIN users AS customer ON orders.customerid = customer.id
WHERE orders.status = 'delivering' AND orders.courierid = $1;


-- name: UpdateOrderStatus :exec
UPDATE orders
SET courierid = $1,
    status = 'delivered'
WHERE orders.id = $2;