package checkout

const StmtInsertCartItem = `
INSERT INTO b2b.cart_item
SET id = :id,
	order_id = :order_id,
	price = :price,
	new_copy_count = :new_copy_count,
	renewal_list = :renewal_list
`