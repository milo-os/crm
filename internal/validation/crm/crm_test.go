// SPDX-License-Identifier: AGPL-3.0-only

package crm

import (
	"context"
	"errors"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	billingv1alpha1 "go.miloapis.com/billing/api/v1alpha1"
	iamv1alpha1 "go.miloapis.com/milo/pkg/apis/iam/v1alpha1"
	resourcemanagerv1alpha1 "go.miloapis.com/milo/pkg/apis/resourcemanager/v1alpha1"

	crmv1alpha1 "go.miloapis.com/crm/api/crm/v1alpha1"
)

const staffUser = "340583683847097153"

func scheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := crmv1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	for _, add := range []func(*runtime.Scheme) error{
		resourcemanagerv1alpha1.AddToScheme, iamv1alpha1.AddToScheme, billingv1alpha1.AddToScheme,
	} {
		if err := add(s); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

func organization(name string) *resourcemanagerv1alpha1.Organization {
	return &resourcemanagerv1alpha1.Organization{ObjectMeta: metav1.ObjectMeta{Name: name}}
}

func user(name, email string, verification iamv1alpha1.EmailVerificationState) *iamv1alpha1.User {
	return &iamv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       iamv1alpha1.UserSpec{Email: email},
		Status:     iamv1alpha1.UserStatus{EmailVerification: verification},
	}
}

func billingAccount(organization, name string) *billingv1alpha1.BillingAccount {
	return &billingv1alpha1.BillingAccount{ObjectMeta: metav1.ObjectMeta{Namespace: organizationNamespace(organization), Name: name}}
}

// records holds acme and globex, Jane at acme, Hank at globex, an Opportunity
// and an onboarding with acme, an invitation to it, the Milo organization
// acme with a billing account, and Users: staff, Jane (verified), and Hank
// (unverified).
func records(t *testing.T) client.Reader {
	t.Helper()
	return fake.NewClientBuilder().WithScheme(scheme(t)).WithObjects(
		company("acme"), company("globex"),
		person("jane", "acme"), person("hank", "globex"),
		opportunity("acme-sale", "acme"),
		onboarding("acme-onboarding", "acme", "acme-sale", ""),
		onboarding("globex-onboarding", "globex", "globex-sale", ""),
		invitation("acme-onboarding", "jane"),
		organization("acme"),
		billingAccount("acme", "acme-billing"),
		user(staffUser, "alice@datum.example", iamv1alpha1.EmailVerificationStateVerified),
		user("jane-user", "Jane@Example.com", iamv1alpha1.EmailVerificationStateVerified),
		user("hank-user", "hank@example.com", iamv1alpha1.EmailVerificationStateUnverified),
	).Build()
}

// failingReader fails every read, as when the API server can't be reached.
func failingReader(t *testing.T) client.Reader {
	t.Helper()
	return fake.NewClientBuilder().WithScheme(scheme(t)).WithInterceptorFuncs(interceptor.Funcs{
		Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
			return errors.New("connection refused")
		},
	}).Build()
}

func company(name string) *crmv1alpha1.Company {
	return &crmv1alpha1.Company{ObjectMeta: metav1.ObjectMeta{Name: name}, Spec: crmv1alpha1.CompanySpec{DisplayName: name}}
}

func person(name, company string) *crmv1alpha1.Person {
	return &crmv1alpha1.Person{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: crmv1alpha1.PersonSpec{
			GivenName: name, Email: name + "@example.com",
			CompanyRef: crmv1alpha1.CompanyReference{Name: company},
		},
	}
}

func opportunity(name, company string, people ...string) *crmv1alpha1.Opportunity {
	o := &crmv1alpha1.Opportunity{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: crmv1alpha1.OpportunitySpec{
			DisplayName: name,
			CompanyRef:  crmv1alpha1.CompanyReference{Name: company},
			OwnerRef:    crmv1alpha1.UserReference{Name: staffUser},
		},
	}
	for _, p := range people {
		o.Spec.PersonRefs = append(o.Spec.PersonRefs, crmv1alpha1.PersonReference{Name: p})
	}
	return o
}

func onboarding(name, company, opportunity string, phase crmv1alpha1.OnboardingPhase) *crmv1alpha1.OrganizationOnboarding {
	return &crmv1alpha1.OrganizationOnboarding{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: crmv1alpha1.OrganizationOnboardingSpec{
			CompanyRef:     crmv1alpha1.CompanyReference{Name: company},
			OpportunityRef: crmv1alpha1.OpportunityReference{Name: opportunity},
			Billing:        crmv1alpha1.OnboardingBilling{Currency: crmv1alpha1.BillingCurrencyUSD},
		},
		Status: crmv1alpha1.OrganizationOnboardingStatus{Phase: phase},
	}
}

func invitation(onboarding, person string) *crmv1alpha1.OnboardingInvitation {
	return &crmv1alpha1.OnboardingInvitation{
		ObjectMeta: metav1.ObjectMeta{Name: onboarding + "-" + person},
		Spec: crmv1alpha1.OnboardingInvitationSpec{
			OnboardingRef:  crmv1alpha1.OrganizationOnboardingReference{Name: onboarding},
			PersonRef:      crmv1alpha1.PersonReference{Name: person},
			RecipientEmail: person + "@example.com",
		},
	}
}

// withRecipient sets an invitation's recipient email.
func withRecipient(i *crmv1alpha1.OnboardingInvitation, email string) *crmv1alpha1.OnboardingInvitation {
	i.Spec.RecipientEmail = email
	return i
}

// assertErrors checks that errs is empty when want is empty, and otherwise
// that it contains each wanted substring.
func assertErrors(t *testing.T, errs field.ErrorList, want ...string) {
	t.Helper()
	if len(want) == 0 {
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs.ToAggregate())
		}
		return
	}
	if len(errs) == 0 {
		t.Fatalf("expected errors containing %q, got none", want)
	}
	got := errs.ToAggregate().Error()
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("errors %q don't contain %q", got, w)
		}
	}
}

func TestValidatePerson(t *testing.T) {
	ctx := context.Background()
	c := records(t)

	t.Run("create with an existing Company", func(t *testing.T) {
		assertErrors(t, ValidatePersonCreate(ctx, c, person("ann", "acme")))
	})
	t.Run("create with a missing Company", func(t *testing.T) {
		assertErrors(t, ValidatePersonCreate(ctx, c, person("ann", "initech")),
			`spec.companyRef.name: Invalid value: "initech": Company not found`)
	})
	t.Run("update that keeps a Company deleted since is allowed", func(t *testing.T) {
		old := person("ann", "initech")
		updated := person("ann", "initech")
		updated.Spec.Email = "ann@initech.example"
		assertErrors(t, ValidatePersonUpdate(ctx, c, old, updated))
	})
	t.Run("a read failure refuses the request", func(t *testing.T) {
		assertErrors(t, ValidatePersonCreate(ctx, failingReader(t), person("ann", "acme")),
			"Internal error", "connection refused")
	})
}

func TestValidateOpportunity(t *testing.T) {
	ctx := context.Background()
	c := records(t)

	tests := []struct {
		name    string
		old     *crmv1alpha1.Opportunity
		updated *crmv1alpha1.Opportunity
		want    []string
	}{
		{name: "create with an existing Company and its People", updated: opportunity("deal", "acme", "jane")},
		{name: "create with a missing Company", updated: opportunity("deal", "initech"),
			want: []string{`spec.companyRef.name: Invalid value: "initech": Company not found`}},
		{name: "create with a missing Person", updated: opportunity("deal", "acme", "nobody"),
			want: []string{`spec.personRefs[0].name: Invalid value: "nobody": Person not found`}},
		{name: "create with a Person from another Company", updated: opportunity("deal", "acme", "hank"),
			want: []string{`spec.personRefs[0].name`, `Person works for Company "globex", not the Opportunity's Company "acme"`}},
		{name: "update adding a Person checks only the new one",
			old: opportunity("deal", "acme", "gone"), updated: opportunity("deal", "acme", "gone", "jane")},
		{name: "update adding a Person from another Company",
			old: opportunity("deal", "acme", "jane"), updated: opportunity("deal", "acme", "jane", "hank"),
			want: []string{`spec.personRefs[1].name`, `works for Company "globex"`}},
		{name: "update that keeps a Company deleted since is allowed",
			old: opportunity("deal", "initech"), updated: opportunity("deal", "initech")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.old == nil {
				assertErrors(t, ValidateOpportunityCreate(ctx, c, tt.updated), tt.want...)
				return
			}
			assertErrors(t, ValidateOpportunityUpdate(ctx, c, tt.old, tt.updated), tt.want...)
		})
	}
}

func TestValidateOrganizationOnboardingCreate(t *testing.T) {
	ctx := context.Background()
	c := fake.NewClientBuilder().WithScheme(scheme(t)).WithObjects(
		company("acme"), company("globex"), opportunity("acme-sale", "acme"),
	).Build()

	tests := []struct {
		name       string
		onboarding *crmv1alpha1.OrganizationOnboarding
		want       []string
	}{
		{name: "existing Company and its Opportunity", onboarding: onboarding("o", "acme", "acme-sale", "")},
		{name: "missing Company", onboarding: onboarding("o", "initech", "acme-sale", ""),
			want: []string{`spec.companyRef.name: Invalid value: "initech": Company not found`}},
		{name: "missing Opportunity", onboarding: onboarding("o", "acme", "nothing", ""),
			want: []string{`spec.opportunityRef.name: Invalid value: "nothing": Opportunity not found`}},
		{name: "Opportunity with another Company", onboarding: onboarding("o", "globex", "acme-sale", ""),
			want: []string{`Opportunity is with Company "acme", not the onboarding's Company "globex"`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertErrors(t, ValidateOrganizationOnboardingCreate(ctx, c, tt.onboarding), tt.want...)
		})
	}
}

func TestValidateOrganizationOnboardingUpdate(t *testing.T) {
	// The webhook doesn't restrict who changes the details, or when: IAM
	// decides who can write, and the schema locks the currency and payment
	// arrangement once the onboarding is sent.
	for _, phase := range []crmv1alpha1.OnboardingPhase{"", crmv1alpha1.OnboardingPhaseDraft, crmv1alpha1.OnboardingPhaseInvited, crmv1alpha1.OnboardingPhaseInProgress} {
		t.Run("details change while "+string(phase), func(t *testing.T) {
			old := onboarding("o", "acme", "acme-sale", phase)
			updated := old.DeepCopy()
			updated.Spec.OrganizationDetails = &crmv1alpha1.OnboardingOrganizationDetails{DisplayName: "Acme Corp"}
			assertErrors(t, ValidateOrganizationOnboardingUpdate(context.Background(), records(t), old, updated))
		})
	}
}

func TestValidateOnboardingInvitationCreate(t *testing.T) {
	ctx := context.Background()
	c := records(t)

	tests := []struct {
		name       string
		invitation *crmv1alpha1.OnboardingInvitation
		want       []string
	}{
		{name: "existing onboarding and a Person of its Company", invitation: invitation("acme-onboarding", "jane")},
		{name: "missing onboarding", invitation: invitation("nothing", "jane"),
			want: []string{`spec.onboardingRef.name: Invalid value: "nothing": OrganizationOnboarding not found`}},
		{name: "missing Person", invitation: invitation("acme-onboarding", "nobody"),
			want: []string{`spec.personRef.name: Invalid value: "nobody": Person not found`}},
		{name: "both missing reports both", invitation: invitation("nothing", "nobody"),
			want: []string{"OrganizationOnboarding not found", "Person not found"}},
		{name: "Person from another Company", invitation: invitation("acme-onboarding", "hank"),
			want: []string{`Person works for Company "globex", not the onboarding's Company "acme"`}},
		{name: "recipientEmail matches the Person's email ignoring case", invitation: withRecipient(invitation("acme-onboarding", "jane"), "JANE@example.com")},
		{name: "recipientEmail isn't the Person's email", invitation: withRecipient(invitation("acme-onboarding", "jane"), "someone@example.com"),
			want: []string{`spec.recipientEmail: Invalid value: "someone@example.com": must be the Person's email, jane@example.com`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertErrors(t, ValidateOnboardingInvitationCreate(ctx, c, tt.invitation), tt.want...)
		})
	}
}

func TestValidateCompany(t *testing.T) {
	ctx := context.Background()
	c := records(t)
	withOrganizations := func(names ...string) *crmv1alpha1.Company {
		co := company("acme")
		for _, n := range names {
			co.Spec.OrganizationRefs = append(co.Spec.OrganizationRefs, crmv1alpha1.OrganizationReference{Name: n})
		}
		return co
	}

	t.Run("create linking an existing organization", func(t *testing.T) {
		assertErrors(t, ValidateCompanyCreate(ctx, c, withOrganizations("acme")))
	})
	t.Run("create linking a missing organization", func(t *testing.T) {
		assertErrors(t, ValidateCompanyCreate(ctx, c, withOrganizations("acme", "org-8f3k2")),
			`spec.organizationRefs[1].name: Invalid value: "org-8f3k2": Organization not found`)
	})
	t.Run("update checks only newly linked organizations", func(t *testing.T) {
		assertErrors(t, ValidateCompanyUpdate(ctx, c, withOrganizations("gone"), withOrganizations("gone", "acme")))
	})
	t.Run("update linking a missing organization", func(t *testing.T) {
		assertErrors(t, ValidateCompanyUpdate(ctx, c, withOrganizations("acme"), withOrganizations("acme", "org-8f3k2")),
			`spec.organizationRefs[1].name: Invalid value: "org-8f3k2": Organization not found`)
	})
}

func TestValidateOpportunityOwner(t *testing.T) {
	ctx := context.Background()
	c := records(t)
	withOwner := func(owner string) *crmv1alpha1.Opportunity {
		o := opportunity("deal", "acme")
		o.Spec.OwnerRef.Name = owner
		return o
	}

	t.Run("create with an existing owner", func(t *testing.T) {
		assertErrors(t, ValidateOpportunityCreate(ctx, c, withOwner(staffUser)))
	})
	t.Run("create with a missing owner", func(t *testing.T) {
		assertErrors(t, ValidateOpportunityCreate(ctx, c, withOwner("999")),
			`spec.ownerRef.name: Invalid value: "999": User not found`)
	})
	t.Run("update that keeps an owner deleted since is allowed", func(t *testing.T) {
		assertErrors(t, ValidateOpportunityUpdate(ctx, c, withOwner("999"), withOwner("999")))
	})
	t.Run("update to a missing owner", func(t *testing.T) {
		assertErrors(t, ValidateOpportunityUpdate(ctx, c, withOwner(staffUser), withOwner("999")), "User not found")
	})
}

func TestValidatePersonUserLink(t *testing.T) {
	ctx := context.Background()
	c := records(t)
	linked := func(userName string) *crmv1alpha1.Person {
		p := person("jane", "acme")
		if userName != "" {
			p.Status.UserRef = &crmv1alpha1.UserReference{Name: userName}
		}
		return p
	}

	t.Run("linking an existing User", func(t *testing.T) {
		assertErrors(t, ValidatePersonUpdate(ctx, c, linked(""), linked("jane-user")))
	})
	t.Run("linking a missing User", func(t *testing.T) {
		assertErrors(t, ValidatePersonUpdate(ctx, c, linked(""), linked("999")),
			`status.userRef.name: Invalid value: "999": User not found`)
	})
	t.Run("keeping a link to a User deleted since is allowed", func(t *testing.T) {
		assertErrors(t, ValidatePersonUpdate(ctx, c, linked("999"), linked("999")))
	})
}

func TestValidateOnboardingStatus(t *testing.T) {
	ctx := context.Background()
	c := records(t)
	with := func(change func(*crmv1alpha1.OrganizationOnboardingStatus)) *crmv1alpha1.OrganizationOnboarding {
		o := onboarding("acme-onboarding", "acme", "acme-sale", crmv1alpha1.OnboardingPhaseInvited)
		change(&o.Status)
		return o
	}
	none := func(*crmv1alpha1.OrganizationOnboardingStatus) {}

	tests := []struct {
		name   string
		change func(*crmv1alpha1.OrganizationOnboardingStatus)
		want   []string
	}{
		{name: "current invitation for this onboarding", change: func(s *crmv1alpha1.OrganizationOnboardingStatus) {
			s.CurrentInvitationRef = &crmv1alpha1.OnboardingInvitationReference{Name: "acme-onboarding-jane"}
		}},
		{name: "missing current invitation", change: func(s *crmv1alpha1.OrganizationOnboardingStatus) {
			s.CurrentInvitationRef = &crmv1alpha1.OnboardingInvitationReference{Name: "nothing"}
		}, want: []string{`status.currentInvitationRef.name: Invalid value: "nothing": OnboardingInvitation not found`}},
		{name: "existing organization and its billing account", change: func(s *crmv1alpha1.OrganizationOnboardingStatus) {
			s.OrganizationRef = &crmv1alpha1.OrganizationReference{Name: "acme"}
			s.BillingAccountRef = &crmv1alpha1.BillingAccountReference{Name: "acme-billing"}
		}},
		{name: "missing organization", change: func(s *crmv1alpha1.OrganizationOnboardingStatus) {
			s.OrganizationRef = &crmv1alpha1.OrganizationReference{Name: "org-8f3k2"}
		}, want: []string{`status.organizationRef.name: Invalid value: "org-8f3k2": Organization not found`}},
		{name: "billing account missing from the organization's namespace", change: func(s *crmv1alpha1.OrganizationOnboardingStatus) {
			s.OrganizationRef = &crmv1alpha1.OrganizationReference{Name: "acme"}
			s.BillingAccountRef = &crmv1alpha1.BillingAccountReference{Name: "other-billing"}
		}, want: []string{`status.billingAccountRef.name: Invalid value: "other-billing": BillingAccount not found in namespace organization-acme`}},
		{name: "billing account without an organization", change: func(s *crmv1alpha1.OrganizationOnboardingStatus) {
			s.BillingAccountRef = &crmv1alpha1.BillingAccountReference{Name: "acme-billing"}
		}, want: []string{"a billing account needs status.organizationRef"}},
		{name: "invoice authorized by an existing User", change: func(s *crmv1alpha1.OrganizationOnboardingStatus) {
			s.InvoiceAuthorization = &crmv1alpha1.InvoiceAuthorization{UserRef: crmv1alpha1.UserReference{Name: staffUser}, Time: metav1.Now()}
		}},
		{name: "invoice authorized by a missing User", change: func(s *crmv1alpha1.OrganizationOnboardingStatus) {
			s.InvoiceAuthorization = &crmv1alpha1.InvoiceAuthorization{UserRef: crmv1alpha1.UserReference{Name: "999"}, Time: metav1.Now()}
		}, want: []string{`status.invoiceAuthorization.userRef.name: Invalid value: "999": User not found`}},
		{name: "nothing changed", change: none},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := with(none)
			assertErrors(t, ValidateOrganizationOnboardingUpdate(ctx, c, old, with(tt.change)), tt.want...)
		})
	}

	t.Run("current invitation for another onboarding", func(t *testing.T) {
		other := onboarding("globex-onboarding", "globex", "globex-sale", crmv1alpha1.OnboardingPhaseInvited)
		updated := other.DeepCopy()
		updated.Status.CurrentInvitationRef = &crmv1alpha1.OnboardingInvitationReference{Name: "acme-onboarding-jane"}
		assertErrors(t, ValidateOrganizationOnboardingUpdate(ctx, c, other, updated),
			`OnboardingInvitation is for onboarding "acme-onboarding", not "globex-onboarding"`)
	})
	t.Run("references kept after they are deleted are allowed", func(t *testing.T) {
		old := with(func(s *crmv1alpha1.OrganizationOnboardingStatus) {
			s.OrganizationRef = &crmv1alpha1.OrganizationReference{Name: "gone"}
			s.BillingAccountRef = &crmv1alpha1.BillingAccountReference{Name: "gone-billing"}
		})
		updated := old.DeepCopy()
		updated.Status.Phase = crmv1alpha1.OnboardingPhaseInProgress
		assertErrors(t, ValidateOrganizationOnboardingUpdate(ctx, c, old, updated))
	})
}

func TestValidateOnboardingInvitationUpdate(t *testing.T) {
	ctx := context.Background()
	c := records(t)
	with := func(inviter, redeemer string) *crmv1alpha1.OnboardingInvitation {
		i := invitation("acme-onboarding", "jane")
		if inviter != "" {
			i.Status.InviterRef = &crmv1alpha1.UserReference{Name: inviter}
		}
		if redeemer != "" {
			i.Status.RedeemedUserRef = &crmv1alpha1.UserReference{Name: redeemer}
		}
		return i
	}

	tests := []struct {
		name    string
		old     *crmv1alpha1.OnboardingInvitation
		updated *crmv1alpha1.OnboardingInvitation
		want    []string
	}{
		{name: "existing inviter", old: with("", ""), updated: with(staffUser, "")},
		{name: "missing inviter", old: with("", ""), updated: with("999", ""),
			want: []string{`status.inviterRef.name: Invalid value: "999": User not found`}},
		{name: "redeemed by the verified recipient, email compared without case", old: with(staffUser, ""), updated: with(staffUser, "jane-user")},
		{name: "redeemed by a missing User", old: with(staffUser, ""), updated: with(staffUser, "999"),
			want: []string{`status.redeemedUserRef.name: Invalid value: "999": User not found`}},
		{name: "redeemed by someone else", old: with(staffUser, ""), updated: with(staffUser, staffUser),
			want: []string{"User's email isn't the invitation's recipientEmail"}},
		{name: "redeemed with an unverified email", old: with(staffUser, ""), updated: func() *crmv1alpha1.OnboardingInvitation {
			i := with(staffUser, "hank-user")
			i.Spec.RecipientEmail = "hank@example.com"
			return i
		}(), want: []string{"User hasn't verified their email"}},
		{name: "unchanged references aren't checked again", old: with("999", "999"), updated: with("999", "999")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertErrors(t, ValidateOnboardingInvitationUpdate(ctx, c, tt.old, tt.updated), tt.want...)
		})
	}
}
