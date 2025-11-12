
UPDATE products SET category_id = (SELECT id FROM categories WHERE name = 'Clothing')
WHERE products.code in ('PROD001', 'PROD004', 'PROD007');

UPDATE products SET category_id = (SELECT id FROM categories WHERE name = 'Shoes')
WHERE products.code in ('PROD002', 'PROD006');

UPDATE products SET category_id = (SELECT id FROM categories WHERE name = 'Accessories')
WHERE products.code in ('PROD003', 'PROD005', 'PROD008');
