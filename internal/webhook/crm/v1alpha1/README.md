# CRM admission rules at a glance

Every rule that guards the `crm.miloapis.com` resources, in one place: what is
checked, what can change, when, and by whom. Keep it in step with the code when
a rule changes.

## Who enforces what

| Layer | Its job | Where |
|---|---|---|
| IAM | Who may write a resource at all | `config/components/iam/` (staff roles); the customer's access to their own onboarding is granted by CRM (#13) |
| Schema (CEL) | Field shapes, and what can never or no longer change | `+kubebuilder:validation:XValidation` markers in `api/crm/v1alpha1/` |
| Validating webhooks | References exist and agree with each other | rules in `internal/validation/crm/`, wired by this package |
| Defaulting webhook | Each new record is owned by the record it belongs to | `owners.go` |

The webhooks never check who is asking; that is IAM's job. They fail closed
(`failurePolicy=fail`): if a referenced record can't be read, the request is
refused.

**Who writes:**

- **Staff** write the spec of every kind except OnboardingInvitation, which
  only CRM creates. Staff roles can read status but not change it.
- **The customer** writes only their own OrganizationOnboarding's spec, once
  invited.
- **CRM** writes status, and creates invitations.

## Reference rules

A reference is checked when it is **set or changed**, never again afterwards.
Deleting a record later never blocks unrelated edits to the records that name
it.

| Rule | Checked |
|---|---|
| Every reference points to a record that exists: CRM records, Milo Organizations and Users, Billing's BillingAccounts | when set or changed |
| People on an Opportunity work for the Opportunity's Company | when a Person is added |
| An onboarding's Opportunity is with the onboarding's Company | on create |
| An invitation's Person works for the onboarding's Company | on create |
| An invitation's `recipientEmail` is its Person's email (ignoring case) | on create |
| An onboarding's `currentInvitationRef` is an invitation for that onboarding | when set or changed |
| An onboarding's `organizationRef` is linked in its Company's `organizationRefs`, so it can't record another customer's Organization | when set |
| An onboarding's `billingAccountRef` is in its organization's namespace (`organization-<name>`), so `organizationRef` must be set | when either changes |
| An invitation's `redeemedUserRef` is a User whose email is `recipientEmail` and who has verified it | when set |

Most cross references only need checking on create, because the references
they compare can never change (see below).

## What can change, and when

**Checked by** means: **CEL** is the schema; **webhook** is
`internal/validation/crm`.

### Company

| Field | Can change? | Checked by |
|---|---|---|
| `displayName`, `website` | Yes | CEL: not blank; an http(s) URL |
| `organizationRefs` | Yes, add or remove | webhook: each newly linked Organization exists |

### Person

| Field | Can change? | Checked by |
|---|---|---|
| `givenName`, `familyName`, `email` | Yes | CEL: not blank; an email address |
| `companyRef` | **Never**: someone who moves is a new Person | CEL: immutable; webhook: the Company exists (create) |
| `status.userRef` | CRM sets it once | CEL: can't change or be removed once set; webhook: the User exists |

### Opportunity

| Field | Can change? | Checked by |
|---|---|---|
| `displayName`, `description`, `stage` | Yes | CEL |
| `companyRef` | **Never**: a sale to another Company is a new Opportunity | CEL: immutable; webhook: the Company exists (create) |
| `ownerRef` | Yes | webhook: the User exists, when it changes |
| `personRefs` | Yes, add or remove | webhook: each added Person exists and works for the Opportunity's Company |

### OrganizationOnboarding

The spec gets stricter as the onboarding progresses. The phase is the one
recorded before the request:

| Field | Draft | Sent: `Invited`, `InProgress` | Submitted or ended: `Completing`, `Complete`, `Cancelled` |
|---|---|---|---|
| `companyRef`, `opportunityRef` | Never | Never | Never |
| `billing.currency`, `billing.paymentArrangement` | Yes | **No** | No |
| `organizationDetails`, `billing.contactInfo`, `billing.taxIds`, `billing.invoice` | Yes | Yes | **No** |

These locks are CEL rules on the OrganizationOnboarding. On create, the webhook
checks that the Company and Opportunity exist and that the Opportunity is with
that Company.

Status is written only by CRM:

| Field | Can change? | Checked by |
|---|---|---|
| `currentInvitationRef` | Yes: resend and replace move it | webhook: exists, and is an invitation for this onboarding |
| `organizationRef` | Set once | CEL: can't change or be removed; webhook: the Organization exists and the onboarding's Company links it (CRM links it there first) |
| `billingAccountRef` | Set once | CEL: can't change or be removed; webhook: exists in the organization's namespace |
| `invoiceAuthorization.userRef` | Yes | webhook: the User exists |

### OnboardingInvitation

Only CRM creates invitations.

| Field | Can change? | Checked by |
|---|---|---|
| `onboardingRef`, `personRef`, `recipientEmail` | **Never** | CEL: immutable; webhook on create: both exist, the Person works for the onboarding's Company, and `recipientEmail` is the Person's email |
| `status.inviterRef` | CRM sets it once | CEL: can't change or be removed; webhook: the User exists |
| `status.redeemedUserRef` | CRM sets it once | CEL: can't change or be removed; webhook: the User exists, has `recipientEmail` as their email, and has verified it |

## Ownership and deletion

Deleting is always allowed. CRM's records form one ownership tree, so Milo's
garbage collector deletes everything about a Company with it:

```
Company
├── Person
├── Opportunity
└── OrganizationOnboarding
    └── OnboardingInvitation
```

- **Setting the owner:** the defaulting webhook adds the owner reference on
  create (`owners.go`). It patches only `metadata.ownerReferences`.
- **Protecting it:** the validating webhooks reject any other owner on
  create, and any change to the owner references afterwards. This applies to
  everyone, so an orphaning delete (`--cascade=orphan`) of a Company or
  onboarding never finishes.
- **One owner per record:** the garbage collector only deletes a record once
  all its owners are gone. That is why an invitation is owned by its
  onboarding, not its Person; deleting a Person deletes their invitations
  through CRM's controller (TODO #13).
- **Never CRM-owned:** Organizations and BillingAccounts. Deleting a CRM
  record never deletes them.
