# Integration Test Report — miniMule Backend

**Date:** 2026-04-26  
**Branch:** dev  
**Commit:** 7b8eadc  
**Go version:** 1.23+  
**Test command:** `go test ./tests/... -v -timeout 60s`  
**Result:** ✅ 34 / 34 PASS — 0 FAIL  
**Total time:** 4.619 s

---

## Summary

| Category          | Tests | Passed | Failed |
|-------------------|------:|-------:|-------:|
| Health / Infra    |     1 |      1 |      0 |
| Auth              |     4 |      4 |      0 |
| Products          |     5 |      5 |      0 |
| Categories        |     1 |      1 |      0 |
| Users / Profile   |     3 |      3 |      0 |
| Cart              |     2 |      2 |      0 |
| Orders            |     2 |      2 |      0 |
| Social / Feed     |     6 |      6 |      0 |
| Search / Promos   |     2 |      2 |      0 |
| Notifications     |     2 |      2 |      0 |
| Rate Limiting     |     1 |      1 |      0 |
| Shipping          |     1 |      1 |      0 |
| Posts             |     2 |      2 |      0 |
| **Total**         |**34** |  **34**|   **0**|

---

## Test Results

| # | Test Name                      | Status | Duration |
|---|-------------------------------|--------|----------|
|  1 | TestHealth                    | PASS   | 0.01 s   |
|  2 | TestLogin_ValidCredentials    | PASS   | 0.21 s   |
|  3 | TestLogin_InvalidPassword     | PASS   | 0.20 s   |
|  4 | TestLogin_UnknownEmail        | PASS   | 0.00 s   |
|  5 | TestRegister_NewUser          | PASS   | 0.20 s   |
|  6 | TestRegister_DuplicateEmail   | PASS   | 0.20 s   |
|  7 | TestProducts_List             | PASS   | 0.01 s   |
|  8 | TestProducts_WithImages       | PASS   | 0.02 s   |
|  9 | TestProducts_FilterByFeatured | PASS   | 0.00 s   |
| 10 | TestProduct_Single            | PASS   | 0.00 s   |
| 11 | TestProduct_NotFound          | PASS   | 0.00 s   |
| 12 | TestCategories                | PASS   | 0.00 s   |
| 13 | TestMe_Unauthenticated        | PASS   | 0.00 s   |
| 14 | TestMe_Authenticated          | PASS   | 0.20 s   |
| 15 | TestMe_ArtistProfile          | PASS   | 0.20 s   |
| 16 | TestMyCart_Authenticated      | PASS   | 0.21 s   |
| 17 | TestAddToCart                 | PASS   | 0.21 s   |
| 18 | TestShippingMethods           | PASS   | 0.00 s   |
| 19 | TestFeed                      | PASS   | 0.01 s   |
| 20 | TestFeed_Authenticated_LikedByMe | PASS | 0.20 s  |
| 21 | TestPostComments              | PASS   | 0.00 s   |
| 22 | TestStickerDesigns            | PASS   | 0.01 s   |
| 23 | TestMyOrders_Authenticated    | PASS   | 0.20 s   |
| 24 | TestOrder_Ownership           | PASS   | 0.21 s   |
| 25 | TestSearchProducts            | PASS   | 0.00 s   |
| 26 | TestCheckPromoCode_Invalid    | PASS   | 0.00 s   |
| 27 | TestMyNotifications_Authenticated | PASS | 0.21 s |
| 28 | TestMySearchHistory           | PASS   | 0.21 s   |
| 29 | TestShowcase                  | PASS   | 0.00 s   |
| 30 | TestLikeUnlikePost            | PASS   | 0.21 s   |
| 31 | TestVoteUnvotePost            | PASS   | 0.21 s   |
| 32 | TestAddComment                | PASS   | 0.20 s   |
| 33 | TestCreatePost                | PASS   | 0.20 s   |
| 34 | TestRateLimit                 | PASS   | 0.00 s   |

---

## Notable Fixes Made During Testing

### 1. UpsertCartItem — pgx extended protocol incompatibility
**Symptom:** `invalid input syntax for type uuid: ""` on `addToCart` mutation.  
**Root cause:** pgx v5's extended query protocol cannot infer types for functional `ON CONFLICT` expressions (`COALESCE(variant_id::TEXT, '')`). The parameter type hint comes from the conflict target, but the function call breaks type inference, causing the `variant_id` parameter to be sent as an untyped empty string.  
**Fix:** Replaced the single `INSERT ... ON CONFLICT ... DO UPDATE` with an explicit `SELECT` for an existing row, then a conditional `UPDATE` or `INSERT`. This avoids the ON CONFLICT clause entirely.

### 2. Schema resolver slice nullability
**Symptom:** Server panic at startup — `[]*T is not a pointer` (or `*[]*T is not a slice`).  
**Root cause:** `graph-gophers/graphql-go` requires:
- Non-null list `[T!]!` → Go return type `[]*Resolver`
- Nullable list `[T!]` → Go return type `*[]*Resolver`  
**Fix:** Audited all resolver methods against schema nullability and corrected return types.

### 3. Stale server binary
**Symptom:** `TestAddToCart` still failing after code fix.  
**Root cause:** The running `server.exe` was compiled before the UpsertCartItem rewrite.  
**Fix:** Kill old process, `go build`, restart.

---

## Deployment Note

This backend uses `pgxpool` (persistent connection pool) and Redis — **Vercel serverless is incompatible**. Recommended deployment targets: **Railway** or **Fly.io**.

See [railway-deploy.md](railway-deploy.md) or the `k8s/` directory for Kubernetes manifests.
