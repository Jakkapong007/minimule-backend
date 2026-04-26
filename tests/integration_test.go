// Package tests contains integration tests that run against a live server.
// Prerequisites: server running on localhost:8080, seed data loaded.
// Run: go test ./tests/... -v -timeout 60s
package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

const baseURL = "http://localhost:8080"

// ── helpers ───────────────────────────────────────────────────────────────────

type gqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type gqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func gql(t *testing.T, query string, vars map[string]any, token string) gqlResponse {
	t.Helper()
	body, _ := json.Marshal(gqlRequest{Query: query, Variables: vars})
	req, _ := http.NewRequest(http.MethodPost, baseURL+"/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("HTTP error: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out gqlResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decode error: %v\nBody: %s", err, raw)
	}
	return out
}

func noErrors(t *testing.T, r gqlResponse) {
	t.Helper()
	if len(r.Errors) > 0 {
		msgs := make([]string, len(r.Errors))
		for i, e := range r.Errors {
			msgs[i] = e.Message
		}
		t.Errorf("unexpected GraphQL errors: %s", strings.Join(msgs, "; "))
	}
}

func dataField(t *testing.T, r gqlResponse, path ...string) any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(r.Data, &m); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	var cur any = m
	for _, key := range path {
		if mm, ok := cur.(map[string]any); ok {
			cur = mm[key]
		} else {
			t.Fatalf("path %v: expected object at %q, got %T", path, key, cur)
		}
	}
	return cur
}

func mustLogin(t *testing.T, email, password string) string {
	t.Helper()
	r := gql(t, `mutation Login($e:String!,$p:String!){login(email:$e,password:$p)}`,
		map[string]any{"e": email, "p": password}, "")
	noErrors(t, r)
	token, _ := dataField(t, r, "login").(string)
	if token == "" {
		t.Fatal("login returned empty token")
	}
	return token
}

// ── test suite ────────────────────────────────────────────────────────────────

func TestHealth(t *testing.T) {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestLogin_ValidCredentials(t *testing.T) {
	token := mustLogin(t, "customer@minimule.com", "password123")
	if len(token) < 20 {
		t.Errorf("token looks too short: %q", token)
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	r := gql(t, `mutation{login(email:"customer@minimule.com",password:"wrongpass")}`, nil, "")
	if len(r.Errors) == 0 {
		t.Error("expected error for wrong password, got none")
	}
}

func TestLogin_UnknownEmail(t *testing.T) {
	r := gql(t, `mutation{login(email:"nobody@example.com",password:"pass")}`, nil, "")
	if len(r.Errors) == 0 {
		t.Error("expected error for unknown email")
	}
}

func TestRegister_NewUser(t *testing.T) {
	email := fmt.Sprintf("test_%d@example.com", time.Now().UnixNano())
	r := gql(t, `mutation Reg($e:String!,$p:String!,$n:String!){register(email:$e,password:$p,name:$n){id email}}`,
		map[string]any{"e": email, "p": "password123", "n": "Test User"}, "")
	noErrors(t, r)
	id, _ := dataField(t, r, "register", "id").(string)
	if id == "" {
		t.Error("register returned no id")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	r := gql(t, `mutation{register(email:"customer@minimule.com",password:"password123",name:"Dup"){id}}`, nil, "")
	if len(r.Errors) == 0 {
		t.Error("expected conflict error for duplicate email")
	}
}

func TestProducts_List(t *testing.T) {
	r := gql(t, `{products{id name basePrice status category{name}}}`, nil, "")
	noErrors(t, r)
	list, _ := dataField(t, r, "products").([]any)
	if len(list) == 0 {
		t.Error("expected at least one product")
	}
	// Only active products should appear
	for _, item := range list {
		p := item.(map[string]any)
		if p["status"] != "active" {
			t.Errorf("non-active product in list: %v", p["status"])
		}
	}
}

func TestProducts_WithImages(t *testing.T) {
	r := gql(t, `{products{id name images{imageUrl isPrimary}variants{sku priceModifier}}}`, nil, "")
	noErrors(t, r)
	list, _ := dataField(t, r, "products").([]any)
	for _, item := range list {
		p := item.(map[string]any)
		_ = p["images"]  // nullable — just ensure no error
	}
}

func TestProducts_FilterByFeatured(t *testing.T) {
	r := gql(t, `query($f:Boolean){products(isFeatured:$f){id name isFeatured}}`,
		map[string]any{"f": true}, "")
	noErrors(t, r)
	list, _ := dataField(t, r, "products").([]any)
	if len(list) == 0 {
		t.Error("expected at least one featured product")
	}
	for _, item := range list {
		p := item.(map[string]any)
		if p["isFeatured"] != true {
			t.Errorf("product isFeatured should be true, got %v", p["isFeatured"])
		}
	}
}

func TestProduct_Single(t *testing.T) {
	r := gql(t, `{product(id:"00000000-0000-0000-0004-000000000001"){id name basePrice avgRating reviewCount}}`, nil, "")
	noErrors(t, r)
	id, _ := dataField(t, r, "product", "id").(string)
	if id == "" {
		t.Error("expected product id")
	}
}

func TestProduct_NotFound(t *testing.T) {
	r := gql(t, `{product(id:"00000000-0000-0000-0000-000000000099"){id}}`, nil, "")
	noErrors(t, r)
	prod := dataField(t, r, "product")
	if prod != nil {
		t.Errorf("expected nil for nonexistent product, got %v", prod)
	}
}

func TestCategories(t *testing.T) {
	r := gql(t, `{categories{id name slug isActive}}`, nil, "")
	noErrors(t, r)
	list, _ := dataField(t, r, "categories").([]any)
	if len(list) < 5 {
		t.Errorf("expected at least 5 categories, got %d", len(list))
	}
}

func TestMe_Unauthenticated(t *testing.T) {
	r := gql(t, `{me{id email}}`, nil, "")
	noErrors(t, r)
	me := dataField(t, r, "me")
	if me != nil {
		t.Error("expected nil for unauthenticated me")
	}
}

func TestMe_Authenticated(t *testing.T) {
	token := mustLogin(t, "customer@minimule.com", "password123")
	r := gql(t, `{me{id email role profile{bio preferredLanguage}addresses{label isDefault}}}`, nil, token)
	noErrors(t, r)
	email, _ := dataField(t, r, "me", "email").(string)
	if email != "customer@minimule.com" {
		t.Errorf("expected customer email, got %q", email)
	}
}

func TestMe_ArtistProfile(t *testing.T) {
	token := mustLogin(t, "artist@minimule.com", "password123")
	r := gql(t, `{me{id email role fullName paymentMethods{id type isDefault}}}`, nil, token)
	noErrors(t, r)
	role, _ := dataField(t, r, "me", "role").(string)
	if role != "artist" {
		t.Errorf("expected role=artist, got %q", role)
	}
}

func TestMyCart_Authenticated(t *testing.T) {
	token := mustLogin(t, "customer@minimule.com", "password123")
	r := gql(t, `{myCart{id status subtotal items{id quantity unitPrice product{name}}}}`, nil, token)
	noErrors(t, r)
	status, _ := dataField(t, r, "myCart", "status").(string)
	if status != "active" {
		t.Errorf("expected active cart, got %q", status)
	}
}

func TestAddToCart(t *testing.T) {
	token := mustLogin(t, "customer@minimule.com", "password123")
	r := gql(t, `mutation AddToCart($pid:ID!,$qty:Int!){addToCart(productId:$pid,quantity:$qty){id status subtotal items{quantity}}}`,
		map[string]any{"pid": "00000000-0000-0000-0004-000000000005", "qty": 1}, token)
	noErrors(t, r)
	status, _ := dataField(t, r, "addToCart", "status").(string)
	if status != "active" {
		t.Errorf("expected active cart after addToCart, got %q", status)
	}
}

func TestShippingMethods(t *testing.T) {
	r := gql(t, `{shippingMethods{id name carrier baseFee estimatedDaysMin estimatedDaysMax}}`, nil, "")
	noErrors(t, r)
	// May be empty if no seed data for shipping methods — just check no error
}

func TestFeed(t *testing.T) {
	r := gql(t, `{feed{id caption imageUrl likeCount commentCount voteCount visibility user{fullName}}}`, nil, "")
	noErrors(t, r)
	list, _ := dataField(t, r, "feed").([]any)
	if len(list) == 0 {
		t.Error("expected at least one post in feed")
	}
}

func TestFeed_Authenticated_LikedByMe(t *testing.T) {
	token := mustLogin(t, "customer@minimule.com", "password123")
	r := gql(t, `{feed{id isLikedByMe isVotedByMe}}`, nil, token)
	noErrors(t, r)
}

func TestPostComments(t *testing.T) {
	r := gql(t, `{post(id:"00000000-0000-0000-0011-000000000001"){id likeCount comments{id body user{fullName}}}}`, nil, "")
	noErrors(t, r)
	comments, _ := dataField(t, r, "post", "comments").([]any)
	if len(comments) == 0 {
		t.Error("expected comments on post 1")
	}
}

func TestStickerDesigns(t *testing.T) {
	r := gql(t, `{stickerDesigns{id isStickerDesign}}`, nil, "")
	noErrors(t, r)
	list, _ := dataField(t, r, "stickerDesigns").([]any)
	for _, item := range list {
		p := item.(map[string]any)
		if p["isStickerDesign"] != true {
			t.Errorf("stickerDesigns returned non-sticker post")
		}
	}
}

func TestMyOrders_Authenticated(t *testing.T) {
	token := mustLogin(t, "customer@minimule.com", "password123")
	r := gql(t, `{myOrders{id orderNumber status subtotal discountAmount shippingFee total items{quantity unitPrice product{name}}}}`, nil, token)
	noErrors(t, r)
	list, _ := dataField(t, r, "myOrders").([]any)
	if len(list) < 2 {
		t.Errorf("expected at least 2 seeded orders, got %d", len(list))
	}
}

func TestOrder_Ownership(t *testing.T) {
	// customer can see their own order
	token := mustLogin(t, "customer@minimule.com", "password123")
	r := gql(t, `{order(id:"00000000-0000-0000-0009-000000000001"){id orderNumber status}}`, nil, token)
	noErrors(t, r)
	orderNumber, _ := dataField(t, r, "order", "orderNumber").(string)
	if !strings.HasPrefix(orderNumber, "MML-") {
		t.Errorf("unexpected orderNumber: %q", orderNumber)
	}
}

func TestSearchProducts(t *testing.T) {
	r := gql(t, `{searchProducts(query:"cat"){products{id name}total}}`, nil, "")
	noErrors(t, r)
	total, _ := dataField(t, r, "searchProducts", "total").(float64)
	if int(total) == 0 {
		t.Error("search for 'cat' should return at least one result")
	}
}

func TestCheckPromoCode_Invalid(t *testing.T) {
	r := gql(t, `{checkPromoCode(code:"FAKECODE",orderTotal:500){valid message discountAmount}}`, nil, "")
	noErrors(t, r)
	valid, _ := dataField(t, r, "checkPromoCode", "valid").(bool)
	if valid {
		t.Error("expected invalid promo code result")
	}
}

func TestMyNotifications_Authenticated(t *testing.T) {
	token := mustLogin(t, "customer@minimule.com", "password123")
	r := gql(t, `{myNotifications{unreadCount items{id title isRead}}}`, nil, token)
	noErrors(t, r)
}

func TestMySearchHistory(t *testing.T) {
	token := mustLogin(t, "customer@minimule.com", "password123")
	// Run a search to generate history
	gql(t, `{searchProducts(query:"sticker"){total}}`, nil, token)
	r := gql(t, `{mySearchHistory}`, nil, token)
	noErrors(t, r)
}

func TestShowcase(t *testing.T) {
	r := gql(t, `{showcase{id imageUrl visibility}}`, nil, "")
	noErrors(t, r)
}

func TestLikeUnlikePost(t *testing.T) {
	token := mustLogin(t, "artist@minimule.com", "password123")
	// Like
	r := gql(t, `mutation{likePost(postId:"00000000-0000-0000-0011-000000000003"){id likeCount}}`, nil, token)
	noErrors(t, r)
	// Unlike
	r = gql(t, `mutation{unlikePost(postId:"00000000-0000-0000-0011-000000000003"){id likeCount}}`, nil, token)
	noErrors(t, r)
}

func TestVoteUnvotePost(t *testing.T) {
	token := mustLogin(t, "artist@minimule.com", "password123")
	r := gql(t, `mutation{votePost(postId:"00000000-0000-0000-0011-000000000002"){id voteCount}}`, nil, token)
	noErrors(t, r)
	r = gql(t, `mutation{unvotePost(postId:"00000000-0000-0000-0011-000000000002"){id voteCount}}`, nil, token)
	noErrors(t, r)
}

func TestAddComment(t *testing.T) {
	token := mustLogin(t, "customer@minimule.com", "password123")
	r := gql(t, `mutation AddComment($pid:ID!,$b:String!){addComment(postId:$pid,body:$b){id body}}`,
		map[string]any{"pid": "00000000-0000-0000-0011-000000000004", "b": "Integration test comment"}, token)
	noErrors(t, r)
	body, _ := dataField(t, r, "addComment", "body").(string)
	if body != "Integration test comment" {
		t.Errorf("unexpected comment body: %q", body)
	}
}

func TestCreatePost(t *testing.T) {
	token := mustLogin(t, "artist@minimule.com", "password123")
	r := gql(t, `mutation CreatePost($in:CreatePostInput!){createPost(input:$in){id imageUrl visibility}}`,
		map[string]any{"in": map[string]any{
			"imageUrl":        "https://picsum.photos/seed/inttest/600/600",
			"caption":         "Integration test post",
			"isStickerDesign": false,
			"visibility":      "public",
		}}, token)
	noErrors(t, r)
	id, _ := dataField(t, r, "createPost", "id").(string)
	if id == "" {
		t.Error("createPost returned no id")
	}
	// Cleanup — delete the post
	gql(t, fmt.Sprintf(`mutation{deletePost(id:"%s")}`, id), nil, token)
}

func TestRateLimit(t *testing.T) {
	// Send many rapid unauthenticated requests — should not crash
	for i := 0; i < 5; i++ {
		r := gql(t, `{categories{id}}`, nil, "")
		_ = r
	}
}
