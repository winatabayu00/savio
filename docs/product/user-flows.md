# Savio — User Flows

## Related Documents

- [README.md](../../README.md) — project overview, setup, and documentation index.
- [Product Foundation](product-foundation.md) — product vision and positioning.
- [Business Requirements](business-requirements.md) — business rules each flow must satisfy.
- [Design System](../../DESIGN.md) — visual/UX conventions applied to these flows.

## 1. Document Purpose

This document defines the primary user flows for Savio.

The purpose of this document is to translate the product foundation and business requirements into concrete user journeys that can later be implemented through:

- frontend routes,
- pages,
- forms,
- backend APIs,
- state transitions,
- background jobs,
- deterministic financial calculations,
- and AI-assisted interactions.

This document focuses on:

- what the user is trying to accomplish,
- what steps the user takes,
- what the system does,
- what business rules apply,
- what errors may occur,
- what financial calculations happen,
- and where AI participates.

The core Savio principle remains:

> **Finance Engine calculates. AI interprets. User decides.**

---

# 2. User Flow Principles

Every Savio flow should follow several product principles.

## 2.1 Deterministic Before AI

Whenever a flow contains financial calculation:

```text
User Action
    ↓
Backend Validation
    ↓
Deterministic Finance Engine
    ↓
Financial Result
    ↓
Optional AI Interpretation
    ↓
User Decision
```

AI must not replace deterministic financial calculations.

---

## 2.2 Backend Is the Source of Truth

Frontend validation exists for user experience.

Backend validation remains authoritative.

Example:

```text
Frontend:
amount > 0

Backend:
amount > 0

Database:
appropriate constraint where possible
```

A malicious or outdated frontend must not be able to bypass business rules.

---

## 2.3 No Silent Financial Changes

Any action that changes authoritative financial records should be explicit.

Examples:

```text
Create Transaction
Update Draft Transaction
Void Transaction + Create Replacement
Create Transfer
Create Adjustment
```

AI may suggest these actions but must not silently execute them.

---

## 2.4 Explain Important Results

Important calculations should show not only the result but also the underlying reasoning.

Example:

Instead of:

```text
Budget Risk: HIGH
```

show:

```text
Budget Risk: HIGH

Budget:
Rp2,000,000

Spent:
Rp1,650,000

Current day:
18 / 30

Projected spending:
Rp2,750,000

Projected overspend:
Rp750,000
```

AI may translate this information into natural language.

---

## 2.5 Graceful AI Failure

If AI is unavailable:

```text
Transactions          ✓
Accounts              ✓
Budgets               ✓
Goals                 ✓
Analytics             ✓
Forecast              ✓
Scenario Simulator    ✓

AI Insights           unavailable
AI Copilot            unavailable
```

Core Savio functionality must continue to operate.

---

# 3. Primary User Journey

The overall Savio user journey is:

```text
Visitor
   ↓
Register
   ↓
Authenticated User
   ↓
Initial Financial Setup
   ↓
Create Financial Accounts
   ↓
Add Transactions
   ↓
Configure Recurring Transactions
   ↓
Create Budgets
   ↓
Create Financial Goals
   ↓
View Dashboard
   ↓
Understand Financial Patterns
   ↓
View Cashflow Forecast
   ↓
Receive AI Insights
   ↓
Create Financial Scenario
   ↓
Compare Baseline vs Scenario
   ↓
Ask AI Copilot
   ↓
User Makes Financial Decision
```

---

# 4. Authentication Flow

## 4.1 Registration Flow

### Goal

Allow a new user to create a Savio account securely.

### Flow

```text
Visitor
   ↓
Open Register Page
   ↓
Enter:
- Name
- Email
- Password
- Password Confirmation
   ↓
Frontend Validation
   ↓
POST Registration
   ↓
Backend Validation
   ↓
Email Unique?
   ├── No
   │   ↓
   │  Show Validation Error
   │
   └── Yes
       ↓
Password Valid?
   ├── No
   │   ↓
   │  Show Validation Error
   │
   └── Yes
       ↓
Create User
       ↓
Create Authentication Session
       ↓
Set Secure Cookies
       ↓
Create / Issue CSRF Token
       ↓
Redirect to Onboarding
```

### Success Result

The user becomes authenticated and enters the initial setup flow.

### Possible Errors

```text
EMAIL_ALREADY_EXISTS
INVALID_EMAIL
WEAK_PASSWORD
PASSWORD_CONFIRMATION_MISMATCH
VALIDATION_ERROR
RATE_LIMIT_EXCEEDED
INTERNAL_SERVER_ERROR
```

---

## 4.2 Login Flow

### Goal

Authenticate an existing user.

### Flow

```text
Visitor
   ↓
Open Login
   ↓
Enter Email + Password
   ↓
Submit
   ↓
Backend Authentication
   ↓
Credentials Valid?
   ├── No
   │   ↓
   │  Show Generic Authentication Error
   │
   └── Yes
       ↓
User ACTIVE?
   ├── No
   │   ↓
   │  Authentication Rejected
   │
   └── Yes
       ↓
Create Session
       ↓
Set Access Cookie
       ↓
Set Refresh Cookie
       ↓
Set / Refresh CSRF Token
       ↓
Redirect Dashboard
```

### Security Requirements

The UI must not distinguish unnecessarily between:

```text
Email does not exist
```

and:

```text
Password incorrect
```

A generic response is preferred.

Example:

```text
Invalid email or password.
```

---

## 4.3 Authenticated Application Bootstrap

When Savio loads:

```text
Application Starts
   ↓
GET Current User
   ↓
Access Cookie Valid?
   ├── Yes
   │   ↓
   │  Load User
   │   ↓
   │  Enter Application
   │
   └── No
       ↓
Attempt Refresh
       ↓
Refresh Valid?
   ├── Yes
   │   ↓
   │  Rotate Refresh Session
   │   ↓
   │  Retry Current User
   │
   └── No
       ↓
Clear Authentication State
       ↓
Redirect Login
```

---

## 4.4 Logout Flow

```text
Authenticated User
   ↓
Click Logout
   ↓
POST Logout
   ↓
Backend Revokes Session
   ↓
Authentication Cookies Cleared
   ↓
Frontend Clears Auth State
   ↓
Redirect Login
```

---

# 5. First-Time Onboarding

## 5.1 Goal

Help new users create enough financial context for Savio to become useful.

The onboarding flow should remain short.

Recommended initial flow:

```text
Welcome
   ↓
Profile Preferences
   ↓
Create First Account
   ↓
Optional Initial Balance
   ↓
Optional Recurring Income
   ↓
Finish
   ↓
Dashboard
```

---

## 5.2 Welcome

The user sees:

```text
Welcome to Savio

Understand your money today.
See what comes next.
Test decisions before making them.
```

The product should explain briefly:

```text
Track
Understand
Forecast
Simulate
```

---

## 5.3 Financial Preferences

Collect:

```text
Timezone
Default Currency
```

Initial currency:

```text
IDR
```

Example:

```text
Timezone:
Asia/Jakarta

Currency:
IDR
```

---

## 5.4 First Account

The user creates at least one account.

Example:

```text
Account Name:
BCA

Type:
BANK

Initial Balance:
Rp8,500,000
```

The user may skip some optional onboarding steps, but Savio should clearly communicate that analytics quality improves when financial data is complete.

---

## 5.5 Initial Recurring Income

Optional step:

```text
Do you receive regular income?
```

Example:

```text
Salary

Amount:
Rp12,000,000

Frequency:
Monthly

Date:
25th
```

This immediately improves forecast usefulness.

---

## 5.6 Onboarding Completion

After completion:

```text
Onboarding
   ↓
Create Required Resources
   ↓
Dashboard
```

The dashboard should show appropriate empty states when transaction history is still limited.

---

# 6. Dashboard Flow

## 6.1 Goal

Allow users to understand their current financial position quickly.

The dashboard should answer:

```text
How much money do I currently have?

How much money came in this month?

How much went out?

What is my net cashflow?

Am I spending unusually?

Are my budgets healthy?

What bills are coming?

What does my future balance look like?

Is there anything important Savio wants me to notice?
```

---

## 6.2 Dashboard Load Flow

```text
User Opens Dashboard
   ↓
Request Dashboard Composite Data
   ↓
Backend Loads:
- Account summary
- Current period analytics
- Budget summary
- Goal summary
- Upcoming recurring events
- Forecast preview
- Recent AI insights
   ↓
Return Aggregated Response
   ↓
Render Dashboard
```

A composite API may be preferable to many independent calls if it improves consistency and performance.

---

## 6.3 Dashboard Empty State

New user:

```text
No financial activity yet.
```

Suggested actions:

```text
+ Add Transaction
+ Create Recurring Income
+ Create Budget
```

The user should not see meaningless charts filled with zero data.

---

# 7. Account Management Flow

## 7.1 Create Account

```text
Accounts
   ↓
Add Account
   ↓
Enter:
- Name
- Type
- Initial Balance
- Currency
   ↓
Validate
   ↓
Create Account
   ↓
Success
```

Example:

```text
Name:
BCA Main

Type:
BANK

Initial Balance:
Rp10,000,000
```

---

## 7.2 Account Detail

The account detail page may show:

```text
Account Balance

Income
Expense
Transfers

Transaction History

Balance Trend
```

---

## 7.3 Edit Account

Editable fields may include:

```text
name
type
description
institution
```

Changing `current_balance` directly should not be normal account editing behavior.

If actual money differs from Savio's tracked value, use reconciliation / adjustment.

---

## 7.4 Archive Account

```text
Account Detail
   ↓
Archive
   ↓
Confirmation
   ↓
Has Financial History?
   ↓
Archive Account
```

The account remains available historically.

---

## 7.5 Delete Empty Account

```text
Delete Account
   ↓
Check Dependencies
   ↓
Transactions = 0?
Transfers = 0?
   ├── No
   │   ↓
   │  Reject
   │  ACCOUNT_HAS_FINANCIAL_HISTORY
   │
   └── Yes
       ↓
Delete
```

---

# 8. Income Transaction Flow

## 8.1 Goal

Record money entering an account.

### Flow

```text
Transactions
   ↓
Add Transaction
   ↓
Choose INCOME
   ↓
Enter:
- Account
- Category
- Amount
- Date
- Description
- Optional Notes
   ↓
Frontend Validation
   ↓
Submit
   ↓
Backend Validation
   ↓
Account Belongs to User?
   ↓
Category Type = INCOME?
   ↓
Amount > 0?
   ↓
Database Transaction
   ↓
Create Transaction
   ↓
Update / Recalculate Account Balance
   ↓
Commit
   ↓
Refresh Analytics
```

---

## 8.2 Example

```text
Account:
BCA

Category:
Salary

Amount:
Rp12,000,000

Date:
25 Aug 2026

Description:
August Salary
```

Balance:

```text
Before:
Rp4,000,000

Income:
+Rp12,000,000

After:
Rp16,000,000
```

---

# 9. Expense Transaction Flow

## 9.1 Goal

Record money leaving an account.

### Flow

```text
Add Transaction
   ↓
Choose EXPENSE
   ↓
Enter:
- Account
- Expense Category
- Amount
- Date
- Description
   ↓
Validate
   ↓
Create Transaction
   ↓
Apply Account Balance Effect
   ↓
Refresh:
- Dashboard
- Budget Usage
- Analytics
- Forecast Freshness
```

---

## 9.2 Example

```text
Account:
GoPay

Category:
Food & Dining

Amount:
Rp85,000

Merchant:
GrabFood
```

---

# 10. AI-Assisted Transaction Categorization Flow

## 10.1 Goal

Help users select transaction categories faster.

Example input:

```text
GRAB*FOOD 83219
Rp87,500
```

Flow:

```text
User Enters Description
   ↓
Request AI Category Suggestion
   ↓
Backend Builds Minimal Context
   ↓
AI Returns Structured Suggestion
   ↓
Schema Validation
   ↓
Frontend Shows:
Food & Dining
Confidence 96%
   ↓
User:
Accept / Choose Different Category
```

---

## 10.2 Important Rule

AI suggestion is not authoritative.

```text
AI Suggestion
   ↓
User Confirmation
   ↓
Transaction Creation
```

---

## 10.3 AI Failure

```text
AI unavailable
```

The category dropdown remains fully usable manually.

---

# 11. Transfer Flow

## 11.1 Goal

Move money between user-owned accounts without counting the transfer as income or expense.

### Flow

```text
Transactions
   ↓
Transfer Money
   ↓
Select:
- Source Account
- Destination Account
- Amount
- Date
   ↓
Validate
   ↓
source != destination
   ↓
Both accounts belong to user
   ↓
Both accounts active
   ↓
Database Transaction Begins
   ↓
Source Balance - Amount
Destination Balance + Amount
Create Transfer Record
Create Linked Financial Entries if used
   ↓
Commit
```

---

## 11.2 Example

```text
BCA
Rp5,000,000

Transfer:
Rp1,000,000

GoPay
Rp300,000
```

After:

```text
BCA
Rp4,000,000

GoPay
Rp1,300,000
```

Analytics:

```text
Income:
unchanged

Expense:
unchanged
```

---

# 12. Edit Transaction Flow

## 12.1 Goal

Edit a still-pending DRAFT transaction before it becomes financially effective.

Posted transactions are financially immutable; correcting one is done by voiding the original and creating a replacement.

Example (DRAFT correction):

Original:

```text
Food Expense
Rp100,000
```

Correction:

```text
Rp120,000
```

### Flow

```text
Open Transaction
   ↓
Edit
   ↓
Submit New Data + Version
   ↓
Backend Loads Existing Transaction
   ↓
Ownership Check
   ↓
Status Is DRAFT?
   ├── No
   │   ↓
   │  409 TRANSACTION_IMMUTABLE
   │
   └── Yes
       ↓
Version Matches?
   ├── No
   │   ↓
   │  409 VERSION_CONFLICT
   │
   └── Yes
       ↓
Apply New Financial Effect Atomically
       ↓
Update Transaction
       ↓
Commit
```

---

## 12.2 Example

Old:

```text
Expense:
Rp100,000
```

New:

```text
Expense:
Rp120,000
```

Balance delta:

```text
Additional impact:
-Rp20,000
```

---

# 13. Void Transaction Flow

## 13.1 Goal

Invalidate an incorrect posted transaction without corrupting account state and without losing history.

Posted transactions are never hard-deleted or destructively rewritten.

### Flow

```text
Transaction Detail
   ↓
Void
   ↓
Confirmation Dialog + Reason
   ↓
Backend Loads Transaction
   ↓
Ownership Validation
   ↓
Status Is POSTED?
   ├── No
   │   ↓
   │  409 ALREADY_VOIDED
   │
   └── Yes
       ↓
Reverse Original Financial Effect Atomically
       ↓
Mark Record VOIDED
       ↓
Commit
   ↓
(Optional) Create Replacement Transaction
```

---

## 13.2 Example

Voiding:

```text
Expense:
Rp150,000
```

results in:

```text
Account:
+Rp150,000
```

and the record remains visible as VOIDED.

---

# 14. Account Reconciliation Flow

## 14.1 Goal

Correct Savio when the tracked balance differs from the real account balance.

Example:

```text
Savio Balance:
Rp4,800,000

Actual Bank Balance:
Rp5,000,000
```

Difference:

```text
+Rp200,000
```

### Flow

```text
Account Detail
   ↓
Reconcile Balance
   ↓
Enter Actual Balance
   ↓
Backend Calculates Difference
   ↓
Show:
Current
Actual
Adjustment
   ↓
Require Reason
   ↓
User Confirms
   ↓
Create ADJUSTMENT
   ↓
Update Balance
   ↓
Audit Event
```

---

# 15. Recurring Income Flow

## 15.1 Goal

Represent predictable recurring income.

Example:

```text
Salary
Rp12M
Monthly
25th
```

### Flow

```text
Recurring
   ↓
Add Recurring Transaction
   ↓
Choose INCOME
   ↓
Enter:
- Account
- Category
- Amount
- Frequency
- Start Date
- Optional End Date
   ↓
Validate
   ↓
Create Recurring Rule
   ↓
Calculate Next Occurrence
   ↓
Include in Forecast
```

---

# 16. Recurring Expense Flow

Example:

```text
Netflix
Rp186,000
Monthly
10th
```

Flow:

```text
Create Recurring Expense
   ↓
Store Schedule
   ↓
Calculate Next Occurrence
   ↓
Upcoming Bills
   ↓
Forecast
   ↓
Optional Reminder / Auto Post
```

---

# 17. Recurring Auto-Post Flow

If auto-post is enabled:

```text
Background Worker
   ↓
Find Due Recurring Rules
   ↓
Generate Occurrence
   ↓
Check Idempotency
   ↓
Already Generated?
   ├── Yes
   │   ↓
   │  Skip
   │
   └── No
       ↓
Database Transaction
       ↓
Create Financial Transaction
       ↓
Update Account Balance
       ↓
Advance Next Occurrence
       ↓
Commit
       ↓
Generate Notification if needed
```

Unique occurrence behavior should prevent:

```text
Salary posted twice
```

after retry.

---

# 18. Pause Recurring Transaction Flow

```text
Recurring Detail
   ↓
Pause
   ↓
Status:
ACTIVE → PAUSED
   ↓
No future automatic postings
   ↓
Forecast excludes future occurrences
```

Existing posted transactions remain untouched.

---

# 19. Resume Recurring Transaction Flow

```text
PAUSED
   ↓
Resume
   ↓
Recalculate Next Valid Occurrence
   ↓
ACTIVE
```

The system must avoid generating historical duplicates.

---

# 20. Budget Creation Flow

## 20.1 Goal

Create a spending target for a category.

Example:

```text
Food & Dining

Monthly Budget:
Rp2,000,000
```

### Flow

```text
Budgets
   ↓
Create Budget
   ↓
Choose Expense Category
   ↓
Enter Amount
   ↓
Select Period
   ↓
Validate
   ↓
Conflicting Active Budget?
   ├── Yes
   │   ↓
   │  Reject / Ask to Edit Existing
   │
   └── No
       ↓
Create Budget
       ↓
Calculate Current Usage
       ↓
Show Status
```

---

# 21. Budget Monitoring Flow

When user opens a budget:

```text
Budget
Rp2,000,000

Spent
Rp1,450,000

Remaining
Rp550,000

Utilization
72.5%
```

System computes status.

Example:

```text
0–79%
ON_TRACK

80–99%
WARNING

>=100%
EXCEEDED
```

---

# 22. Budget Risk Forecast Flow

Example:

```text
Budget:
Rp2M

Spent:
Rp1.4M

Day:
15 / 30
```

Flow:

```text
Budget Data
   ↓
Finance Engine
   ↓
Historical Spending Pace
   ↓
Projected End-of-Period Spend
   ↓
Compare to Budget
```

Output:

```text
Projected:
Rp2.7M

Risk:
EXCEEDED

Projected Overspend:
Rp700k
```

AI may explain:

```text
Your food budget is likely to exceed its monthly limit because
spending during the first half of the month is already 70% of
the available budget.
```

AI does not calculate the `Rp2.7M`.

---

# 23. Financial Goal Creation Flow

## 23.1 Goal

Allow users to plan toward a financial target.

Example:

```text
Emergency Fund
Rp30,000,000
Target: 31 Dec 2026
```

### Flow

```text
Goals
   ↓
Create Goal
   ↓
Enter:
- Name
- Target Amount
- Current Amount
- Target Date
- Priority
   ↓
Validate
   ↓
Finance Engine Calculates:
- Progress
- Remaining Amount
- Required Contribution
   ↓
Create
```

---

# 24. Goal Detail Flow

Example:

```text
Emergency Fund

Target:
Rp30M

Current:
Rp12M

Progress:
40%

Remaining:
Rp18M

Target:
6 months

Required Contribution:
Rp3M/month
```

---

# 25. Goal Feasibility Flow

```text
Goal
   ↓
Required Monthly Contribution
   ↓
Estimated Free Cashflow
   ↓
Compare
```

Possible result:

```text
ON_TRACK
AT_RISK
UNLIKELY
```

Example:

```text
Required:
Rp3,000,000/month

Estimated free cashflow:
Rp2,100,000/month

Status:
AT_RISK
```

AI may explain the gap.

---

# 26. Analytics Flow

## 26.1 Monthly Summary

```text
User Opens Analytics
   ↓
Select Period
   ↓
Backend Calculates:
- Income
- Expense
- Net Cashflow
- Savings Rate
- Category Distribution
- Comparison
   ↓
Render Charts + Metrics
```

---

## 26.2 Period Comparison

Example:

```text
August Expense:
Rp8.4M

July Expense:
Rp6.9M
```

Difference:

```text
+Rp1.5M
+21.7%
```

The system may identify which categories contributed most.

---

# 27. "Where Did My Money Go?" Flow

This is one of Savio's primary intelligence experiences.

### Flow

```text
User Opens Monthly Analysis
   ↓
Finance Engine Calculates:
- Category totals
- Previous period
- Historical baseline
- Largest transactions
- Unusual category changes
   ↓
AI Context Builder
   ↓
AI Explanation
```

Example deterministic result:

```text
Current Expense:
Rp8.4M

3-Month Average:
Rp6.8M

Difference:
+Rp1.6M
```

Drivers:

```text
Dining:
+Rp700k

Shopping:
+Rp550k

Transport:
+Rp250k

Other:
+Rp100k
```

AI output:

```text
Your spending increased mainly because of dining and shopping.
Together, these categories account for approximately 78% of the
increase compared with your recent baseline.
```

---

# 28. Spending Anomaly Flow

## 28.1 Detection

```text
Transaction Data
   ↓
Deterministic Pattern Engine
   ↓
Compare Current Period
vs Historical Baseline
   ↓
Threshold Triggered?
   ├── No
   │   ↓
   │  No Insight
   │
   └── Yes
       ↓
Create Structured Signal
       ↓
AI Explanation
```

---

## 28.2 Example

```text
Food & Dining

Current:
Rp2.4M

3-Month Avg:
Rp1.5M

Change:
+60%
```

AI insight:

```text
Dining spending is significantly above your recent average.

The largest increase comes from food delivery,
which contributed Rp720,000 above the historical baseline.
```

---

# 29. AI Insight Lifecycle

```text
Signal Detected
   ↓
Generate Insight Context
   ↓
AI Generation Job
   ↓
Validate Structured Output
   ↓
Store Insight
   ↓
NEW
   ↓
User Opens
   ↓
VIEWED
   ├── Acknowledge
   │   ↓
   │  ACKNOWLEDGED
   │
   └── Dismiss
       ↓
      DISMISSED
```

---

# 30. AI Insight Failure Flow

```text
Signal Exists
   ↓
AI Generation Requested
   ↓
Provider Timeout / Error
   ↓
Record AI Failure
   ↓
Do Not Corrupt Financial State
   ↓
Retry if policy allows
```

The deterministic signal remains valid even if explanation generation fails.

---

# 31. Forecast Entry Flow

## 31.1 Goal

Show users what their finances may look like in the future.

### Flow

```text
Forecast
   ↓
Choose Horizon
   ↓
Backend Builds Baseline
   ↓
Collect:
- Current balances
- Known future events
- Recurring income
- Recurring expense
- Estimated variable spending
- Explicit assumptions
   ↓
Sort Events
   ↓
Calculate Projected Balance
   ↓
Calculate Risk Metrics
   ↓
Return Forecast
```

---

# 32. Forecast Event Model

Forecast timeline may contain events such as:

```text
25 Aug
Salary
+Rp12M
SCHEDULED

27 Aug
Rent
-Rp3M
SCHEDULED

28 Aug
Food Estimate
-Rp120k
ESTIMATED
```

The UI should visually distinguish:

```text
KNOWN
SCHEDULED
ESTIMATED
ASSUMED
```

---

# 33. Forecast Result Flow

Example:

```text
Current:
Rp12.4M

Sep 01:
Rp10.1M

Sep 10:
Rp7.6M

Sep 20:
Rp3.2M

Sep 25:
Rp12.0M
```

Summary:

```text
Lowest Projected Balance:
Rp3.2M

Ending Balance:
Rp12M

Forecast Confidence:
MEDIUM
```

---

# 34. Forecast Low-Balance Warning

```text
Forecast Engine
   ↓
Projected Balance Below User/System Threshold?
   ├── No
   │   ↓
   │  Normal
   │
   └── Yes
       ↓
Create CASHFLOW_RISK Signal
       ↓
Notification
       ↓
Optional AI Explanation
```

Example:

```text
Your projected balance reaches Rp850,000 on 18 September,
before your expected salary on 25 September.
```

---

# 35. Forecast Insufficient Data Flow

Example:

```text
Historical Data:
18 days
```

Result:

```text
Forecast Confidence:
LOW
```

UI:

```text
Savio has limited transaction history.

Variable expense projections may be less accurate until
more historical data is available.
```

---

# 36. Scenario Simulator Entry Flow

## 36.1 Goal

Allow users to evaluate hypothetical financial decisions without changing actual financial data.

### Flow

```text
Scenario Simulator
   ↓
Create Scenario
   ↓
Name Scenario
   ↓
Choose Forecast Horizon
   ↓
Add One or More Modifications
   ↓
Calculate
   ↓
Build Baseline
   ↓
Clone Baseline In Memory / Calculation Model
   ↓
Apply Scenario Modifications
   ↓
Calculate Scenario
   ↓
Compare
   ↓
Store Snapshot
```

---

# 37. Scenario: One-Time Purchase

User enters:

```text
Name:
Buy MacBook

Type:
ONE_TIME_EXPENSE

Amount:
Rp15,000,000

Date:
10 Sep 2026
```

Flow:

```text
Baseline Forecast
   ↓
Insert Hypothetical Expense
   ↓
Recalculate
   ↓
Compare Metrics
```

---

# 38. Scenario: Income Reduction

User enters:

```text
Salary decrease:
30%

Starting:
October
```

Flow:

```text
Find Baseline Salary Events
   ↓
Apply -30%
   ↓
Recalculate Future Cashflow
```

Actual recurring salary configuration remains unchanged.

---

# 39. Scenario: Resignation

Example:

```text
What if I resign next month?
```

Modification:

```text
INCOME_REMOVAL

Income Source:
Salary

Effective:
1 Oct 2026
```

System removes projected future salary events from the scenario only.

---

# 40. Scenario: New Installment

Example:

```text
New Motorcycle Installment

Rp1.8M / month
24 months
```

The scenario engine creates hypothetical recurring expense events.

---

# 41. Scenario: Multiple Changes

Example:

```text
Resign
+
Reduce entertainment by 50%
+
Freelance income Rp3M/month
```

The engine applies all changes to the same scenario.

---

# 42. Scenario Comparison Flow

After calculation:

```text
                    BASELINE      SCENARIO

Ending Balance      Rp18.4M       Rp7.2M

Lowest Balance      Rp8.2M        Rp1.1M

Savings Rate        27%           9%

Cash Runway         4.1 months    2.0 months

Goal Completion     Dec 2026      Mar 2027
```

The UI should clearly distinguish:

```text
What changed?
```

and:

```text
What remained unchanged?
```

---

# 43. AI Scenario Explanation Flow

```text
Scenario Calculation
   ↓
Deterministic Comparison
   ↓
AI Context Builder
   ↓
AI receives:
- baseline metrics
- scenario metrics
- difference
- assumptions
   ↓
AI explains impact
```

Example:

```text
The purchase does not immediately create a negative balance,
but it reduces your lowest projected balance from Rp8.2M to
Rp1.1M and delays your emergency fund target by approximately
three months.
```

The AI must not invent these metrics.

---

# 44. Scenario Stale Flow

If real financial data changes after the scenario was calculated:

```text
Scenario Snapshot
calculated_at:
20:00

New Transaction:
20:15
```

System marks:

```text
STALE
```

UI:

```text
Your financial data has changed since this scenario was calculated.

Recalculate to get an updated comparison.
```

---

# 45. "Can I Afford This?" Flow

This may be exposed as a shortcut.

```text
User:
Can I afford a Rp12M laptop this month?
   ↓
AI Intent Detection
   ↓
Recognize Scenario Question
   ↓
Ask Missing Information if Needed
   ↓
Call Scenario Engine
   ↓
Receive Deterministic Comparison
   ↓
AI Explains
```

Example:

```text
AI:
I can simulate that.

Purchase:
Rp12,000,000

When should I assume the purchase happens?

[Today]
[Next Payday]
[Choose Date]
```

After user selects:

```text
Scenario Engine
→ calculate
```

AI presents the result.

---

# 46. AI Copilot Entry Flow

The user opens:

```text
Savio Copilot
```

Potential suggested prompts:

```text
Why did I spend more this month?

What are my biggest recurring expenses?

Are any budgets at risk?

Can I afford a Rp10M purchase?

What happens if I lose my salary?

Am I on track for my emergency fund?
```

---

# 47. AI Copilot Question Flow

```text
User Question
   ↓
Backend Receives Request
   ↓
Authentication
   ↓
Rate Limit
   ↓
Intent Classification
   ↓
Determine Required Finance Tools
   ↓
Call Deterministic Tools
   ↓
Build Structured Context
   ↓
Call LLM
   ↓
Validate Output
   ↓
Return Answer
```

---

# 48. Copilot Tool Flow

Example question:

```text
What are my largest recurring expenses?
```

AI orchestration:

```text
Intent:
RECURRING_EXPENSE_ANALYSIS
   ↓
Tool:
get_recurring_expenses
   ↓
Finance Service Returns:
1. Rent Rp3M
2. Installment Rp1.5M
3. Internet Rp450k
4. Netflix Rp186k
   ↓
LLM
   ↓
Human-Readable Response
```

---

# 49. Copilot: "Why Did I Spend More?"

```text
User Question
   ↓
Tool:
compare_periods
   ↓
Tool:
get_category_breakdown
   ↓
Tool:
get_largest_spending_changes
   ↓
AI Explanation
```

Example answer:

```text
Your spending increased by Rp1.6M compared with your
three-month average.

The main contributors were:

1. Dining: +Rp700k
2. Shopping: +Rp550k
3. Transport: +Rp250k

Dining and shopping account for most of the increase.
```

---

# 50. Copilot Write Proposal Flow

If Savio later supports AI-assisted writes:

User:

```text
Create a food budget of Rp1.5M per month.
```

Flow:

```text
AI Understands Intent
   ↓
AI Generates Proposed Action
   ↓
Backend Validates Proposal
   ↓
Show Confirmation:
Create monthly Food budget Rp1.5M?
   ↓
User Confirms
   ↓
Backend Executes Deterministically
```

AI never directly bypasses confirmation.

---

# 51. AI Copilot Error Flow

Possible conditions:

```text
AI_PROVIDER_UNAVAILABLE
AI_TIMEOUT
AI_OUTPUT_INVALID
RATE_LIMIT_EXCEEDED
INSUFFICIENT_CONTEXT
```

UI should distinguish between:

```text
AI temporarily unavailable
```

and:

```text
Your financial data is insufficient to answer this question.
```

---

# 52. Notification Flow

Example:

```text
Budget reaches warning threshold
   ↓
Budget Engine Creates Signal
   ↓
Notification Job
   ↓
Deduplication Check
   ↓
Create Notification
   ↓
User Sees Notification
```

Notification:

```text
Food budget is 82% used.

Rp360,000 remains for the current period.
```

---

# 53. Upcoming Bill Flow

```text
Background Job
   ↓
Find recurring expenses due soon
   ↓
Notification Preference Check
   ↓
Deduplication
   ↓
Create Notification
```

Example:

```text
Internet bill Rp450,000 is expected in 2 days.
```

---

# 54. Low Forecast Balance Notification

```text
Forecast Recalculated
   ↓
Minimum Balance Below Threshold
   ↓
Create Risk Signal
   ↓
Generate Notification
```

Example:

```text
Your projected balance may fall below Rp1M before your next salary.
```

---

# 55. Search Transaction Flow

```text
Transactions
   ↓
Enter Search:
"Grab"
   ↓
Backend Query
   ↓
Search Across Allowed Fields
   ↓
Return Paginated Result
```

Potential searchable fields:

```text
description
merchant
notes
```

---

# 56. Filter Transaction Flow

User may combine:

```text
Type:
EXPENSE

Account:
BCA

Category:
Food

Date:
1 Aug – 31 Aug

Amount:
Rp50k – Rp500k
```

Flow:

```text
Frontend Builds Query
   ↓
GET Transactions
   ↓
Backend Validates Filters
   ↓
Scope user_id
   ↓
Apply Filter
   ↓
Apply Sort
   ↓
Apply Pagination
   ↓
Return Results
```

---

# 57. Sort Flow

Supported examples:

```text
Newest
Oldest
Highest Amount
Lowest Amount
```

Backend should allow only known sort fields.

Invalid sort input should not become raw SQL input.

---

# 58. Pagination Flow

```text
Page 1
20 rows

Next
   ↓
Page 2
```

Response metadata may include:

```json
{
  "page": 1,
  "limit": 20,
  "total": 142,
  "total_pages": 8
}
```

---

# 59. Reports Flow

```text
Reports
   ↓
Choose:
- Month
- Quarter
- Year
- Custom
   ↓
Backend Aggregation
   ↓
Render Report
```

Possible reports:

```text
Income vs Expense
Category Breakdown
Budget Performance
Goal Progress
Cashflow Trend
Savings Rate
```

---

# 60. AI Insight Feedback Flow

User opens AI insight.

Actions:

```text
Helpful
Not Helpful
Dismiss
```

Flow:

```text
Insight
   ↓
User Feedback
   ↓
Store Feedback
   ↓
Future AI Evaluation / Analytics
```

The system should not automatically retrain a model from one feedback action.

---

# 61. User Settings Flow

```text
Settings
   ↓
Profile
Financial Preferences
AI Preferences
Notifications
Sessions
Security
```

---

# 62. Change Timezone Flow

```text
Settings
   ↓
Timezone
   ↓
Select Asia/Jakarta
   ↓
Save
```

Changing timezone affects interpretation of future schedules.

It must not silently shift historical transaction date meaning incorrectly.

---

# 63. AI Preference Flow

```text
AI Insights:
Enabled / Disabled
```

If disabled:

```text
Automatic AI Insights
→ not generated
```

But:

```text
Finance Analytics
Forecast
Scenario
```

remain available.

Depending on product policy, AI Copilot may have a separate setting.

---

# 64. Session Management Flow

```text
Security
   ↓
Active Sessions
```

Example:

```text
MacBook Chrome
Current Session

iPhone Safari
Last active 2 hours ago
```

Actions:

```text
Revoke Session
Revoke All Other Sessions
```

---

# 65. Revoke Session Flow

```text
User Clicks Revoke
   ↓
Backend Ownership Check
   ↓
Revoke Refresh Session
   ↓
Session Cannot Refresh Again
```

---

# 66. CSRF Flow

Before state-changing request:

```text
Frontend
   ↓
Read CSRF Token
   ↓
Set CSRF Header
   ↓
POST / PATCH / PUT / DELETE
   ↓
Backend Validates:
Cookie Token
Header Token
Session Binding / Signature if implemented
```

Invalid token:

```text
Request Rejected
```

---

# 67. Axios 401 Flow

```text
API Request
   ↓
401
   ↓
Was this already retried?
   ├── Yes
   │   ↓
   │  Logout
   │
   └── No
       ↓
Is refresh already running?
   ├── Yes
   │   ↓
   │  Wait for same refresh promise
   │
   └── No
       ↓
Start Refresh
       ↓
Refresh Success?
   ├── Yes
   │   ↓
   │  Retry queued requests once
   │
   └── No
       ↓
Logout
```

This avoids:

```text
infinite retry
duplicate refresh
refresh race condition
```

---

# 68. Axios 403 Flow

```text
API
→ 403
```

Frontend:

```text
Show Permission / Access Message
```

Do not automatically logout because the user is still authenticated.

---

# 69. Axios 422 Flow

```text
API
→ 422
```

Frontend maps validation errors to:

```text
specific form fields
```

and may also show a general business error.

---

# 70. Axios 429 Flow

```text
API
→ 429
```

Frontend:

```text
Show rate-limit message
Disable repeated rapid submission
Respect Retry-After if provided
```

Especially relevant for AI requests.

---

# 71. Axios 500 Flow

```text
API
→ 500
```

Frontend:

```text
Show generic failure
Offer retry where safe
Do not expose stack trace
```

---

# 72. Optimistic Lock Conflict Flow

Example:

```text
Budget Version 5
```

User opens it.

Another request updates it:

```text
Version 6
```

First user submits old version:

```text
version = 5
```

Backend:

```text
409 VERSION_CONFLICT
```

Frontend:

```text
This resource has changed since you opened it.

Reload latest data.
```

Potential actions:

```text
Reload
Cancel
```

Avoid silently overwriting newer data.

---

# 73. AI Request Rate Limit Flow

```text
User Sends Repeated AI Requests
   ↓
Rate Limit Threshold Reached
   ↓
429 RATE_LIMIT_EXCEEDED
   ↓
UI:
Please wait before sending another AI request.
```

---

# 74. Audit Flow

Important write:

```text
Transaction Updated
```

Backend:

```text
Perform Business Operation
   ↓
Create Audit Event
```

Audit metadata may include:

```text
entity type
entity id
action
request id
timestamp
safe change metadata
```

Never include:

```text
password
refresh token
secret
```

---

# 75. Forecast Freshness Flow

```text
Forecast Generated
   ↓
User Adds New Transaction
   ↓
Forecast Source Data Changed
   ↓
Forecast Marked STALE
```

UI:

```text
Your forecast is based on older financial data.

Recalculate forecast.
```

Automatic recalculation may later improve UX, but the freshness model must remain explicit.

---

# 76. Goal Impact from Scenario

Example:

Baseline:

```text
Laptop Goal Completion:
December
```

Scenario:

```text
Buy Motorcycle:
-Rp15M
```

Scenario engine recalculates goal trajectory.

Result:

```text
Baseline:
December

Scenario:
March
```

Difference:

```text
+3 months
```

AI may explain the trade-off.

---

# 77. Cash Runway Flow

User asks:

```text
How long can I survive if I stop receiving salary?
```

Flow:

```text
User Question
   ↓
Scenario:
Remove Salary
   ↓
Calculate Liquid Balance
   ↓
Calculate Essential Monthly Expense
   ↓
Runway Engine
   ↓
Result
   ↓
AI Explanation
```

Example:

```text
Liquid Balance:
Rp42M

Essential Monthly Expense:
Rp7M

Estimated Runway:
6 months
```

Savio must explain that this is an estimate.

---

# 78. Financial Health Flow

If implemented:

```text
Finance Engine
   ↓
Calculate Components
   ↓
Savings Rate
Cashflow Stability
Budget Adherence
Emergency Buffer
Expense Volatility
   ↓
Weighted Formula
   ↓
Health Score
   ↓
AI Explanation
```

Example:

```text
74 / 100

Positive:
Stable income
Healthy savings rate

Risk:
Low emergency buffer
High discretionary volatility
```

---

# 79. Background Insight Generation Flow

```text
Scheduled Worker
   ↓
Find Users Eligible for Analysis
   ↓
Calculate Deterministic Signals
   ↓
Signal Exists?
   ├── No
   │   ↓
   │  Finish
   │
   └── Yes
       ↓
Deduplication Check
       ↓
Already Generated?
   ├── Yes
   │   ↓
   │  Skip
   │
   └── No
       ↓
Queue AI Generation
       ↓
Call AI
       ↓
Validate
       ↓
Store Insight
```

---

# 80. AI Structured Output Validation Flow

```text
LLM Response
   ↓
Parse JSON
   ↓
Schema Validation
   ↓
Valid?
   ├── Yes
   │   ↓
   │  Continue
   │
   └── No
       ↓
Retry / Fail Safely
```

Invalid AI output must never become trusted application state.

---

# 81. Future CSV Import Flow

Post-MVP:

```text
Transactions
   ↓
Import
   ↓
Upload CSV
   ↓
Store Temporary File
   ↓
Background Parse
   ↓
Validate Rows
   ↓
Show Import Review
```

Example:

```text
1,200 rows

Valid:
1,170

Warnings:
20

Invalid:
10
```

User:

```text
Review
   ↓
Resolve / Skip Invalid Rows
   ↓
Confirm Import
   ↓
Create Transactions Atomically / Batched
```

---

# 82. Future Receipt Flow

```text
Upload Receipt
   ↓
MinIO
   ↓
AI / Extraction Service
   ↓
Structured Draft:
- Merchant
- Amount
- Date
- Category Suggestion
   ↓
User Review
   ↓
Confirm
   ↓
Create Transaction
```

Extraction never directly creates financial state without review.

---

# 83. Mobile Responsive Flow

Savio should support:

```text
Desktop
Tablet
Mobile
```

On mobile:

```text
Sidebar
→ collapses to drawer / bottom navigation where appropriate

Tables
→ card / horizontally scrollable layout

Charts
→ responsive

Forms
→ stacked
```

Critical actions remain reachable.

---

# 84. Loading State Flow

Every async page must show an appropriate loading state.

Examples:

```text
Skeleton
Spinner
Loading card
```

Avoid blank pages during data fetching.

---

# 85. Empty State Flow

Examples:

## No Transactions

```text
No transactions yet.

Start by adding your first income or expense.

[Add Transaction]
```

## No Budget

```text
You have not created any budgets.

Create one to monitor your monthly spending.

[Create Budget]
```

## No Goals

```text
No financial goals yet.

[Create Goal]
```

---

# 86. Error State Flow

Page-level error:

```text
Failed to load transactions.

[Try Again]
```

Never show raw backend stack traces.

---

# 87. Form Submission Flow

General form behavior:

```text
Idle
   ↓
User Edits
   ↓
Frontend Validation
   ↓
Submit
   ↓
Disable Submit
   ↓
Loading
   ↓
Success / Validation Error / Server Error
```

Double-clicking submit should not accidentally create duplicate financial records.

---

# 88. Confirmation Flow

Potentially destructive actions require confirmation.

Examples:

```text
Void Transaction
Void Transfer
Archive Account
End Recurring Rule
Revoke Session
```

Confirmation should explain consequence.

Example:

```text
Void this transaction?

The original financial effect will be reversed
and your account balance will be recalculated.
```

---

# 89. Toast Feedback

Suitable actions:

```text
Transaction created.
Budget updated.
Scenario recalculated.
Session revoked.
```

Toasts must not replace important inline errors.

---

# 90. Permission and Ownership Failure Flow

Even if the user manually changes a URL:

```text
/transactions/{other-user-id}
```

backend ownership check:

```text
Authenticated user
!= resource owner
```

Result:

```text
404
```

or:

```text
403
```

depending on the chosen security policy.

A `404` may be preferable where hiding resource existence improves privacy.

The policy must be consistent.

---

# 91. Critical Demo Journey

The final Savio demo should ideally follow one coherent story.

Example user:

```text
Name:
Alex

Bank Balance:
Rp20M

Monthly Salary:
Rp12M

Average Expense:
Rp8M
```

Demo flow:

```text
1. Login

2. Dashboard
   Show:
   - current balance
   - income
   - expense
   - net cashflow

3. Add Expense
   GrabFood Rp250k

4. AI suggests:
   Food & Dining

5. Create Monthly Food Budget
   Rp2M

6. Show current budget utilization

7. Show AI Insight:
   Dining spending is above baseline

8. Open Forecast
   Show projected future balance

9. Create Goal:
   Emergency Fund Rp30M

10. Open Scenario Simulator

11. Scenario:
    Buy Laptop Rp15M

12. Compare:
    baseline vs scenario

13. Show:
    - ending balance
    - minimum balance
    - savings rate
    - emergency buffer
    - goal delay

14. Ask AI Copilot:
    "What is the biggest impact if I buy this laptop?"

15. AI explains deterministic scenario result

16. User remains in control
```

This demonstrates Savio's full thesis:

```text
TRACK
   ↓
UNDERSTAND
   ↓
PREDICT
   ↓
SIMULATE
   ↓
EXPLAIN
   ↓
DECIDE
```

---

# 92. Error Demo Journey

A technical review may also demonstrate:

```text
Attempt invalid expense category
→ 422

Attempt transfer to same account
→ 422

Submit stale budget version
→ 409

Expired authentication
→ single refresh
→ request retry

Invalid refresh
→ logout

AI provider unavailable
→ deterministic finance functionality still works
```

This demonstrates that Savio is not only happy-path software.

---

# 93. User Flow to Backend Mapping

High-level mapping:

```text
Authentication
→ Auth Module

Accounts
→ Account Module

Transactions
→ Transaction Module

Transfers
→ Transfer Module

Recurring
→ Recurring Module

Budgets
→ Budget Module

Goals
→ Goal Module

Analytics
→ Analytics Module

Forecast
→ Forecast Module

Scenarios
→ Scenario Module

AI Insights
→ AI Insight Module

AI Copilot
→ AI Orchestration Module

Notifications
→ Notification Module

Audit
→ Audit Module
```

---

# 94. User Flow to Frontend Mapping

Potential frontend feature structure:

```text
features/
├── auth/
├── onboarding/
├── dashboard/
├── accounts/
├── transactions/
├── recurring/
├── budgets/
├── goals/
├── analytics/
├── forecast/
├── scenarios/
├── insights/
├── copilot/
├── notifications/
└── settings/
```

Exact route design will be defined in the frontend architecture document.

---

# 95. Core Workflow Dependencies

The primary dependency chain is:

```text
User
   ↓
Account
   ↓
Transaction
   ↓
Analytics
   ↓
Budget / Goal
   ↓
Forecast
   ↓
Scenario
   ↓
AI Insight / Copilot
```

This means:

- forecast depends on reliable financial data,
- scenario depends on forecast,
- AI explanations depend on deterministic output.

AI should therefore never become the lowest-level dependency.

---

# 96. MVP User Flow Priority

## P0

Must work end-to-end:

```text
Registration
Login
Logout
Session Refresh

Onboarding

Create Account
Archive Account

Create Income
Create Expense
Edit Transaction
Void Transaction

Transfer

Recurring Income
Recurring Expense

Budget
Budget Utilization

Financial Goal

Dashboard
Analytics

Cashflow Forecast

Scenario Simulator

AI Insight

AI Copilot

Search
Filter
Sort
Pagination

Validation
Errors
Responsive UI
```

---

## P1

High-value additions:

```text
AI Transaction Categorization

Budget Overspend Prediction

Spending Anomaly Detection

Goal Feasibility

Financial Health

Notifications

AI Insight Feedback

Session Management

Background AI Jobs
```

---

## P2

Enhancements:

```text
CSV Import
Receipt Upload
Receipt Extraction
Household Finance
Multi-Currency
Advanced Reports
Bank Integration
Advanced Recurring Detection
```

---

# 97. Flow Completion Rule

A Savio flow is considered implemented only when:

```text
happy path works
+
validation works
+
authorization works
+
financial integrity is preserved
+
error state exists
+
loading state exists
+
empty state exists where applicable
+
business rules are tested
```

For AI-enabled flows:

```text
deterministic source exists
+
AI context is scoped
+
AI output is validated
+
failure mode exists
+
user remains in control
```

---

# 98. Product Flow Principle

Savio should always prioritize the following experience:

```text
DATA
  ↓
CONTEXT
  ↓
UNDERSTANDING
  ↓
FORECAST
  ↓
SIMULATION
  ↓
EXPLANATION
  ↓
DECISION
```

The application should never degrade into:

```text
DATA
  ↓
CHART
  ↓
END
```

or:

```text
DATA
  ↓
LLM
  ↓
UNVERIFIED ANSWER
```

---

# 99. Final User Flow Model

The complete conceptual flow of Savio is:

```text
USER
 │
 ▼
FINANCIAL INPUT
 │
 ├── Accounts
 ├── Transactions
 ├── Recurring Events
 ├── Budgets
 └── Goals
 │
 ▼
FINANCE ENGINE
 │
 ├── Balance
 ├── Cashflow
 ├── Analytics
 ├── Budget Calculation
 ├── Goal Calculation
 ├── Forecast
 └── Scenario Simulation
 │
 ▼
FINANCIAL CONTEXT
 │
 ├── Current State
 ├── Historical Pattern
 ├── Future Projection
 └── Scenario Difference
 │
 ▼
AI LAYER
 │
 ├── Explanation
 ├── Pattern Interpretation
 ├── Categorization Suggestion
 ├── Insight Generation
 └── Copilot
 │
 ▼
USER
 │
 ▼
DECISION
```

The authoritative order must always remain:

> **User data → deterministic finance engine → financial result → AI interpretation → user decision.**

---

# 100. Final User Experience Statement

Savio should allow a user to move naturally from:

> **"I know how much money I have."**

to:

> **"I understand where my money is going."**

then:

> **"I can see what my cashflow may look like next."**

then:

> **"I can test a decision before I make it."**

and finally:

> **"I understand the trade-offs well enough to decide for myself."**

That progression defines the intended Savio user experience.

> **Finance Engine calculates. AI interprets. User decides.**