CREATE TABLE members (
    id uuid,
    email varchar(128),
    hashed_password varchar(128),
    name varchar(128),
    PRIMARY KEY (user_id)
);

CREATE TABLE orders (
    id uuid,
    member_id uuid,
    amount decimal,
    PRIMARY KEY (id)
);

CREATE TABLE order_items (
    order_id uuid,
    product_id uuid,
    amount int,
    PRIMARY KEY (order_id, item_id)
);

CREATE TABLE product (
    id uuid,
    name varchar(128),
    ean varchar(128),
    price decimal,
    category_id uuid,
    PRIMARY KEY (id)
);

CREATE TABLE categories (
    id uuid,
    name varchar(128),
    parent_id uuid,
    PRIMARY KEY (id)
);

-- Example queries
SELECT
    *
FROM
    orders
WHERE
    orders.id = '...';

-- Query plan for Matriarch
-- 1. parse sql statement FROM clause to detect the table involved in the statement
-- 1.1. get its vindexes
-- 2. parse sql statement WHERE clause to see if it matches any of the columns used in
-- vindexes of the table involved (starting with the primary vindex)
-- 2.1 WHERE column matches the primary vindex column 'id'
-- 3 Matriarch hashes the id and routes the query to the appropriate shard
-- 4 Matriarch proxies the response back to the client

SELECT
    *
FROM
    orders
    INNER JOIN order_items ON order_items.order_id = orders.id
WHERE
    orders.id = '...';

-- Query plan for Matriarch
-- 1. parse sql statement FROM clause to detect tables involved in the statement
-- and the join condition
-- 1.1. for each table get the vindexes
-- 2. parse sql statement WHERE clause to see if it matches any of the
-- vindexes of the tables involved (starting with the primary vindex)
-- 2.1 WHERE column matches the primary vindex column of table ecommerce.orders
-- 2.2 Matriarch hashes the order id and identifies the appropriate shard for ecommerce.orders
-- 2.3 By parsing the JOIN condition, Matriarch understands that the column used in the
-- WHERE clause is also used to match data in table order_items.
-- 3 Matriarch routes the whole query to the appropriate shard
-- 4 Matriarch proxies the response back to the client
