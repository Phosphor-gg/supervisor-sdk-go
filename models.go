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
	LabelSexualMinors ModerationLabel = "sexual/minors"
	LabelSexualExp    ModerationLabel = "sexual/explicit"
	LabelSensitive    ModerationLabel = "sensitive"
	LabelViolence     ModerationLabel = "violence"
	LabelSelfHarm     ModerationLabel = "self-harm"
	LabelMedical      ModerationLabel = "medical"
	LabelSpam         ModerationLabel = "spam"
	LabelPromotional  ModerationLabel = "promotional"
	LabelScam         ModerationLabel = "scam"
	LabelIllegal      ModerationLabel = "illegal"
	LabelPersonalData ModerationLabel = "personal-data"
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
	TierFree     Tier = "Free"
	TierBasic    Tier = "Basic"
	TierStandard Tier = "Standard"
	TierPremium  Tier = "Premium"
)

// BillingCycle represents billing period options.
type BillingCycle string

const (
	BillingMonthly    BillingCycle = "Monthly"
	BillingQuarterly  BillingCycle = "Quarterly"
	BillingAnnual     BillingCycle = "Annual"
	BillingTriennial  BillingCycle = "Triennial"
)

// ModerationRequest is the request body for POST /api/moderate.
type ModerationRequest struct {
	Text           *string           `json:"text,omitempty"`
	Image          *string           `json:"image,omitempty"`
	Model          *ModerationModel  `json:"model,omitempty"`
	EnabledLabels  []ModerationLabel `json:"enabled_labels,omitempty"`
	IncludeContext bool              `json:"include_context,omitempty"`
}

// BatchModerationRequest is the request body for POST /api/batch.
type BatchModerationRequest struct {
	Texts          []string          `json:"texts"`
	Model          *ModerationModel  `json:"model,omitempty"`
	EnabledLabels  []ModerationLabel `json:"enabled_labels,omitempty"`
	IncludeContext bool              `json:"include_context,omitempty"`
}

// UsernameCheckRequest is the request body for POST /api/username.
type UsernameCheckRequest struct {
	Username string `json:"username"`
}

// ModerationResponse is the result of a moderation request.
type ModerationResponse struct {
	Flagged        bool              `json:"flagged"`
	Labels         []ModerationLabel `json:"labels"`
	ImplicitLabels []ModerationLabel `json:"implicit_labels,omitempty"`
	ModelVersion   *string           `json:"model_version,omitempty"`
	NeedsContext   *bool             `json:"needs_context,omitempty"`
	ContextLabels  []ModerationLabel `json:"context_labels,omitempty"`
	RewrittenText  *string           `json:"rewritten_text,omitempty"`
}

// UsernameCheckResponse is the result of a username check.
type UsernameCheckResponse struct {
	Flagged bool    `json:"flagged"`
	Score   float64 `json:"score"`
}

// PartnerTokenRequest is the OAuth2 client credentials request.
type PartnerTokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
}

// PartnerTokenResponse is the OAuth2 access token response.
type PartnerTokenResponse struct {
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
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	IsNewAccount bool   `json:"is_new_account"`
	IsNewlyLinked bool  `json:"is_newly_linked"`
}

// PartnerUserInfo represents a partner's view of a linked user.
type PartnerUserInfo struct {
	UserID                string `json:"user_id"`
	Email                 string `json:"email"`
	LinkedAt              string `json:"linked_at"`
	Authorized            bool   `json:"authorized"`
	HasActiveSubscription bool   `json:"has_active_subscription"`
	Tier                  Tier   `json:"tier"`
}

// PartnerModerationRequest moderates content on behalf of a linked user.
type PartnerModerationRequest struct {
	UserEmail      string            `json:"user_email"`
	Text           *string           `json:"text,omitempty"`
	Image          *string           `json:"image,omitempty"`
	Model          *ModerationModel  `json:"model,omitempty"`
	EnabledLabels  []ModerationLabel `json:"enabled_labels,omitempty"`
	IncludeContext bool              `json:"include_context,omitempty"`
}

// PartnerCheckoutRequest creates a checkout session for a partner user.
type PartnerCheckoutRequest struct {
	UserEmail    string       `json:"user_email"`
	Tier         Tier         `json:"tier"`
	BillingCycle BillingCycle `json:"billing_cycle"`
	SuccessURL   string       `json:"success_url"`
	CancelURL    string       `json:"cancel_url"`
}

// PartnerCheckoutResponse contains the checkout URL.
type PartnerCheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
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
