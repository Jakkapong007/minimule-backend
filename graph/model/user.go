package model

import "time"

// ── Enums ─────────────────────────────

type UserRole string

const (
	UserRoleCustomer UserRole = "customer"
	UserRoleArtist   UserRole = "artist"
	UserRoleAdmin    UserRole = "admin"
)

// ── Core User ─────────────────────────

type User struct {
	ID           string
	Email        string
	Phone        *string
	PasswordHash string
	FullName     *string
	AvatarURL    *string
	Role         UserRole
	IsActive     bool

	PDPAConsent   bool
	PDPAConsentAt *time.Time
	PDPAVersion   *string

	CreatedAt time.Time
	UpdatedAt *time.Time

	Profile        *UserProfile
	Addresses      []*UserAddress
	PaymentMethods []*UserPaymentMethod
}

// ── Profile ───────────────────────────

type UserProfile struct {
	ID                string
	UserID            string
	Bio               *string
	DateOfBirth       *time.Time
	Gender            *string
	PreferredLanguage string

	NotificationPush  bool
	NotificationEmail bool
	NotificationSMS   bool

	PDPAConsent   bool
	PDPAConsentAt *time.Time
	PDPAVersion   *string
}

// ── Address ───────────────────────────

type UserAddress struct {
	ID            string
	UserID        string
	Label         *string
	RecipientName string
	Phone         string

	AddressLine1 string
	AddressLine2 *string
	Subdistrict  *string
	District     *string
	Province     *string
	PostalCode   *string

	IsDefault bool
}

// ── Payment ───────────────────────────

type UserPaymentMethod struct {
	ID            string
	UserID        string
	Type          string
	Provider      *string
	Last4Digits   *string
	TokenVaultRef *string
	IsDefault     bool
	ExpiresAt     *time.Time
}
