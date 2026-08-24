# Savio — Business Requirements

## Related Documents

- [README.md](../../README.md) — project overview, setup, and documentation index.
- [Product Foundation](product-foundation.md) — product vision, positioning, and principles.
- [User Flows](user-flows.md) — end-to-end user workflows.
- [Database Design](../database/database-design.md) — persistence of authoritative financial data.
- [API Contract](../api/api-contract.md) — REST endpoints that expose these rules.
- [Security Architecture](../engineering/security.md) — enforcement of business security requirements.

## 1. Document Purpose

This document defines the business rules, domain behavior, lifecycle rules, validations, permissions, and system constraints for Savio.

The purpose of this document is to translate the product foundation into concrete and testable application behavior.

This document is the business-rule source of truth for implementation.

The main principle remains:

> **Finance Engine calculates. AI interprets. User decides.**

Business-critical financial calculations must remain deterministic and testable.

AI may assist users, explain results, classify information, and generate suggestions, but AI must not become the authoritative source for balances, budgets, forecasts, or financial state transitions.

---

# 2. Domain Overview

Savio consists of the following main domains:

```text
Identity & Access
│
├── Users
├── Authentication
└── Sessions

Financial Core
│
├── Accounts
├── Categories
├── Transactions
├── Transfers
└── Recurring Transactions

Planning
│
├── Budgets
├── Financial Goals
└── Scheduled Financial Events

Intelligence
│
├── Cashflow Analytics
├── Forecast Engine
├── Scenario Simulator
└── Financial Health

AI
│
├── AI Insights
├── AI Transaction Categorization
├── AI Copilot
└── AI Feedback

Supporting
│
├── Notifications
├── Audit Logs
└── User Settings
# 3. Core Business Principles
## 3.1 Data Ownership

All financial data belongs to a user.

A user must not access another user's:

accounts,
transactions,
budgets,
recurring transactions,
goals,
forecasts,
scenarios,
AI insights,
reports,
or financial settings.

Authorization must always be enforced by the backend.

Frontend filtering is not considered a security mechanism.

## 3.2 Financial Amount Precision

Financial amounts must not use floating-point arithmetic for authoritative calculations.

Recommended implementation:

Database:
BIGINT integer minor units

Application:
decimal-safe representation over integer minor units

API:
decimal-safe strings converted to/from minor units

Examples:

Rp10,000
→ 1000000 minor units (2-decimal scale)

USD 10.50
→ 1050 minor units

The implementation strategy must be consistent across the system.

## 3.3 Currency

Each financial account has a currency.

Initial MVP:

Default:
IDR

Multi-currency support is not required for the first implementation.

A user may still have a currency preference field for future extensibility.

Cross-currency transfer is out of MVP scope unless explicitly implemented.

## 3.4 Timezone

Every user has a timezone.

Financial dates and recurring schedules must be interpreted using the user's timezone.

Stored timestamps should use UTC.

Example:

User timezone:
Asia/Jakarta

Transaction entered:
2026-08-25 08:00 WIB

Stored:
UTC timestamp

Date-only financial events should preserve the user's intended local date.

## 3.5 Soft Delete vs Financial History

Financial records must not silently disappear if deleting them would invalidate financial history.

Deletion behavior must depend on the resource.

Examples:

Unused category
→ can be deleted

Category already used
→ archive instead

Account with transactions
→ archive instead

Posted transaction
→ may be voided and replaced

The system must preserve financial integrity.

# 4. Users
## 4.1 User Fields

A user contains at minimum:

id
name
email
password_hash
timezone
default_currency
status
created_at
updated_at

Optional profile fields may include:

avatar
locale
preferred_date_format
preferred_number_format
## 4.2 User Status

Supported status:

ACTIVE
DISABLED

Rules:

ACTIVE users may authenticate.
DISABLED users may not authenticate.
Disabling a user invalidates active sessions.
## 4.3 Email

Email must:

be valid,
be normalized,
be unique,
and be treated case-insensitively where possible.

Example:

USER@EXAMPLE.COM
user@example.com

must not create duplicate identities.

# 5. Authentication
## 5.1 Registration

A user may register using:

name
email
password
password_confirmation

Rules:

email must be unique,
password must satisfy password policy,
password confirmation must match,
password must never be stored in plaintext.

Successful registration may automatically authenticate the user.

## 5.2 Login

Login requires:

email
password

Successful login creates an authenticated session.

Authentication uses secure cookies.

Authentication tokens must not be stored in:

localStorage
sessionStorage
## 5.3 Access Session

The access credential should be short-lived.

Example target:

~15 minutes

The exact duration is configuration-driven.

## 5.4 Refresh Session

Refresh credentials must support:

expiration,
revocation,
rotation,
server-side session tracking.

A secure design may use:

opaque random refresh token

with only its hash stored in the database.

## 5.5 Refresh Rotation

When a refresh succeeds:

Old refresh token
→ revoked / rotated

New refresh token
→ issued

A refresh token must not remain indefinitely reusable after rotation.

## 5.6 Logout

Logout must:

invalidate the current session,
clear authentication cookies,
prevent future refresh using the revoked session.
## 5.7 Logout All Sessions

The user may revoke all active sessions.

Example:

Laptop
Phone
Tablet

After revoke-all:

all refresh sessions
→ revoked

except optionally the current session if the implementation explicitly provides that behavior.

# 6. Accounts
## 6.1 Purpose

Accounts represent places where money is held.

Examples:

Cash
Bank
Savings
E-Wallet
Other
## 6.2 Account Fields

An account contains:

id
user_id
name
type
currency
initial_balance
current_balance
status
created_at
updated_at

Optional:

description
icon
institution_name
## 6.3 Account Types

Initial supported types:

CASH
BANK
EWALLET
SAVINGS
OTHER

The type does not directly change financial calculation behavior in MVP.

## 6.4 Initial Balance

When creating an account, a user may define an initial balance.

The implementation must maintain a clear distinction between:

initial_balance
transaction activity
current balance

Current balance must not become an arbitrary editable value after transactions exist.

## 6.5 Current Balance

Current balance must be derived consistently.

Conceptually:

Current Balance
=
Initial Balance
+ Income
- Expense
+ Incoming Transfers
- Outgoing Transfers
+ Adjustments

The implementation may maintain a cached balance for performance, but transactional integrity must remain correct.

## 6.6 Account Archival

An account with transaction history should normally be archived rather than deleted.

Status:

ACTIVE
ARCHIVED

Archived accounts:

remain visible in historical reports where relevant,
cannot receive new transactions by default,
may be restored if supported.
## 6.7 Account Deletion

An account may only be hard-deleted if:

transaction_count = 0
AND
transfer_count = 0
AND
no active dependency

Otherwise the system should return a conflict.

Example:

# 409 ACCOUNT_HAS_FINANCIAL_HISTORY
# 7. Categories
## 7.1 Category Types

Categories belong to one of:

INCOME
EXPENSE

A category cannot be used for the opposite transaction type.

Example:

Salary
→ INCOME

Food
→ EXPENSE
## 7.2 System Categories

Savio may seed default system categories.

Examples:

Income
Salary
Freelance
Business
Interest
Gift
Other Income
Expense
Food & Dining
Transport
Housing
Utilities
Shopping
Entertainment
Healthcare
Education
Subscriptions
Other Expense
## 7.3 Custom Categories

Users may create custom categories.

Custom category requirements:

name
type

Optional:

parent_category_id
icon
description
## 7.4 Category Uniqueness

Within a user:

category name + category type

should normally be unique.

Case-insensitive uniqueness is preferred.

## 7.5 Category Archival

Categories already referenced by transactions must not be hard-deleted.

They should be archived.

Archived categories:

remain attached to historical transactions,
are not shown as default options for new transactions.
# 8. Transactions
## 8.1 Transaction Types

Supported transaction types:

INCOME
EXPENSE
TRANSFER
ADJUSTMENT

TRANSFER may be represented internally using a dedicated transfer entity and linked transaction entries.

## 8.2 Transaction Fields

A standard income or expense transaction contains:

id
user_id
account_id
category_id
type
amount
transaction_date
description
notes
status
created_at
updated_at

Optional:

merchant
tags
recurring_transaction_id
source
## 8.3 Amount Rules

Transaction amount must:

amount > 0

Transaction direction is determined by type.

Do not represent expense using negative input amounts.

Example:

type = EXPENSE
amount = 50000

not:

amount = -50000

This avoids ambiguous input semantics.

## 8.4 Income

An income transaction increases account balance.

Account
+ amount

Example:

Salary
Rp12,000,000
## 8.5 Expense

An expense transaction decreases account balance.

Account
- amount

Example:

Food
Rp75,000
## 8.6 Transaction Category

Rules:

INCOME transaction
→ requires INCOME category

EXPENSE transaction
→ requires EXPENSE category

The backend must reject mismatched category types.

Example:

Expense + Salary category
→ 422 INVALID_CATEGORY_TYPE
## 8.7 Transaction Account Ownership

The account used by a transaction must belong to the authenticated user.

A valid UUID alone is insufficient.

## 8.8 Transaction Status

A transaction uses one status:

DRAFT
→ POSTED
→ VOIDED

DRAFT
A pending transaction that has not yet taken financial effect.
May still be edited.

POSTED
The transaction is financially effective and updates the account balance.
Posted financial fields (amount, account, type, date) are immutable.

VOIDED
The transaction is invalidated without being hard-deleted.
It is excluded from active income/expense analytics and its balance effect is reversed.

Correction is performed by voiding the original and creating a replacement transaction, preserving audit history.

## 8.9 Transaction Date

Transaction date:

cannot be invalid,
may represent past transactions,
may represent current transactions.

Future transactions should normally use the scheduled financial event system rather than posting immediately.

## 8.10 Transaction Editing

A still-pending DRAFT transaction may be edited before it is posted.

Posted transactions are financially immutable.

To change a posted transaction:

- void the original,
- create a replacement transaction,
- recompute the derived balance atomically,
- preserve audit history.

Example:

Old expense:
Rp100,000

Replacement expense:
Rp150,000

Net account effect:

- reversal of the voided original:
+Rp100,000

- posting of the replacement:
-Rp150,000

must both be reflected atomically.

Never silently overwrite a posted transaction's financial fields.

## 8.11 Transaction Voiding

Voiding must reverse the original balance effect atomically and preserve the historical record.

Example:

Void a Rp100,000 expense
→ account balance +Rp100,000

This operation must be transactional.

A VOIDED transaction must not be voided twice.

Voiding is preferred over hard deletion for auditability.

## 8.12 Transaction Adjustment

ADJUSTMENT is intended for explicit reconciliation of account balance.

Example:

Recorded balance:
Rp4,800,000

Actual balance:
Rp5,000,000

Adjustment:
+Rp200,000

Adjustments must:

require a reason,
be auditable,
not silently overwrite previous financial records.
# 9. Transfers
## 9.1 Purpose

Transfer moves money between two user-owned accounts.

Example:

Bank BCA
→ Rp1,000,000
→ GoPay

A transfer must not be treated as income or expense in aggregate financial analytics.

## 9.2 Transfer Fields

A transfer contains:

id
user_id
source_account_id
destination_account_id
amount
transfer_date
description
created_at
updated_at
## 9.3 Transfer Rules

Rules:

source_account_id != destination_account_id
amount > 0
same owner
accounts active
## 9.4 Transfer Balance Effect

Transfer must execute atomically.

Source account
- amount

Destination account
+ amount

Both operations succeed or neither succeeds.

## 9.5 Transfer Analytics

Transfers must not inflate:

total income
total expense
savings rate

because money remains within the user's financial system.

## 9.6 Transfer Voiding

Voiding a transfer must atomically reverse both account effects.

# 10. Recurring Transactions
## 10.1 Purpose

Recurring transactions represent predictable future financial activity.

Examples:

Salary
Rent
Internet
Netflix
Loan installment
Monthly saving contribution
## 10.2 Recurring Types

Recurring transaction type:

INCOME
EXPENSE

Transfers may be supported later.

## 10.3 Frequency

Initial frequency options:

DAILY
WEEKLY
MONTHLY
YEARLY

Optional future support:

CUSTOM
## 10.4 Recurring Fields

A recurring transaction contains:

id
user_id
account_id
category_id
type
amount
frequency
start_date
next_occurrence_date
end_date
status
description
created_at
updated_at

Optional:

interval
day_of_month
day_of_week
auto_post
## 10.5 Recurring Status

Supported status:

ACTIVE
PAUSED
ENDED
## 10.6 Recurring Lifecycle

Typical lifecycle:

ACTIVE
→ PAUSED
→ ACTIVE

or:

ACTIVE
→ ENDED

or:

ACTIVE
→ PAUSED
→ ENDED
## 10.7 Scheduled Occurrences

Recurring rules generate future occurrences.

The implementation must prevent duplicate generation.

Each generated occurrence should have a deterministic reference to its recurring source.

Example uniqueness concept:

recurring_transaction_id
+
occurrence_date

must not produce duplicate posted events.

## 10.8 Automatic vs Manual Posting

Savio may support:

AUTO_POST
MANUAL_CONFIRM

For MVP, a simpler policy may be selected.

If AUTO_POST is implemented, background processing must be idempotent.

## 10.9 Recurring Forecast

Recurring transactions are always important inputs to the forecast engine.

A recurring rule may affect forecast even before the occurrence becomes a posted transaction.

# 11. Budgets
## 11.1 Purpose

Budgets define intended spending limits for a financial period.

Initial budget scope:

Monthly category budget
## 11.2 Budget Fields

A budget contains:

id
user_id
category_id
amount
period_type
start_date
end_date
status
created_at
updated_at

Initial period type:

MONTHLY
## 11.3 Budget Category

Budget categories must be expense categories.

Income categories cannot have expense budgets.

## 11.4 Budget Uniqueness

A user should not have multiple conflicting active budgets for the same:

category
+
period

unless an explicit versioning model is implemented.

## 11.5 Budget Utilization

Budget utilization:

actual expense in category
÷
budget amount

Example:

Budget:
Rp2,000,000

Spent:
Rp1,500,000

Utilization:
75%
## 11.6 Budget Remaining
Remaining
=
Budget
-
Actual Expense

A negative value indicates overspending.

## 11.7 Budget Status

Possible computed status:

ON_TRACK
WARNING
EXCEEDED

Example thresholds may be:

0–79%     ON_TRACK
80–99%    WARNING
>=100%    EXCEEDED

Thresholds should be configuration-driven or clearly documented.

## 11.8 Projected Overspend

The forecast engine may estimate whether a category is likely to exceed budget.

Example:

Spent:
Rp1.4M

Current day:
15 / 30

Projected:
Rp2.8M

Budget:
Rp2M

Result:

Projected overspend:
Rp800k

The projection algorithm must be deterministic.

# 12. Financial Goals
## 12.1 Purpose

Financial goals represent money the user intends to accumulate.

Examples:

Emergency Fund
Travel
Laptop
Wedding
Education
## 12.2 Goal Fields

A goal contains:

id
user_id
name
target_amount
current_amount
target_date
status
priority
created_at
updated_at

Optional:

linked_account_id
description
## 12.3 Goal Status

Supported statuses:

ACTIVE
ACHIEVED
PAUSED
CANCELLED
## 12.4 Goal Progress

Progress:

current_amount
÷
target_amount

Example:

Current:
Rp12M

Target:
Rp30M

Progress:
40%
## 12.5 Required Monthly Contribution

If a target date exists:

Remaining Amount
=
Target Amount
-
Current Amount

Then:

Required Contribution
=
Remaining Amount
÷
Remaining Period

The exact period calculation must be deterministic.

## 12.6 Goal Feasibility

Goal feasibility may compare:

required monthly contribution
vs
estimated free cashflow

Savio may categorize:

ON_TRACK
AT_RISK
UNLIKELY

AI may explain the result but must not calculate it independently.

## 12.7 Goal Completion

A goal may become ACHIEVED when:

current_amount >= target_amount

If current amount is not automatically linked to an account, explicit user confirmation may be required.

# 13. Cashflow Analytics
## 13.1 Income Calculation

For a selected period:

Total Income
=
sum(INCOME transactions)

Transfers are excluded.

## 13.2 Expense Calculation
Total Expense
=
sum(EXPENSE transactions)

Transfers are excluded.

## 13.3 Net Cashflow
Net Cashflow
=
Total Income
-
Total Expense
## 13.4 Savings Rate

A simple initial savings rate may be:

Net Cashflow
÷
Total Income
×
100

If total income is zero, the system must avoid division-by-zero and return an appropriate unavailable state.

## 13.5 Period Comparison

Savio should support comparison such as:

Current Month
vs
Previous Month

or:

Current Month
vs
3-Month Average
## 13.6 Category Analytics

For each expense category:

amount
percentage of total expense
change vs previous period
change vs historical baseline
# 14. Financial Baseline
## 14.1 Purpose

The financial baseline represents the user's expected future cashflow without hypothetical scenario modifications.

Baseline is required for:

forecast,
scenario comparison,
AI explanation.
## 14.2 Baseline Inputs

Possible inputs include:

current account balances
scheduled income
scheduled expenses
recurring transactions
historical variable spending
known one-time future events
## 14.3 Forecast Event Types

Forecast events must be classified.

KNOWN
SCHEDULED
ESTIMATED
ASSUMED
KNOWN

Financial event explicitly known.

Example:

Invoice already confirmed.
SCHEDULED

Generated from recurring rules.

Example:

Monthly rent.
ESTIMATED

Derived from historical behavior.

Example:

Estimated food spending.
ASSUMED

Explicit forecast assumption.

Example:

Assume monthly transport remains Rp600k.
# 15. Cashflow Forecast
## 15.1 Forecast Horizon

Initial supported horizons may include:

30 days
60 days
90 days
6 months
12 months

Very long projections should communicate increasing uncertainty.

## 15.2 Forecast Output

A forecast includes:

opening balance
projected balance timeline
projected income
projected expense
minimum projected balance
ending projected balance
cashflow risk periods
assumptions used
## 15.3 Forecast Balance

Conceptually:

Projected Balance(t)
=
Starting Balance
+ Income Events
- Expense Events

for all events occurring before time t.

## 15.4 Estimated Variable Spending

Variable spending estimation must use deterministic logic.

Possible MVP approach:

average daily discretionary spending
×
future number of days

or:

category historical average

The selected formula must be documented.

## 15.5 Forecast Confidence

Forecast confidence should not pretend to be statistically precise if the model is simple.

A practical initial classification may be:

LOW
MEDIUM
HIGH

based on:

amount of historical data,
percentage of scheduled vs estimated events,
forecast horizon.
## 15.6 Insufficient Data

If historical data is insufficient:

Forecast confidence:
LOW

Savio should explain why.

Example:

Only 18 days of transaction history are available.
# 16. Financial Scenario Simulator
## 16.1 Purpose

A scenario is a hypothetical modification applied to the baseline.

The original financial records must not be changed.

## 16.2 Scenario Lifecycle

Suggested lifecycle:

DRAFT
→ CALCULATED
→ ARCHIVED

The user may recalculate after modifications.

## 16.3 Scenario Fields

A scenario contains:

id
user_id
name
description
forecast_horizon
status
created_at
updated_at

Scenario modifications are stored separately.

## 16.4 Scenario Modification Types

Initial supported modifications:

ONE_TIME_INCOME
ONE_TIME_EXPENSE
RECURRING_INCOME
RECURRING_EXPENSE
INCOME_REDUCTION
INCOME_REMOVAL
EXPENSE_REDUCTION
SAVINGS_ADJUSTMENT
## 16.5 One-Time Expense

Example:

Buy laptop
Rp15,000,000
September 15

Scenario applies:

-Rp15M

on the specified date.

## 16.6 Income Reduction

Example:

Salary decreases 30%
starting October

The baseline salary projection is modified.

## 16.7 Income Removal

Example:

Salary stops
starting November

Projected salary events after the effective date are removed from the scenario.

## 16.8 Recurring Expense

Example:

New installment
Rp1,500,000
monthly
12 months

Scenario generates temporary recurring future expense events.

## 16.9 Multiple Modifications

A scenario may contain multiple modifications.

Example:

Resign from job
+
reduce entertainment by 50%
+
add freelance income Rp3M/month

All modifications must be combined deterministically.

## 16.10 Baseline Comparison

Each calculated scenario must compare:

BASELINE
vs
SCENARIO

Potential comparison metrics:

ending balance
minimum balance
net cashflow
cash runway
savings rate
goal impact
budget impact
## 16.11 Scenario Isolation

Scenario calculations must never modify:

real account balance,
real transactions,
real recurring rules,
real budgets,
real financial goals.

Scenario data is hypothetical.

# 17. Cash Runway
## 17.1 Purpose

Cash runway estimates how long available financial resources may support expected expenses when income is reduced or absent.

## 17.2 Basic Calculation

A simple deterministic approximation:

Liquid Balance
÷
Average Monthly Essential Expense

Example:

Liquid Balance:
Rp42M

Average Monthly Essential Expense:
Rp7M

Runway:
6 months

The exact definition of liquid assets and essential expenses must be documented.

## 17.3 Runway Limitations

Runway is an estimate, not a guarantee.

It must clearly state assumptions.

# 18. Financial Health
## 18.1 Purpose

Financial health may summarize multiple deterministic indicators.

Financial health is not a credit score.

## 18.2 Potential Components

Example:

Savings Rate
Cashflow Stability
Budget Adherence
Emergency Buffer
Expense Volatility
## 18.3 Explainability

The score must expose contributing factors.

Example:

Financial Health:
74 / 100

Positive:
+ Stable recurring income
+ Savings rate above threshold

Negative:
- High discretionary volatility
- Emergency buffer below target

AI may explain factors but cannot invent the score.

## 18.4 MVP Status

Financial health may be P1 rather than P0 if necessary to protect delivery scope.

# 19. AI Insights
## 19.1 Purpose

AI insights convert deterministic financial signals into understandable explanations.

AI must operate on structured context.

## 19.2 Insight Types

Initial types may include:

SPENDING_ANOMALY
INCOME_CHANGE
BUDGET_RISK
CASHFLOW_RISK
GOAL_RISK
RECURRING_COST
SAVINGS_PATTERN
POSITIVE_TREND
## 19.3 Insight Source

Every AI insight should reference the deterministic data that triggered it.

Example:

SPENDING_ANOMALY

Current dining:
Rp2.4M

3-month average:
Rp1.5M

Difference:
+60%

AI may explain why this is notable.

## 19.4 Insight Fields

An AI insight may contain:

id
user_id
type
severity
title
summary
structured_context
status
model
generated_at
created_at

Optional:

confidence
drivers
recommended_actions
## 19.5 Insight Status

Suggested status:

NEW
VIEWED
DISMISSED
ACKNOWLEDGED

AI insights are not financial transactions.

## 19.6 Insight Severity

Suggested severity:

INFO
LOW
MEDIUM
HIGH

Severity should be based primarily on deterministic triggers.

AI should not freely assign criticality without constraints.

## 19.7 Insight Deduplication

The system should avoid repeatedly generating the same insight.

Example:

Dining budget risk

should not generate identical notifications every background run.

A deduplication key may incorporate:

user
insight type
reference period
affected category/resource
# 20. AI Transaction Categorization
## 20.1 Purpose

AI may suggest a category based on transaction description.

Example:

GRAB*FOOD
→ Food & Dining
## 20.2 AI Suggestion

AI categorization returns:

suggested_category
confidence
reason
## 20.3 Human Control

The suggestion must not silently become authoritative unless the product explicitly supports user-approved automatic rules.

For MVP:

AI suggests
→ user confirms

is preferred.

## 20.4 Learning Preference

When a user repeatedly changes:

Merchant X
→ Category Y

future versions may use that history as preference context.

This should not require fine-tuning.

# 21. AI Financial Copilot
## 21.1 Purpose

The AI Copilot provides natural-language interaction with Savio's financial intelligence.

## 21.2 Supported Question Categories

Examples:

Spending
Why did I spend more this month?
Recurring Cost
What are my largest recurring expenses?
Budget
Which budgets are at risk?
Cashflow
Will my balance become low before payday?
Scenario
Can I simulate buying a Rp10M laptop?
Goals
Am I on track for my emergency fund?
## 21.3 AI Context

The model should receive only relevant structured context.

Example:

User question
+
requested period
+
analytics summary
+
forecast result
+
relevant categories

Avoid sending the entire database unnecessarily.

## 21.4 Tool-Based AI

The AI Copilot should prefer calling deterministic application tools/functions.

Conceptual tools:

get_cashflow_summary
get_category_breakdown
get_budget_status
get_goal_status
get_forecast
compare_periods
calculate_scenario
get_recurring_expenses

The AI uses these tools to obtain authoritative numbers.

## 21.5 Copilot Write Actions

Initial AI Copilot should preferably be read-oriented.

If future write actions are supported:

AI proposes
→ user confirms
→ backend validates
→ action executes

Example:

AI:
Would you like me to create a Rp1.5M monthly food budget?

User confirms.

Backend validates and creates budget.

No silent write.

# 22. AI Safety & Reliability
## 22.1 No Direct Database Authority

The LLM must not have unrestricted database mutation access.

## 22.2 No Authoritative Calculation

The LLM must not independently calculate:

balances
financial health
budget utilization
forecast baseline
goal progress
scenario outcome
## 22.3 Structured Output Validation

AI structured output must be validated by backend schema before use.

Malformed AI output must not break business state.

## 22.4 AI Failure

If the AI provider is unavailable:

financial core remains available

Users should still be able to:

create transactions,
view balances,
manage budgets,
view deterministic analytics,
calculate forecast,
run scenario simulation.

AI-dependent features may show degraded state.

## 22.5 AI Timeout

AI requests must have bounded timeout.

The UI should not remain indefinitely loading.

## 22.6 AI Auditability

Important AI requests should record:

feature
model
status
latency
token usage if available
generated_at

Sensitive raw financial data should not be unnecessarily duplicated in logs.

# 23. Notifications
## 23.1 Notification Types

Potential notifications include:

BUDGET_WARNING
BUDGET_EXCEEDED
UPCOMING_BILL
LOW_PROJECTED_BALANCE
GOAL_AT_RISK
RECURRING_TRANSACTION_DUE
AI_INSIGHT_AVAILABLE
## 23.2 Notification Status
UNREAD
READ

Optional:

ARCHIVED
## 23.3 Notification Deduplication

Repeated jobs must not create duplicate notifications for the same event.

# 24. Background Jobs

Potential background jobs include:

Recurring transaction processing
Upcoming bill detection
Budget risk calculation
Forecast snapshot generation
AI insight generation
Notification generation
Expired session cleanup

All jobs must be idempotent where possible.

# 25. Idempotency
## 25.1 Recurring Generation

A recurring occurrence must not be generated twice.

## 25.2 Background AI Insight

Repeated job execution must not create identical insight records unnecessarily.

## 25.3 Financial Write Operations

Where duplicate client retries are possible, important financial commands may support idempotency keys.

Example:

POST /transactions

Idempotency-Key:
abc-123

This is especially valuable if network retries could duplicate financial entries.

# 26. Optimistic Concurrency
## 26.1 Purpose

Concurrent editing must not silently overwrite changes.

Important mutable entities may include a:

version

field.

## 26.2 Example

Current budget:

version = 5
amount = Rp2M

User A changes:

Rp2.5M

Result:

version = 6

User B submits old version 5.

Result:

# 409 VERSION_CONFLICT
## 26.3 Candidate Resources

Optimistic locking may be used for:

accounts
budgets
goals
recurring rules
scenarios
user settings

Transaction updates may also use locking depending on design.

# 27. Audit Trail
## 27.1 Purpose

Important financial and security changes should be traceable.

## 27.2 Audit Events

Examples:

USER_LOGIN
USER_LOGOUT

ACCOUNT_CREATED
ACCOUNT_ARCHIVED

TRANSACTION_CREATED
TRANSACTION_UPDATED
TRANSACTION_VOIDED

TRANSFER_CREATED
TRANSFER_VOIDED

BUDGET_CREATED
BUDGET_UPDATED

GOAL_UPDATED

RECURRING_CREATED
RECURRING_PAUSED

SCENARIO_CALCULATED

AI_INSIGHT_GENERATED
AI_INSIGHT_DISMISSED
## 27.3 Audit Content

Audit records may contain:

id
user_id
action
entity_type
entity_id
metadata
request_id
created_at

Sensitive credentials must never be included.

# 28. Validation

All input must be validated by the backend.

Frontend validation improves UX but is not authoritative.

## 28.1 Common Validation

Examples:

UUID format
required field
string length
enum
numeric range
date
email
currency
pagination
sort field
## 28.2 Financial Amount Validation

General rules:

amount > 0
maximum amount limit
supported decimal precision

Reasonable technical maximums should prevent overflow or abuse.

## 28.3 Date Validation

Date ranges must be logical.

Examples:

goal target date
> creation/current date where required

recurring end date
>= start date

report end
>= report start
# 29. Search, Filter, Sort & Pagination

Large resource listings must use backend-driven filtering.

## 29.1 Transactions

Search/filter may include:

search
type
account
category
date_from
date_to
min_amount
max_amount
sort
page
limit
## 29.2 Accounts

Filters:

type
status
sort
## 29.3 Budgets

Filters:

status
period
category
## 29.4 Goals

Filters:

status
priority
target_date
## 29.5 AI Insights

Filters:

type
severity
status
date
## 29.6 Pagination

Default:

page = 1
limit = 20

Maximum limit should be bounded.

Example:

max = 100
# 30. Error Semantics

Savio must return consistent business errors.

Examples:

AUTHENTICATION_REQUIRED
PERMISSION_DENIED
VALIDATION_ERROR
RESOURCE_NOT_FOUND
RESOURCE_CONFLICT
VERSION_CONFLICT

ACCOUNT_ARCHIVED
ACCOUNT_HAS_FINANCIAL_HISTORY

INVALID_CATEGORY_TYPE

TRANSFER_SAME_ACCOUNT

INVALID_RECURRING_DATE

DUPLICATE_BUDGET

INSUFFICIENT_FORECAST_DATA

INVALID_SCENARIO_MODIFICATION

AI_PROVIDER_UNAVAILABLE
AI_OUTPUT_INVALID

RATE_LIMIT_EXCEEDED
INTERNAL_SERVER_ERROR
# 31. HTTP Status Mapping

Recommended mapping:

# 200 OK
→ successful read/update/action

# 201 Created
→ resource created

# 204 No Content
→ successful delete/logout where applicable

# 400 Bad Request
→ malformed request

# 401 Unauthorized
→ authentication required/invalid

# 403 Forbidden
→ authenticated but not allowed

# 404 Not Found
→ resource unavailable

# 409 Conflict
→ duplicate/conflicting state/version

# 422 Unprocessable Entity
→ business/validation failure

# 429 Too Many Requests
→ rate limit

# 500 Internal Server Error
→ unexpected server failure
# 32. Business Transactions

Operations that modify multiple related financial records must use database transactions.

Examples:

Transfer
Transaction update affecting account balance
Transaction deletion
Adjustment
Recurring auto-post
Account reconciliation
# 33. Data Integrity

Foreign key constraints should protect relationships.

Examples:

transaction.account_id
→ accounts.id

transaction.category_id
→ categories.id

budget.category_id
→ categories.id

goal.user_id
→ users.id

Application-level validation must complement database constraints.

# 34. Reporting

Reports are derived views of financial records.

Reports must not create duplicate financial state.

Initial reports:

Income vs Expense
Category Breakdown
Savings Rate
Budget Performance
Goal Progress
Cashflow Trend
# 35. Dashboard

The dashboard should answer:

How much money do I currently have?

How much came in this period?

How much went out?

Am I spending more than usual?

Are any budgets at risk?

What bills are coming?

What does my cashflow look like next?

Is there any important AI insight?

Potential dashboard sections:

Net Worth / Total Balance
Income
Expense
Net Cashflow
Savings Rate
Cashflow Chart
Budget Summary
Goal Summary
Upcoming Recurring Events
AI Insight Cards
Forecast Preview
# 36. User Settings

User-configurable settings may include:

timezone
default currency
locale
date format
number format
AI preference
notification preference
# 37. AI Preference

Users may be allowed to:

enable AI insights
disable AI insights

Core deterministic functionality must remain available regardless.

# 38. Data Export

Post-MVP users may export:

transactions
reports
account history

Formats:

CSV

PDF may be added later.

# 39. Data Import

Post-MVP transaction import should use a staged workflow:

UPLOAD
→ PARSE
→ VALIDATE
→ REVIEW
→ IMPORT

Invalid rows must not silently enter financial records.

# 40. Receipt Processing

Future receipt processing:

UPLOAD
→ AI/OCR EXTRACTION
→ STRUCTURED DRAFT
→ USER REVIEW
→ CREATE TRANSACTION

The extracted result must never immediately become an authoritative transaction without validation.

# 41. Security Requirements

Financial information must be treated as sensitive.

Minimum requirements:

Secure cookie authentication
CSRF protection
Password hashing
Session revocation
Refresh rotation
Rate limiting
Backend authorization
Request validation
Sensitive log filtering
Secure headers
CORS restriction
No hardcoded secrets

Detailed security design belongs in:

docs/engineering/security.md
# 42. Data Visibility Rules

Every domain query must scope data by authenticated ownership.

Example:

Bad:

SELECT * FROM transactions WHERE id = ?

Preferred logical behavior:

SELECT *
FROM transactions
WHERE id = ?
AND user_id = ?

or equivalent repository policy.

# 43. AI Data Minimization

AI context should contain only information required for the current task.

Example:

Question:

Why did my food spending increase?

The system may provide:

Food transactions
category summary
historical comparison
relevant recurring expenses

It should not automatically send:

all financial history
all profile data
all unrelated transactions
# 44. AI Provider Failure Behavior

When AI is unavailable:

AI Insight
→ temporarily unavailable

AI Copilot
→ temporarily unavailable

But:

Transactions
Accounts
Budgets
Goals
Forecast
Scenario
Analytics

must continue operating.

# 45. Rate Limiting

Sensitive endpoints should have appropriate rate limits.

Examples:

login
register
refresh
AI Copilot
AI insight generation

AI endpoints require stronger limits because they consume external resources.

# 46. Financial Calculation Versioning

As calculation algorithms evolve, Savio should be capable of identifying which calculation version produced important generated results.

Potential fields:

calculation_version
forecast_version

This is especially useful for stored forecast or scenario snapshots.

# 47. Scenario Snapshot

When a user calculates a scenario, Savio may persist a snapshot of:

baseline inputs
scenario modifications
calculated outputs
calculation version
calculated_at

This prevents historical scenario results from becoming ambiguous when underlying data later changes.

# 48. Forecast Freshness

A forecast result should expose when it was generated.

Example:

Generated:
# 24 Aug 2026 21:00

Based on data through:
# 24 Aug 2026

If financial data changes, an old forecast may be marked:

STALE

and require recalculation.

# 49. Scenario Freshness

A scenario becomes stale if its baseline financial data changes after scenario calculation.

Examples:

new transaction
budget change
recurring rule change
account balance adjustment

The UI should indicate:

Scenario results may be outdated.
Recalculate.
# 50. Financial Event Ordering

When multiple financial events occur on the same date, the system must define deterministic ordering if the exact order affects projected balance.

Possible rule:

Explicit scheduled events
→ recurring events
→ estimated daily spending

or use event timestamps where available.

The selected rule must remain consistent.

# 51. Negative Balance

Savio may allow account balances to become negative if this reflects reality.

Example:

credit-like balance
overdraft
manual tracking mismatch

However, the system should surface negative balances as a financial signal.

It must not reject all transactions solely because they would create a negative balance unless the account type explicitly requires that rule.

# 52. Essential vs Discretionary Spending

To support runway and richer analytics, expense categories may optionally include:

ESSENTIAL
DISCRETIONARY

Examples:

Rent
→ ESSENTIAL

Entertainment
→ DISCRETIONARY

Users should be able to override classification.

This classification may be P1 if MVP scope must remain smaller.

# 53. Transaction Tags

Transactions may support tags for flexible grouping.

Examples:

vacation
work
family
reimbursable

Tags do not replace financial categories.

# 54. Merchant Normalization

Future intelligence may normalize merchant descriptions.

Example:

GRAB*FOOD
GRAB FOOD JAKARTA
GRAB-FOOD

→

GrabFood

AI may assist, but user correction remains authoritative.

# 55. Business Rule Priority

When rules conflict, priority is:

Security & ownership
        ↓
Financial integrity
        ↓
Deterministic calculation
        ↓
Explicit user configuration
        ↓
AI recommendation

AI must never override higher-level rules.

# 56. MVP Business Flow

The primary MVP journey should demonstrate:

Register
    ↓
Login
    ↓
Create Accounts
    ↓
Create Income / Expense Transactions
    ↓
Create Recurring Income / Expenses
    ↓
Create Budget
    ↓
Create Financial Goal
    ↓
View Dashboard Analytics
    ↓
View Cashflow Forecast
    ↓
Receive Explainable AI Insight
    ↓
Create Financial Scenario
    ↓
Compare Baseline vs Scenario
    ↓
Ask AI Copilot
    ↓
User Makes Decision
# 57. Example End-to-End Scenario

Example user financial state:

Bank Balance:
Rp20,000,000

Monthly Salary:
Rp12,000,000

Monthly Expenses:
Rp8,000,000

Emergency Fund Goal:
Rp30,000,000

The user asks:

What happens if I buy a Rp15,000,000 laptop this month?

Savio creates a scenario.

Baseline:

Month-end balance:
Rp24,000,000

Savings rate:
33%

Emergency buffer:
4.2 months

Scenario:

Month-end balance:
Rp9,000,000

Savings rate:
8%

Emergency buffer:
1.9 months

Savio then provides deterministic comparison.

AI may explain:

The purchase does not create an immediate negative balance,
but it materially reduces your emergency buffer and delays
progress toward your emergency fund goal.

The user remains responsible for the final decision.

# 58. Product Boundary

Features should be rejected or postponed if they turn Savio into:

banking software
trading software
investment prediction software
tax automation
accounting ERP
credit underwriting
loan marketplace

Savio remains focused on:

personal cashflow understanding
financial planning
forecasting
scenario simulation
AI-assisted explanation
# 59. Definition of Correctness

A Savio feature is not complete merely because the UI works.

A financial feature is considered complete only if:

business rules defined
backend validation enforced
database integrity preserved
authorization enforced
financial calculations tested
error states handled
concurrency considered
UI states handled
audit behavior defined where relevant

AI-enabled features additionally require:

deterministic source data
structured context
structured output validation
failure handling
human control
# 60. Final Business Rule

The central rule governing the entire Savio product is:

Financial records and calculations are authoritative only when produced by deterministic application logic. AI may interpret those results, but it may never replace them.

The responsibility chain is:

USER DATA
    ↓
DETERMINISTIC FINANCE ENGINE
    ↓
FINANCIAL RESULT
    ↓
AI INTERPRETATION
    ↓
USER DECISION

Or, in its shortest form:

Finance Engine calculates. AI interprets. User decides.