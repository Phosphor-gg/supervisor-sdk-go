package supervisor

// ModerationLabel represents a content moderation category.
type ModerationLabel string

const (
	LabelProfanity    ModerationLabel = "profanity"
	LabelToxicity     ModerationLabel = "toxicity"
	LabelHarassment   ModerationLabel = "harassment"
	LabelHate         ModerationLabel = "hate"
	LabelInsult       ModerationLabel = "insult"
	LabelSexual       ModerationLabel = "sexual"
	LabelSexualUnlawful ModerationLabel = "sexual/unlawful"
	LabelSexualExp    ModerationLabel = "sexual/explicit"
	LabelSensitive    ModerationLabel = "sensitive"
	LabelViolence     ModerationLabel = "violence"
	LabelSelfHarm     ModerationLabel = "self-harm"
	LabelMedical      ModerationLabel = "medical"
	LabelSpam         ModerationLabel = "spam"
	LabelPromotional  ModerationLabel = "promotional"
	LabelScam         ModerationLabel = "scam"
	LabelIllegal      ModerationLabel = "illegal"
)

// ModerationModel represents available AI moderation models.
type ModerationModel string

const (
	ModelAuto     ModerationModel = "auto"
	ModelObserver ModerationModel = "observer"
	ModelSentinel ModerationModel = "sentinel"
	ModelArbiter  ModerationModel = "arbiter"
)

// Tier represents subscription tiers.
type Tier string

const (
	TierFree     Tier = "free"
	TierBasic    Tier = "basic"
	TierStandard Tier = "standard"
	TierPremium  Tier = "premium"
)

// BillingCycle represents billing period options.
type BillingCycle string

const (
	BillingMonthly   BillingCycle = "Monthly"
	BillingQuarterly BillingCycle = "Quarterly"
	BillingAnnual    BillingCycle = "Annual"
	BillingTriennial BillingCycle = "Triennial"
)

// ModerationRequest is the request body for POST /api/moderate.
type ModerationRequest struct {
	Text            *string           `json:"text,omitempty"`
	Image           *string           `json:"image,omitempty"`
	Model           *ModerationModel  `json:"model,omitempty"`
	EnabledLabels   []ModerationLabel `json:"enabled_labels,omitempty"`
	IncludeContext  bool              `json:"include_context,omitempty"`
	IncludeImplicit bool              `json:"include_implicit,omitempty"`
}

// BatchModerationRequest is the request body for POST /api/batch.
type BatchModerationRequest struct {
	Texts           []string          `json:"texts"`
	Images          []string          `json:"images,omitempty"`
	Model           *ModerationModel  `json:"model,omitempty"`
	EnabledLabels   []ModerationLabel `json:"enabled_labels,omitempty"`
	IncludeContext  bool              `json:"include_context,omitempty"`
	IncludeImplicit bool              `json:"include_implicit,omitempty"`
}

// UsernameCheckRequest is the request body for POST /api/username.
type UsernameCheckRequest struct {
	Username string `json:"username"`
}

// ModerationResponse is the result of a moderation request.
//
// Label fields are plain strings (not the ModerationLabel enum) so new or
// aliased labels returned by the server pass through rather than being rejected.
type ModerationResponse struct {
	Flagged        bool     `json:"flagged"`
	Labels         []string `json:"labels"`
	ImplicitLabels []string `json:"implicit_labels,omitempty"`
	ModelVersion   *string  `json:"model_version,omitempty"`
	NeedsContext   *bool    `json:"needs_context,omitempty"`
	ContextLabels  []string `json:"context_labels,omitempty"`
	RewrittenText  *string  `json:"rewritten_text,omitempty"`
}

// UsernameCheckResponse is the result of a username check.
type UsernameCheckResponse struct {
	Flagged bool    `json:"flagged"`
	Score   float64 `json:"score"`
}

// PlatformTokenRequest is the OAuth2 client credentials request.
type PlatformTokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
}

// PlatformTokenResponse is the OAuth2 access token response.
type PlatformTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// ProvisionUserRequest is the request to provision/link a user.
type ProvisionUserRequest struct {
	Email string `json:"email"`
}

// ProvisionUserResponse is the result of provisioning a user.
type ProvisionUserResponse struct {
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	IsNewAccount  bool   `json:"is_new_account"`
	IsNewlyLinked bool   `json:"is_newly_linked"`
}

// PlatformUserInfo represents a platform's view of a linked user.
type PlatformUserInfo struct {
	UserID                string `json:"user_id"`
	Email                 string `json:"email"`
	LinkedAt              string `json:"linked_at"`
	Authorized            bool   `json:"authorized"`
	HasActiveSubscription bool   `json:"has_active_subscription"`
	Tier                  Tier   `json:"tier"`
}

// PlatformModerationRequest moderates content on behalf of a linked user.
type PlatformModerationRequest struct {
	UserEmail       string            `json:"user_email"`
	Text            *string           `json:"text,omitempty"`
	Image           *string           `json:"image,omitempty"`
	Model           *ModerationModel  `json:"model,omitempty"`
	EnabledLabels   []ModerationLabel `json:"enabled_labels,omitempty"`
	IncludeContext  bool              `json:"include_context,omitempty"`
	IncludeImplicit bool              `json:"include_implicit,omitempty"`
}

// PlatformCheckoutRequest creates a checkout session for a platform user.
type PlatformCheckoutRequest struct {
	UserEmail    string       `json:"user_email"`
	Tier         Tier         `json:"tier"`
	BillingCycle BillingCycle `json:"billing_cycle"`
	SuccessURL   string       `json:"success_url"`
	CancelURL    string       `json:"cancel_url"`
}

// PlatformCheckoutResponse contains the checkout URL.
type PlatformCheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
}

// PlatformChangePlanRequest changes the plan of a platform user's active subscription.
type PlatformChangePlanRequest struct {
	UserEmail    string       `json:"user_email"`
	Tier         Tier         `json:"tier"`
	BillingCycle BillingCycle `json:"billing_cycle"`
}

// PlatformChangePlanResponse is the result of a plan change.
type PlatformChangePlanResponse struct {
	SubscriptionID string       `json:"subscription_id"`
	Tier           Tier         `json:"tier"`
	BillingCycle   BillingCycle `json:"billing_cycle"`
}

// PlanPrice is a subscription plan price a platform can sell.
type PlanPrice struct {
	PriceID      string       `json:"price_id"`
	ProductID    string       `json:"product_id"`
	Tier         Tier         `json:"tier"`
	BillingCycle BillingCycle `json:"billing_cycle"`
	// Amount is the price in cents.
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	// PaymentLink is always null on the platform products endpoint: mint
	// links via CreateCheckout so the revenue share applies.
	PaymentLink *string `json:"payment_link"`
}

// CreditPack is a one-time credit pack a platform can sell.
type CreditPack struct {
	ID            string  `json:"id"`
	PriceID       string  `json:"price_id"`
	Name          string  `json:"name"`
	Description   *string `json:"description"`
	PriceCents    int64   `json:"price_cents"`
	Currency      string  `json:"currency"`
	CreditsAmount int64   `json:"credits_amount"`
}

// PlatformProductsResponse lists everything a platform can sell.
type PlatformProductsResponse struct {
	Plans       []PlanPrice  `json:"plans"`
	CreditPacks []CreditPack `json:"credit_packs"`
}

// PlatformCreditCheckoutRequest is a credit pack checkout for a linked user.
type PlatformCreditCheckoutRequest struct {
	UserEmail  string `json:"user_email"`
	PriceID    string `json:"price_id"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

// PlatformUserCreditsResponse is the remaining credits of an authorized user.
type PlatformUserCreditsResponse struct {
	UserID             string  `json:"user_id"`
	Email              string  `json:"email"`
	Balance            int64   `json:"balance"`
	MonthlyAllocation  int64   `json:"monthly_allocation"`
	UsedThisMonth      int64   `json:"used_this_month"`
	RemainingThisMonth int64   `json:"remaining_this_month"`
	ExtraCredits       int64   `json:"extra_credits"`
	ResetDate          *string `json:"reset_date"`
}

// ConfirmAuthorizationRequest confirms user authorization.
type ConfirmAuthorizationRequest struct {
	Code string `json:"code"`
}

// ConfirmAuthorizationResponse is the result of confirming authorization.
type ConfirmAuthorizationResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// StripeConnectStatusResponse is the Stripe Connect onboarding status.
type StripeConnectStatusResponse struct {
	AccountID          *string `json:"account_id,omitempty"`
	OnboardingComplete bool    `json:"onboarding_complete"`
	ChargesEnabled     bool    `json:"charges_enabled"`
}

// ErrorResponse is the API error response format.
type ErrorResponse struct {
	Error   string  `json:"error"`
	Details *string `json:"details,omitempty"`
}

// String is a helper that returns a pointer to the given string.
func String(s string) *string {
	return &s
}
