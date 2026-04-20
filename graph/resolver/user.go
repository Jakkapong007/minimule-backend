package resolver

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"

	"github.com/jakka/minimule-backend/graph/model"
	"github.com/jakka/minimule-backend/internal/service"
)

// ── UserResolver ──────────────────────────────────────────────────────────────

type UserResolver struct {
	u    *model.User
	root *RootResolver
}

func (r *UserResolver) ID() graphql.ID   { return graphql.ID(r.u.ID) }
func (r *UserResolver) Email() string    { return r.u.Email }
func (r *UserResolver) Phone() *string   { return r.u.Phone }
func (r *UserResolver) FullName() *string { return r.u.FullName }
func (r *UserResolver) AvatarUrl() *string { return r.u.AvatarURL }
func (r *UserResolver) Role() string     { return string(r.u.Role) }
func (r *UserResolver) IsActive() bool   { return r.u.IsActive }
func (r *UserResolver) CreatedAt() graphql.Time {
	return graphql.Time{Time: r.u.CreatedAt}
}
func (r *UserResolver) UpdatedAt() *graphql.Time {
	if r.u.UpdatedAt == nil {
		return nil
	}
	t := graphql.Time{Time: *r.u.UpdatedAt}
	return &t
}

func (r *UserResolver) Profile(ctx context.Context) (*UserProfileResolver, error) {
	if r.u.Profile != nil {
		return &UserProfileResolver{p: r.u.Profile}, nil
	}
	p, err := r.root.Auth.GetUserProfile(ctx, r.u.ID)
	if errors.Is(err, service.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &UserProfileResolver{p: p}, nil
}

func (r *UserResolver) PaymentMethods(ctx context.Context) (*[]*PaymentMethodResolver, error) {
	methods, err := r.root.PaymentSvc.GetMethods(ctx, r.u.ID)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*PaymentMethodResolver, len(methods))
	for i, m := range methods {
		resolvers[i] = &PaymentMethodResolver{m: m}
	}
	return &resolvers, nil
}

func (r *UserResolver) Addresses(ctx context.Context) (*[]*UserAddressResolver, error) {
	if r.u.Addresses != nil {
		resolvers := make([]*UserAddressResolver, len(r.u.Addresses))
		for i, a := range r.u.Addresses {
			resolvers[i] = &UserAddressResolver{a: a}
		}
		return &resolvers, nil
	}
	addrs, err := r.root.Auth.GetUserAddresses(ctx, r.u.ID)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*UserAddressResolver, len(addrs))
	for i, a := range addrs {
		resolvers[i] = &UserAddressResolver{a: a}
	}
	return &resolvers, nil
}

// ── UserProfileResolver ───────────────────────────────────────────────────────

type UserProfileResolver struct {
	p *model.UserProfile
}

func (r *UserProfileResolver) ID() graphql.ID            { return graphql.ID(r.p.ID) }
func (r *UserProfileResolver) Bio() *string              { return r.p.Bio }
func (r *UserProfileResolver) PreferredLanguage() string { return r.p.PreferredLanguage }
func (r *UserProfileResolver) PDPAConsent() bool         { return r.p.PDPAConsent }
func (r *UserProfileResolver) PDPAVersion() *string      { return r.p.PDPAVersion }

// ── UserAddressResolver ───────────────────────────────────────────────────────

type UserAddressResolver struct {
	a *model.UserAddress
}

func (r *UserAddressResolver) ID() graphql.ID           { return graphql.ID(r.a.ID) }
func (r *UserAddressResolver) Label() *string           { return r.a.Label }
func (r *UserAddressResolver) RecipientName() string    { return r.a.RecipientName }
func (r *UserAddressResolver) Phone() string            { return r.a.Phone }
func (r *UserAddressResolver) AddressLine1() string     { return r.a.AddressLine1 }
func (r *UserAddressResolver) AddressLine2() *string    { return r.a.AddressLine2 }
func (r *UserAddressResolver) Subdistrict() *string     { return r.a.Subdistrict }
func (r *UserAddressResolver) District() *string        { return r.a.District }
func (r *UserAddressResolver) Province() *string        { return r.a.Province }
func (r *UserAddressResolver) PostalCode() *string      { return r.a.PostalCode }
func (r *UserAddressResolver) IsDefault() bool          { return r.a.IsDefault }
