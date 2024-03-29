package reader

const colMembership = `
SELECT vip_id AS compound_id,
	NULLIF(vip_id, vip_id_alias) AS ftc_id,
	vip_id_alias AS union_id,
	vip_type,
	expire_time,
	member_tier AS tier,
	billing_cycle AS cycle,
	expire_date,
	payment_method,
	ftc_plan_id,
	stripe_subscription_id AS stripe_subs_id,
	stripe_plan_id,
	IFNULL(auto_renewal, FALSE) AS auto_renewal,
	sub_status AS subs_status,
	apple_subscription_id AS apple_subs_id,
	b2b_licence_id,
	standard_addon,
	premium_addon
FROM premium.ftc_vip`

const StmtSelectMember = colMembership + `
WHERE ? IN (vip_id, vip_id_alias)
LIMIT 1
`

// StmtLockMember builds SQL to retrieve membership in a transaction.
// Retrieve membership by compound id extracted from request header.
// The request might provide ftc id or union id, or both,
// and we cannot be sure the current state account ids
// are consistent with the those in db.
// There are chances that the request provides union id
// while vip_id is ftc id and vip_id_alias is union id.
// (Chances of such case are rare).
// In such case we won't be able to find the membership
// simply querying the vip_id column.
const StmtLockMember = StmtSelectMember + `
FOR UPDATE`

const mUpsertCols = `
vip_type = :vip_type,
expire_time = :expire_time,
member_tier = :tier,
billing_cycle = :cycle,
expire_date = :expire_date,
payment_method = :payment_method,
ftc_plan_id = :ftc_plan_id,
stripe_subscription_id = :stripe_subs_id,
stripe_plan_id = :stripe_plan_id,
auto_renewal = :auto_renewal,
sub_status = :subs_status,
apple_subscription_id = :apple_subs_id,
b2b_licence_id = :b2b_licence_id,
standard_addon = :standard_addon,
premium_addon = :premium_addon
`

const StmtCreateMember = `
INSERT INTO premium.ftc_vip
SET vip_id = :compound_id,
	vip_id_alias = :union_id,
	ftc_user_id = :ftc_id,
	wx_union_id = :union_id,
` + mUpsertCols

const StmtUpdateMember = `
UPDATE premium.ftc_vip
SET ` + mUpsertCols + `
WHERE vip_id = :compound_id
LIMIT 1`
