-- +goose Up
CREATE TABLE IF NOT EXISTS orderitem (
    order_id INT NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    menu_item_id INT NOT NULL REFERENCES menuitem (id) ON DELETE RESTRICT,
    quanity INT NOT NULL,
    PRIMARY KEY (order_id, menu_item_id)
);

-- +goose Down
DROP TABLE IF EXISTS orderitem;