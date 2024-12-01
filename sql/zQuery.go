package sql

var ViewProductsQuery = `
		SELECT 
			p.id, 
			p.name, 
			p.description, 
			p.price, 
			c.name AS category_name,
			CASE 
				WHEN p.stock > 0 THEN 'available'
				ELSE 'out of stock'
			END AS status
		FROM products p
		JOIN categories c ON p.category_id = c.id;
	`
var AdminViewProductsQuery = `
		SELECT 
			p.id, 
			p.name, 
			p.description, 
			p.price, 
			p.category_id,
			p.stock,
			p.deleted
		FROM products p;
	`
