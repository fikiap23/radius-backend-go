// Package coordination holds cross-bounded-context use cases that need a single
// database transaction in the monolith (one PostgreSQL, one ent.Client).
//
// Use deps.RunInTransaction from module.Dependencies — implemented via
// database.RunInTx. Build repositories with the transactional client:
//
//	return deps.RunInTransaction(ctx, func(ctx context.Context, tx *ent.Client) error {
//	    userRepo := userpostgres.NewUserRepository(tx)
//	    // otherRepo := otherpostgres.New...(tx)
//	    // ...
//	    return nil
//	})
//
// Performance: keep the callback short (DB only). Do not call HTTP, OAuth, or
// other slow I/O inside RunInTransaction. Do not nest RunInTransaction calls.
//
// Intra-context multi-write (e.g. users SSO) may call RunInTransaction from
// that context's application service instead of living here.
package coordination
